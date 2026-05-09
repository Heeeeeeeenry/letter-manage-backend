package model

// DispatchPermission represents the dispatch_permissions table
type DispatchPermission struct {
	ID            uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	UnitID        *uint  `json:"unit_id" gorm:"column:unit_id"`
	UnitName      string `json:"unit_name" gorm:"column:unit_name;size:128"`
	CanDispatchTo string `json:"can_dispatch_to" gorm:"column:dispatch_scope;type:json"`
}

func (DispatchPermission) TableName() string { return "dispatch_permissions" }

// Prompt represents the prompts table
type Prompt struct {
	ID         uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	PromptType string `json:"prompt_type" gorm:"column:prompt_type;size:64"`
	Content    string `json:"content" gorm:"column:content;type:text"`
}

func (Prompt) TableName() string { return "prompts" }
