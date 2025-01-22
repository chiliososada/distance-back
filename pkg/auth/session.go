package auth

import (
	"fmt"
	"time"

	"github.com/chiliososada/distance-back/internal/api/response"
	"github.com/chiliososada/distance-back/pkg/cache"
	"github.com/chiliososada/distance-back/pkg/logger"
	"github.com/gin-gonic/gin"
)

type SessionData struct {
	response.LoginInfo
	Cookie string
}

const SessionDuration = time.Hour * 24 * 5

func sessionKey(uid string, session string) string {
	return fmt.Sprintf("%s:%s", uid, session)
}

func (data *SessionData) IsValid() bool {
	return data != nil && data.CsrfToken != ""
}

func SetSessionData(uid string, session string, value *SessionData) error {
	fmt.Printf("sess: %v", *value)
	return cache.Set(sessionKey(uid, session), value, SessionDuration)
}

func RemoveSessionData(uid string, session string) error {
	return cache.Delete(sessionKey(uid, session))
}

func RemoveUserSession(uid string) error {
	prefix := fmt.Sprintf("%s*", uid)
	fmt.Printf("prefix: %v\n", prefix)
	keys, err := cache.Scan(prefix)
	if err != nil {
		return err
	}

	for _, k := range keys {
		fmt.Printf("user key: %v\n", k)
		err := cache.Delete(k)
		if err != nil {
			logger.Error("remove user session", logger.Err(err))
		}
	}
	return nil

}

func GetSessionData(uid string, session string) (*SessionData, error) {
	var data SessionData
	if err := cache.Get(sessionKey(uid, session), &data); err != nil {
		return nil, err
	} else if data.IsValid() {
		return &data, nil
	} else {
		return nil, nil
	}

}

func SetSessionInContext(c *gin.Context, data *SessionData) {
	c.Set("session", data)
}

func GetSessionFromContext(c *gin.Context) (data *SessionData, exist bool) {
	r, e := c.Get("session")
	if !e {
		return
	}

	t, ok := r.(*SessionData)
	if !ok {
		return
	} else {
		data = t
		exist = true
		return
	}
}
