package model

import "time"

// PermissionLevel defines user permission levels
type PermissionLevel string

const (
	PermissionCity     PermissionLevel = "CITY"
	PermissionDistrict PermissionLevel = "DISTRICT"
	PermissionOfficer  PermissionLevel = "OFFICER"
)

// PoliceUser represents the police_users table
type PoliceUser struct {
	ID              uint            `json:"id" gorm:"primaryKey;autoIncrement"`
	PasswordHash    string          `json:"-" gorm:"column:password_hash;size:128;not null"`
	Name            string          `json:"name" gorm:"column:name;size:64;not null"`
	Nickname        string          `json:"nickname" gorm:"column:nickname;size:64"`
	PoliceNumber    string          `json:"police_number" gorm:"column:police_number;uniqueIndex;size:32;not null"`
	Phone           string          `json:"phone" gorm:"column:phone;size:32"`
	UnitName        string          `json:"unit_name" gorm:"column:unit_name;size:128"`
	PermissionLevel PermissionLevel `json:"permission_level" gorm:"column:permission_level;type:enum('CITY','DISTRICT','OFFICER');not null"`
	IsActive        bool            `json:"is_active" gorm:"column:is_active;default:true"`
	CreatedAt       time.Time       `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	LastLogin       *time.Time      `json:"last_login" gorm:"column:last_login"`
}

func (PoliceUser) TableName() string { return "police_users" }

// UserSession represents the user_sessions table
type UserSession struct {
	ID         uint       `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID     uint       `json:"user_id" gorm:"column:user_id;not null;index"`
	SessionKey string     `json:"session_key" gorm:"column:session_key;uniqueIndex;size:64;not null"`
	IPAddress  string     `json:"ip_address" gorm:"column:ip_address;size:64"`
	UserAgent  string     `json:"user_agent" gorm:"column:user_agent;size:256"`
	CreatedAt  time.Time  `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	ExpiresAt  time.Time  `json:"expires_at" gorm:"column:expires_at;not null"`
	User       PoliceUser `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (UserSession) TableName() string { return "user_sessions" }

// Unit represents the units table
type Unit struct {
	ID         uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Level1     string    `json:"level1" gorm:"column:level1;size:128"`
	Level2     string    `json:"level2" gorm:"column:level2;size:128"`
	Level3     string    `json:"level3" gorm:"column:level3;size:128"`
	SystemCode string    `json:"system_code" gorm:"column:system_code;uniqueIndex;size:64;not null"`
	CreatedAt  time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
}

func (Unit) TableName() string { return "units" }
