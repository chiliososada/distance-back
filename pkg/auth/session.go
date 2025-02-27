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
	CsrfToken   string   `mapstructure:"csrf_token" json:"csrf_token"`
	ChatToken   string   `mapstructure:"chat_token" json:"chat_token"`
	UID         string   `mapstructure:"uid" json:"uid"`
	DisplayName string   `mapstructure:"display_name" json:"display_name"`
	PhotoUrl    string   `mapstructure:"photo_url" json:"photo_url"`
	Email       string   `mapstructure:"email" json:"email"`
	Gender      string   `mapstructure:"gender" json:"gender"`
	Bio         string   `mapstructure:"bio" json:"bio"`
	Session     string   `mapstructure:"session" json:"-"`
	ChatID      []string `mapstructure:"-" json:"chat_id"`
	ChatUrl     string   `mapstructure:"chat_url" json:"chat_url"`
}

func getChatUrl(_ string) string {
	return "https://192.168.0.143:55372/ws/distance"
}

type SessionRecentTopicCursor struct {
	cursor time.Time
}

func sessionDataKey(uid string) string {
	return fmt.Sprintf("user-session-data:%s", uid)
}

func userSessionKey(session string) string {
	return fmt.Sprintf("user-session:%s", session)
}

func userChatKey(uid string) string {
	return fmt.Sprintf("user-chat:%s", uid)
}

func (data *SessionData) IsValid() bool {
	return data != nil && data.CsrfToken != ""
}

func setSessionData(c *gin.Context, uid string, value *SessionData) error {

	var data map[string]interface{}
	key := sessionDataKey(uid)
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
			chatKey := userChatKey(uid)
			if err := cache.RedisClient.SAdd(ctx, chatKey, value.ChatID).Err(); err != nil {
				cache.RedisClient.Del(ctx, key)
				return err
			} else {
				if err := cache.RedisClient.Expire(ctx, chatKey, SessionDuration).Err(); err != nil {
					cache.RedisClient.Del(ctx, key)
					cache.RedisClient.Del(ctx, chatKey)
					return err
				}
			}

			return nil
		}

	}

}

func setUserSession(c *gin.Context, uid string, session string) error {
	return cache.RedisClient.Set(c.Request.Context(), userSessionKey(session), uid, SessionDuration).Err()

}

func UpdateSessionData(c *gin.Context, uid string, newSession *SessionData) error {
	ctx := c.Request.Context()
	if _, err := cache.RedisClient.Get(ctx, userSessionKey(uid)).Result(); err != nil {
		return err
	} else {
		return setSessionData(c, uid, newSession)
	}
}

func CreateUserSession(c *gin.Context, uid string, session string, csrfToken string, chatToken string,
	user *firebase_auth.UserRecord, userRecord *model.User) (*SessionData, error) {

	gender := "unknown"
	bio := ""
	if userRecord != nil {
		gender = userRecord.Gender
		bio = userRecord.Bio
	}

	chatID := []string{}
	if userRecord != nil {
		for _, chat := range userRecord.Chat {
			chatID = append(chatID, chat.ChatRoomUID)
		}
	}

	sd := SessionData{
		CsrfToken:   csrfToken,
		ChatToken:   chatToken,
		UID:         uid,
		DisplayName: user.DisplayName,
		PhotoUrl:    user.PhotoURL,
		Email:       user.Email,
		Gender:      gender,
		Bio:         bio,
		Session:     session,
		ChatID:      chatID,
		ChatUrl:     getChatUrl(uid),
	}

	err := setSessionData(c, uid, &sd)
	if err != nil {
		return nil, err
	} else {
		err = setUserSession(c, uid, session)
		if err != nil {
			removeSessionData(uid)
			return nil, err

		} else {

			return &sd, nil
		}

	}

}

func removeSessionData(uid string) error {
	return cache.Delete(sessionDataKey(uid))
}

func RemoveUserSession(c *gin.Context, uid string) error {

	ctx := c.Request.Context()
	session, err := cache.RedisClient.HGet(ctx, sessionDataKey(uid), "session").Result()
	if err != nil {
		if err == redis.Nil {
			return nil
		} else {
			return err
		}
	}

	_, err = cache.RedisClient.Del(ctx, sessionDataKey(uid)).Result()
	if err != nil {
	}
	_, err = cache.RedisClient.Del(ctx, userSessionKey(session)).Result()
	if err != nil {
	}

	_, err = cache.RedisClient.Del(ctx, userChatKey(uid)).Result()
	if err != nil {
	}
	return nil

}

func SetSessionDataInContext(c *gin.Context, uid string, session string) error {
	var sd SessionData
	key := sessionDataKey(uid)
	count, err := cache.RedisClient.Exists(c.Request.Context(), key).Result()
	if err != nil {
		return nil
	} else if count == 0 {
		return errors.New("no session")
	}
	result := cache.RedisClient.HGetAll(c.Request.Context(), sessionDataKey(uid))
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
			//read chat id
			chatKey := userChatKey(uid)
			chatID, err := cache.RedisClient.SMembers(c.Request.Context(), chatKey).Result()
			if err != nil && err != redis.Nil {
				return err
			} else if err != redis.Nil {
				sd.ChatID = chatID
			} else {
			}
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
