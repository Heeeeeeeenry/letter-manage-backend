package controller

import (
	"net/http"
	"strconv"

	"letter-manage-backend/dao"
	"letter-manage-backend/middleware"
	"letter-manage-backend/model"

	"github.com/gin-gonic/gin"
)

// GetUnreadCount returns unread notification count
func GetUnreadCount(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	count, err := dao.GetUnreadCount(user.ID)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(map[string]int64{"count": count}))
}

// GetNotifications returns recent notifications
func GetNotifications(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	list, err := dao.GetNotifications(user.ID, limit)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	if list == nil {
		list = []model.Notification{}
	}
	c.JSON(http.StatusOK, model.SuccessResp(list))
}

// MarkNotificationRead marks a single notification as read
func MarkNotificationRead(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResp("invalid id"))
		return
	}
	if err := dao.MarkAsRead(uint(id), user.ID); err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(nil))
}

// MarkAllRead marks all notifications as read
func MarkAllRead(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if err := dao.MarkAllAsRead(user.ID); err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(nil))
}
