package model

import "time"

// Notification represents the notifications table
type Notification struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    uint      `json:"user_id" gorm:"column:user_id;index;not null"`
	Type      string    `json:"type" gorm:"column:type;size:32;not null"`       // dispatch/process/audit/handle_by_self
	Title     string    `json:"title" gorm:"column:title;size:128;not null"`
	Message   string    `json:"message" gorm:"column:message;size:256"`
	LetterNo  string    `json:"letter_no" gorm:"column:letter_no;size:64;index"`
	IsRead    bool      `json:"is_read" gorm:"column:is_read;default:false;index"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
}

func (Notification) TableName() string { return "notifications" }
