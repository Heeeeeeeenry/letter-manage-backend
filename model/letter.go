package model

import "time"

// Letter represents the letters table
type Letter struct {
	ID              uint        `json:"id" gorm:"primaryKey;autoIncrement"`
	LetterNo        string      `json:"letter_no" gorm:"column:letter_no;uniqueIndex;size:64;not null"`
	CitizenName     string      `json:"citizen_name" gorm:"column:citizen_name;size:64"`
	Phone           string      `json:"phone" gorm:"column:phone;size:32"`
	IDCard          string      `json:"id_card" gorm:"column:id_card;size:32"`
	Channel         ChannelCode `json:"channel" gorm:"column:channel;type:tinyint"`
	CategoryID      *uint       `json:"category_id" gorm:"column:category_id"`
	Category        *Category   `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	Content         string      `json:"content" gorm:"column:content;type:text"`
	CurrentUnitID   *uint       `json:"current_unit_id" gorm:"column:current_unit_id"`
	CurrentUnitObj  *Unit       `json:"current_unit,omitempty" gorm:"foreignKey:CurrentUnitID"`
	HandlerUserID   *uint       `json:"handler_user_id" gorm:"column:handler_user_id"`
	HandlerUnitID   *uint       `json:"handler_unit_id" gorm:"column:handler_unit_id"`
	CurrentStatus   StatusCode  `json:"current_status" gorm:"column:current_status;type:tinyint"`
	CurrentOperator string      `json:"current_operator" gorm:"column:current_operator;size:64"`
	DeadlineAt      *time.Time  `json:"deadline_at" gorm:"column:deadline_at"`
	FocusID         *uint       `json:"focus_id,omitempty" gorm:"-"`
	CreatedAt       time.Time   `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       time.Time   `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

func (Letter) TableName() string { return "letters" }

// GetChannelName returns the Chinese name for the channel code
func (l Letter) GetChannelName() string {
	if name, ok := ChannelToName[l.Channel]; ok {
		return name
	}
	return ""
}

// GetStatusName returns the Chinese name for the status code
func (l Letter) GetStatusName() string {
	if name, ok := StatusCodeToName[l.CurrentStatus]; ok {
		return name
	}
	return ""
}

// LetterAttachment represents the letter_attachments table
type LetterAttachment struct {
	ID                    uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	LetterNo              string    `json:"letter_no" gorm:"column:letter_no;uniqueIndex;size:64;not null"`
	CityDispatchFiles     JSONRaw   `json:"city_dispatch_files" gorm:"column:city_dispatch_files;type:json"`
	DistrictDispatchFiles JSONRaw   `json:"district_dispatch_files" gorm:"column:district_dispatch_files;type:json"`
	HandlerFeedbackFiles  JSONRaw   `json:"handler_feedback_files" gorm:"column:handler_feedback_files;type:json"`
	DistrictFeedbackFiles JSONRaw   `json:"district_feedback_files" gorm:"column:district_feedback_files;type:json"`
	CallRecordings        JSONRaw   `json:"call_recordings" gorm:"column:call_recordings;type:json"`
	CitizenFiles          JSONRaw   `json:"citizen_files" gorm:"column:citizen_files;type:json"`
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

// LetterFlow represents the letter_flows table
type LetterFlow struct {
	ID          uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	LetterNo    string    `json:"letter_no" gorm:"column:letter_no;uniqueIndex;size:64;not null"`
	FlowRecords JSONRaw   `json:"flow_records" gorm:"column:flow_records;type:json"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

func (LetterFlow) TableName() string { return "letter_flows" }

// SpecialFocus 专项关注
type SpecialFocus struct {
	ID          uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string    `json:"name" gorm:"column:tag_name;size:64;not null"`
	Description string    `json:"description" gorm:"column:description;type:text"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

func (SpecialFocus) TableName() string { return "special_focuses" }

// LetterSpecialFocus 信件-专项关注绑定关系（中间表）
type LetterSpecialFocus struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	LetterNo  string    `json:"letter_no" gorm:"column:letter_no;index;size:64;not null"`
	FocusID   uint      `json:"focus_id" gorm:"column:focus_id;index;not null"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
}

func (LetterSpecialFocus) TableName() string { return "letter_special_focuses" }

// ──── Channel Code Mapping ────

type ChannelCode int

const (
	ChannelCitizenReport ChannelCode = 1 // 市民上报
	ChannelDirectorMail  ChannelCode = 2 // 局长信箱
	ChannelVisit         ChannelCode = 3 // 来访
	ChannelPhone         ChannelCode = 4 // 电话
	ChannelMail          ChannelCode = 5 // 信件
	ChannelWeb           ChannelCode = 6 // 网络
	ChannelOther         ChannelCode = 7 // 其他
)

var ChannelToName = map[ChannelCode]string{
	ChannelCitizenReport: "市民上报",
	ChannelDirectorMail:  "局长信箱",
	ChannelVisit:         "来访",
	ChannelPhone:         "电话",
	ChannelMail:          "信件",
	ChannelWeb:           "网络",
	ChannelOther:         "其他",
	8:                    "12345热线",
	9:                    "12389举报",
	10:                   "网上信访",
	11:                   "上级交办",
}

var ChannelNameToCode = map[string]ChannelCode{
	"市民上报": ChannelCitizenReport,
	"局长信箱": ChannelDirectorMail,
	"来访":   ChannelVisit,
	"电话":   ChannelPhone,
	"信件":   ChannelMail,
	"网络":   ChannelWeb,
	"其他":   ChannelOther,
	"12345热线": 8,
	"12389举报": 9,
	"网上信访": 10,
	"上级交办": 11,
}

// ──── Status Code Mapping ────

type StatusCode int

const (
	StatusCodePreProcess           StatusCode = 1
	StatusCodePendingDisDispatch   StatusCode = 2
	StatusCodeCityDispatched       StatusCode = 3
	StatusCodeCityDirectDispatch   StatusCode = 4
	StatusCodeDispatched           StatusCode = 5
	StatusCodeProcessing           StatusCode = 6
	StatusCodePendingVerification  StatusCode = 7
	StatusCodePendingDistrictAudit StatusCode = 8
	StatusCodePendingCityAudit     StatusCode = 9
	StatusCodeDone                 StatusCode = 10
	StatusCodeInvalid              StatusCode = 11
	StatusCodeReturned             StatusCode = 12
	StatusCodeExtended             StatusCode = 13
)

// Letter status string constants (kept for backward compat with service layer)
const (
	StatusPreProcess           = "预处理"
	StatusPendingDistrictDispatch = "待区县局下发"
	StatusCityDispatched       = "已下发至分县局/支队"
	StatusCityDirectDispatch   = "市局越级下发"
	StatusDispatched           = "已下发至处理单位"
	StatusProcessing           = "处理中"
	StatusPendingVerification  = "待核查"
	StatusPendingDistrictAudit = "待分县局/支队审核"
	StatusPendingCityAudit     = "待市局审核"
	StatusDone                 = "已办结"
	StatusInvalid              = "无效"
	StatusReturned             = "已退回"
	StatusExtended             = "已延期"
)

var StatusCodeToName = map[StatusCode]string{
	StatusCodePreProcess:           StatusPreProcess,
	StatusCodePendingDisDispatch:   StatusPendingDistrictDispatch,
	StatusCodeCityDispatched:       StatusCityDispatched,
	StatusCodeCityDirectDispatch:   StatusCityDirectDispatch,
	StatusCodeDispatched:           StatusDispatched,
	StatusCodeProcessing:           StatusProcessing,
	StatusCodePendingVerification:  StatusPendingVerification,
	StatusCodePendingDistrictAudit: StatusPendingDistrictAudit,
	StatusCodePendingCityAudit:     StatusPendingCityAudit,
	StatusCodeDone:                 StatusDone,
	StatusCodeInvalid:              StatusInvalid,
	StatusCodeReturned:             StatusReturned,
	StatusCodeExtended:             StatusExtended,
}

var StatusNameToCode = map[string]StatusCode{
	StatusPreProcess:              StatusCodePreProcess,
	StatusPendingDistrictDispatch: StatusCodePendingDisDispatch,
	StatusCityDispatched:          StatusCodeCityDispatched,
	StatusCityDirectDispatch:      StatusCodeCityDirectDispatch,
	StatusDispatched:              StatusCodeDispatched,
	StatusProcessing:              StatusCodeProcessing,
	StatusPendingVerification:     StatusCodePendingVerification,
	StatusPendingDistrictAudit:    StatusCodePendingDistrictAudit,
	StatusPendingCityAudit:        StatusCodePendingCityAudit,
	StatusDone:                    StatusCodeDone,
	StatusInvalid:                 StatusCodeInvalid,
	StatusReturned:                StatusCodeReturned,
	StatusExtended:                StatusCodeExtended,
}

var LegacyStatusToCode = map[string]StatusCode{
	"待受理":         StatusCodePreProcess,
	"市局下发至区县局/支队": StatusCodeCityDispatched,
}

// OperationLog 操作日志
type OperationLog struct {
	ID           uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID       uint      `json:"user_id" gorm:"column:user_id;index;not null"`
	UserName     string    `json:"user_name" gorm:"column:user_name;size:64"`
	PoliceNumber string    `json:"police_number" gorm:"column:police_number;size:32"`
	Action       string    `json:"action" gorm:"column:action;size:32;not null"`
	Target       string    `json:"target" gorm:"column:target;size:64;not null"`
	TargetID     string    `json:"target_id" gorm:"column:target_id;size:64"`
	Detail       string    `json:"detail" gorm:"column:detail;type:text"`
	CreatedAt    time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
}

func (OperationLog) TableName() string { return "operation_logs" }

// LetterSignoff 信件的签收/办理/退回记录（letter_signoffs 表）
type LetterSignoff struct {
	ID            uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	LetterNo      string    `json:"letter_no" gorm:"column:letter_no;index:idx_signoff_letter_no;size:64;not null"`
	Action        string    `json:"action" gorm:"column:action;size:32;not null"`
	FromUnit      string    `json:"from_unit" gorm:"column:from_unit;size:256"`
	ToUnit        string    `json:"to_unit" gorm:"column:to_unit;size:256"`
	Operator      string    `json:"operator" gorm:"column:operator;size:64"`
	OperatorID    uint      `json:"operator_id" gorm:"column:operator_id"`
	PrevStatus    string    `json:"prev_status" gorm:"column:prev_status;size:64"`
	CurrentStatus string    `json:"current_status" gorm:"column:current_status;size:64"`
	Remark        string    `json:"remark" gorm:"column:remark;type:text"`
	RecordedAt    time.Time `json:"recorded_at" gorm:"column:recorded_at"`
	CreatedAt     time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
}

func (LetterSignoff) TableName() string { return "letter_signoffs" }
