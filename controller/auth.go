package controller

import (
	"fmt"
	"net/http"
	"time"

	"letter-manage-backend/middleware"
	"letter-manage-backend/model"
	"letter-manage-backend/service"

	"github.com/gin-gonic/gin"
)

// AuthController handles /api/auth/
func AuthController(c *gin.Context) {
	var req model.APIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResp("invalid request"))
		return
	}

	switch req.Order {
	case "login":
		handleLogin(c, req.Args)
	case "logout":
		handleLogout(c)
	case "check":
		handleCheck(c)
	default:
		c.JSON(http.StatusBadRequest, model.ErrorResp("unknown order: "+req.Order))
	}
}

func handleLogin(c *gin.Context, args map[string]interface{}) {
	policeNumber, _ := args["police_number"].(string)
	password, _ := args["password"].(string)
	rememberMe, _ := args["remember_me"].(bool)

	if policeNumber == "" || password == "" {
		c.JSON(http.StatusOK, model.ErrorResp("police_number and password required"))
		return
	}

	ip := c.ClientIP()
	userAgent := c.Request.UserAgent()
	result, err := service.Login(policeNumber, password, rememberMe, ip, userAgent)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}

	maxAge := 8 * 3600
	if rememberMe {
		maxAge = 30 * 24 * 3600
	}
	c.Header("Set-Cookie", fmt.Sprintf("session_key=%s; Path=/; Max-Age=%d; HttpOnly", result.SessionKey, maxAge))

	// Return user info (without password)
	userInfo := map[string]interface{}{
		"id":               result.User.ID,
		"name":             result.User.Name,
		"nickname":         result.User.Nickname,
		"police_number":    result.User.PoliceNumber,
		"phone":            result.User.Phone,
		"unit_id":          result.User.UnitID,
		"permission_level": result.User.PermissionLevel,
		"is_active":        result.User.IsActive,
		"last_login":       result.User.LastLogin,
	}
	c.JSON(http.StatusOK, model.SuccessResp(map[string]interface{}{
		"user": userInfo,
	}))
}

func handleLogout(c *gin.Context) {
	sessionKey, err := c.Cookie("session_key")
	if err == nil && sessionKey != "" {
		service.Logout(sessionKey)
	}
	c.SetCookie("session_key", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, model.SuccessResp(nil))
}

func handleCheck(c *gin.Context) {
	sessionKey, err := c.Cookie("session_key")
	if err != nil || sessionKey == "" {
		c.JSON(http.StatusOK, model.SuccessResp(map[string]interface{}{
			"logged_in": false,
		}))
		return
	}
	user, err := service.CheckSession(sessionKey)
	if err != nil {
		c.JSON(http.StatusOK, model.SuccessResp(map[string]interface{}{
			"logged_in": false,
		}))
		return
	}
	now := time.Now()
	userInfo := map[string]interface{}{
		"id":               user.ID,
		"name":             user.Name,
		"nickname":         user.Nickname,
		"police_number":    user.PoliceNumber,
		"phone":            user.Phone,
		"unit_id":          user.UnitID,
		"permission_level": user.PermissionLevel,
		"is_active":        user.IsActive,
		"server_time":      now.Format("2006-01-02 15:04:05"),
	}
	c.JSON(http.StatusOK, model.SuccessResp(map[string]interface{}{
		"logged_in": true,
		"user":      userInfo,
	}))
}

// AuthMiddlewareRequired is a helper to get current user or fail
func getCurrentUserOrFail(c *gin.Context) *model.PoliceUser {
	return middleware.GetCurrentUser(c)
}
