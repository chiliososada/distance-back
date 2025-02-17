package auth

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"time"

	firebase_auth "firebase.google.com/go/v4/auth"
	"github.com/chiliososada/distance-back/internal/model"
	"github.com/chiliososada/distance-back/pkg/cache"
	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
	"github.com/redis/go-redis/v9"
)

const (
	SessionDuration = time.Hour * 24 * 5
)

var sessionDataDecoderConfig = &mapstructure.DecoderConfig{
	DecodeHook: mapstructure.ComposeDecodeHookFunc(sessionDataDecodeHook),
}

type RedisRecentTopicCursor struct {
	c time.Time
}

func (c RedisRecentTopicCursor) ToString() string {
	return strconv.FormatInt(c.c.Unix(), 10)
}

func (c RedisRecentTopicCursor) MarshalBinary() ([]byte, error) {
	return []byte(c.ToString()), nil
}

func (c *RedisRecentTopicCursor) UnmarshalBinary(data []byte) error {
	i, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return err
	}
	c.c = time.Unix(i, 0)
	return nil

}
func NewRedisRecentTopicCursor(score float64) *RedisRecentTopicCursor {
	return &RedisRecentTopicCursor{
		c: time.Unix(int64(score), 0),
	}
}

func sessionDataDecodeHook(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {

	if from.Kind() == reflect.String && to.Kind() == reflect.Ptr && to.Elem().Kind() == reflect.Struct {
		to_name := to.Elem().Name()
		if to_name == "RedisRecentTopicCursor" {
			// string => *RedisRecentTopicCursor
			score, err := strconv.ParseInt(data.(string), 10, 64)
			if err != nil {
				return nil, err
			} else {
				return NewRedisRecentTopicCursor(float64(score)), nil

			}
		}
	}

	if to.Kind() == reflect.String && from.Kind() == reflect.Ptr && from.Elem().Kind() == reflect.Struct {
		from_name := from.Elem().Name()
		if from_name == "RedisRecentTopicCursor" {
			// *RedisRecentTopicCursor => string
			val := reflect.ValueOf(data)
			to_string := val.MethodByName("ToString")
			if !to_string.IsValid() {
				return nil, fmt.Errorf("method %s not found", "ToString")
			} else {
				str := to_string.Call(nil)
				//fmt.Printf("to string result: %+v\n", str)
				if len(str) > 0 {
					return str[0].Interface(), nil
				}
			}
		}
	}
	return data, nil
}

type SessionData struct {
	CsrfToken   string `mapstructure:"csrf_token" json:"csrf_token"`
	UID         string `mapstructure:"uid" json:"uid"`
	DisplayName string `mapstructure:"display_name" json:"display_name"`
	PhotoUrl    string `mapstructure:"photo_url" json:"photo_url"`
	Email       string `mapstructure:"email" json:"email"`
	Gender      string `mapstructure:"gender" json:"gender"`
	Bio         string `mapstructure:"bio" json:"bio"`
	//RecentTopicCursor *RedisRecentTopicCursor `mapstructure:"recent_topic_cursor,omitempty" json:"_"`
}

type SessionRecentTopicCursor struct {
	cursor time.Time
}

func sessionDataKey(uid string, session string) string {
	return fmt.Sprintf("user-session-data:%s:%s", uid, session)
}

func userSessionKey(uid string) string {
	return fmt.Sprintf("user-session:%s", uid)
}

func (data *SessionData) IsValid() bool {
	return data != nil && data.CsrfToken != ""
}

func setSessionData(c *gin.Context, uid string, session string, value *SessionData) error {

	var data map[string]interface{}
	key := sessionDataKey(uid, session)
	ctx := c.Request.Context()
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(sessionDataDecodeHook),
		Result:     &data,
	})
	if err != nil {
		return err
	}
	if err := decoder.Decode(*value); err != nil {
		return err
	}
	result := cache.RedisClient.HSet(ctx, key, data)
	if result.Err() != nil {
		return result.Err()
	} else {

		if err := cache.RedisClient.Expire(ctx, key, SessionDuration).Err(); err != nil {
			cache.RedisClient.Del(ctx, key)
			return err
		} else {
			return nil
		}

	}

}

func setUserSession(c *gin.Context, uid string, session string) error {
	return cache.RedisClient.Set(c.Request.Context(), userSessionKey(uid), session, SessionDuration).Err()

}

func UpdateSessionData(c *gin.Context, uid string, newSession *SessionData) error {
	ctx := c.Request.Context()
	if session, err := cache.RedisClient.Get(ctx, userSessionKey(uid)).Result(); err != nil {
		return err
	} else {
		return setSessionData(c, uid, session, newSession)
	}
}

func CreateUserSession(c *gin.Context, uid string, session string, csrfToken string,
	user *firebase_auth.UserRecord, userRecord *model.User) (*SessionData, error) {

	gender := "unknown"
	bio := ""
	if userRecord != nil {
		gender = userRecord.Gender
		bio = userRecord.Bio
	}

	sd := SessionData{
		CsrfToken:   csrfToken,
		UID:         uid,
		DisplayName: user.DisplayName,
		PhotoUrl:    user.PhotoURL,
		Email:       user.Email,
		Gender:      gender,
		Bio:         bio,
	}

	err := setSessionData(c, uid, session, &sd)
	if err != nil {
		return nil, err
	} else {
		err = setUserSession(c, uid, session)
		if err != nil {
			removeSessionData(uid, session)
			return nil, err

		} else {

			return &sd, nil
		}

	}

}

func removeSessionData(uid string, session string) error {
	return cache.Delete(sessionDataKey(uid, session))
}

func RemoveUserSession(c *gin.Context, uid string) error {

	ctx := c.Request.Context()
	session, err := cache.RedisClient.Get(ctx, userSessionKey(uid)).Result()
	if err == redis.Nil {
		return nil
	}
	if err != nil {
		return err
	} else {
		cache.RedisClient.Del(ctx, userSessionKey(uid))
		cache.RedisClient.Del(ctx, sessionDataKey(uid, session))
		return nil
	}

}

func SetSessionDataInContext(c *gin.Context, uid string, session string) error {
	var sd SessionData
	key := sessionDataKey(uid, session)
	count, err := cache.RedisClient.Exists(c.Request.Context(), key).Result()
	if err != nil {
		return nil
	} else if count == 0 {
		return errors.New("no session")
	}
	result := cache.RedisClient.HGetAll(c.Request.Context(), sessionDataKey(uid, session))
	if result.Err() != nil {
		return result.Err()
	} else {

		decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			DecodeHook: mapstructure.ComposeDecodeHookFunc(sessionDataDecodeHook),
			Result:     &sd,
		})
		if err != nil {
			return err
		}
		if err := decoder.Decode(result.Val()); err != nil {
			return err
		} else {
			//store session data in context
			c.Set("session", sd)
			return nil
		}
	}
}

func GetSessionFromContext(c *gin.Context) *SessionData {
	r, e := c.Get("session")
	if !e {
		c.AbortWithStatus(http.StatusUnauthorized)
		return nil
	}

	t, ok := r.(SessionData)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return nil
	} else {
		return &t
	}
}
