package dao

import (
	"letter-manage-backend/model"
)

// CreateNotification inserts a notification
func CreateNotification(n *model.Notification) error {
	return DB.Create(n).Error
}

// GetUnreadCount returns count of unread notifications for a user
func GetUnreadCount(userID uint) (int64, error) {
	var count int64
	err := DB.Model(&model.Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Count(&count).Error
	return count, err
}

// GetNotifications returns recent notifications for a user
func GetNotifications(userID uint, limit int) ([]model.Notification, error) {
	var list []model.Notification
	if limit <= 0 {
		limit = 20
	}
	err := DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&list).Error
	return list, err
}

// MarkAsRead marks a notification as read
func MarkAsRead(id uint, userID uint) error {
	return DB.Model(&model.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_read", true).Error
}

// MarkAllAsRead marks all notifications for a user as read
func MarkAllAsRead(userID uint) error {
	return DB.Model(&model.Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Update("is_read", true).Error
}

// NotifyUser creates a notification for a specific user
func NotifyUser(userID uint, notifType, title, message, letterNo string) error {
	return CreateNotification(&model.Notification{
		UserID:   userID,
		Type:     notifType,
		Title:    title,
		Message:  message,
		LetterNo: letterNo,
	})
}

// NotifyUnitUsers creates notifications for all users in a unit and its subordinate units
func NotifyUnitUsers(unitID uint, notifType, title, message, letterNo string) {
	unitIDs := GetSubordinateUnitIDs(unitID)
	if len(unitIDs) == 0 {
		unitIDs = []uint{unitID}
	}
	var users []model.PoliceUser
	DB.Where("unit_id IN ? AND is_active = true", unitIDs).Find(&users)
	for _, u := range users {
		CreateNotification(&model.Notification{
			UserID:   u.ID,
			Type:     notifType,
			Title:    title,
			Message:  message,
			LetterNo: letterNo,
		})
	}
}
