package model

// DispatchTarget represents a dispatch permission: dispatcher → target unit mapping
type DispatchTarget struct {
	ID                uint  `json:"id" gorm:"primaryKey;autoIncrement"`
	DispatcherUnitID  uint  `json:"dispatcher_unit_id" gorm:"column:dispatcher_unit_id;not null"`
	TargetUnitID      uint  `json:"target_unit_id" gorm:"column:target_unit_id;not null"`
	// Preloaded relations
	TargetUnit        *Unit `json:"target_unit,omitempty" gorm:"foreignKey:TargetUnitID;references:ID"`
}

func (DispatchTarget) TableName() string { return "dispatch_targets" }

// Prompt represents the prompts table
type Prompt struct {
	ID         uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	PromptType string `json:"prompt_type" gorm:"column:prompt_type;size:64"`
	Content    string `json:"content" gorm:"column:content;type:text"`
}

func (Prompt) TableName() string { return "prompts" }
