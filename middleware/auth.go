package middleware

import (
	"net/http"

	"letter-manage-backend/dao"
	"letter-manage-backend/model"

	"github.com/gin-gonic/gin"
)

const UserKey = "current_user"
const SessionKey = "session_key"

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionKey, err := c.Cookie("session_key")
		if err != nil || sessionKey == "" {
			c.JSON(http.StatusUnauthorized, model.ErrorResp("未登录"))
			c.Abort()
			return
		}
		session, err := dao.GetSessionByKey(sessionKey)
		if err != nil || session == nil {
			c.JSON(http.StatusUnauthorized, model.ErrorResp("会话已过期，请重新登录"))
			c.Abort()
			return
		}
		c.Set(UserKey, &session.User)
		c.Set(SessionKey, sessionKey)
		c.Next()
	}
}

func GetCurrentUser(c *gin.Context) *model.PoliceUser {
	if v, exists := c.Get(UserKey); exists {
		if user, ok := v.(*model.PoliceUser); ok {
			return user
		}
	}
	return nil
}

func GetSessionKeyFromCtx(c *gin.Context) string {
	if v, exists := c.Get(SessionKey); exists {
		if key, ok := v.(string); ok {
			return key
		}
	}
	return ""
}
