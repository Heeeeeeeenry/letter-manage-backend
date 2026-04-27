package model

import "time"

// Letter represents the letters table
type Letter struct {
	ID            uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	LetterNo      string    `json:"letter_no" gorm:"column:letter_no;uniqueIndex;size:64;not null"`
	CitizenName   string    `json:"citizen_name" gorm:"column:citizen_name;size:64"`
	Phone         string    `json:"phone" gorm:"column:phone;size:32"`
	IDCard        string    `json:"id_card" gorm:"column:id_card;size:32"`
	ReceivedAt    time.Time `json:"received_at" gorm:"column:received_at"`
	Channel       string    `json:"channel" gorm:"column:channel;size:64"`
	CategoryL1    string    `json:"category_l1" gorm:"column:category_l1;size:64"`
	CategoryL2    string    `json:"category_l2" gorm:"column:category_l2;size:64"`
	CategoryL3    string    `json:"category_l3" gorm:"column:category_l3;size:64"`
	Content       string    `json:"content" gorm:"column:content;type:text"`
	SpecialTags   JSONRaw   `json:"special_tags" gorm:"column:special_tags;type:json"`
	CurrentUnit     string    `json:"current_unit" gorm:"column:current_unit;size:128"`
	CurrentStatus   string    `json:"current_status" gorm:"column:current_status;size:64"`
	CurrentOperator string    `json:"current_operator" gorm:"column:current_operator;size:64"`
	DeadlineAt    *time.Time `json:"deadline_at" gorm:"column:deadline_at"`
	CreatedAt     *time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     *time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

func (Letter) TableName() string { return "letters" }

// LetterFlow represents the letter_flows table
type LetterFlow struct {
	ID          uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	LetterNo    string    `json:"letter_no" gorm:"column:letter_no;index;size:64;not null"`
	FlowRecords JSONRaw   `json:"flow_records" gorm:"column:flow_records;type:json"`
	CreatedAt   *time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   *time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

func (LetterFlow) TableName() string { return "letter_flows" }

// LetterAttachment represents the letter_attachments table
type LetterAttachment struct {
	ID                    uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	LetterNo              string    `json:"letter_no" gorm:"column:letter_no;uniqueIndex;size:64;not null"`
	CityDispatchFiles     JSONRaw   `json:"city_dispatch_files" gorm:"column:city_dispatch_files;type:json"`
	DistrictDispatchFiles JSONRaw   `json:"district_dispatch_files" gorm:"column:district_dispatch_files;type:json"`
	HandlerFeedbackFiles  JSONRaw   `json:"handler_feedback_files" gorm:"column:handler_feedback_files;type:json"`
	DistrictFeedbackFiles JSONRaw   `json:"district_feedback_files" gorm:"column:district_feedback_files;type:json"`
	CallRecordings        JSONRaw   `json:"call_recordings" gorm:"column:call_recordings;type:json"`
	CreatedAt             time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt             time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

func (LetterAttachment) TableName() string { return "letter_attachments" }

// Feedback represents the feedbacks table
type Feedback struct {
	ID           uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	LetterNo     string    `json:"letter_no" gorm:"column:letter_no;index;size:64;not null"`
	FeedbackInfo JSONRaw   `json:"feedback_info" gorm:"column:feedback_info;type:json"`
	CreatedAt    time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

func (Feedback) TableName() string { return "feedbacks" }

// Category represents the categories table
type Category struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Level1    string    `json:"level1" gorm:"column:level1;size:64;not null"`
	Level2    string    `json:"level2" gorm:"column:level2;size:64"`
	Level3    string    `json:"level3" gorm:"column:level3;size:64"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

func (Category) TableName() string { return "categories" }

// DispatchPermission represents the dispatch_permissions table
type DispatchPermission struct {
	ID            uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	UnitName      string    `json:"unit_name" gorm:"column:unit_name;uniqueIndex;size:128;not null"`
	DispatchScope JSONRaw   `json:"dispatch_scope" gorm:"column:dispatch_scope;type:json"`
	CreatedAt     time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

func (DispatchPermission) TableName() string { return "dispatch_permissions" }

// Prompt represents the prompts table
type Prompt struct {
	ID         uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	PromptType string    `json:"prompt_type" gorm:"column:prompt_type;size:64;not null"`
	Content    string    `json:"content" gorm:"column:content;type:text"`
	CreatedAt  time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

func (Prompt) TableName() string { return "prompts" }

// SpecialFocus represents the special_focuses table
type SpecialFocus struct {
	ID          uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	TagName     string    `json:"tag_name" gorm:"column:tag_name;size:64;not null"`
	Description string    `json:"description" gorm:"column:description;type:text"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

func (SpecialFocus) TableName() string { return "special_focuses" }

// Letter status constants
const (
	StatusPreProcess           = "预处理"
	StatusCityDispatched       = "已下发至分县局/支队"
	StatusCityDirectDispatch   = "市局越级下发"
	StatusDispatched           = "已下发至处理单位"
	StatusProcessing           = "处理中"
	StatusPendingDistrictAudit = "待分县局/支队审核"
	StatusPendingCityAudit     = "待市局审核"
	StatusDone                 = "已办结"
	StatusInvalid              = "无效"
	StatusReturned             = "已退回"
	StatusExtended             = "已延期"
)
