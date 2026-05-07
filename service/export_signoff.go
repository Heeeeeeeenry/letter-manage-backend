package service

import (
	"encoding/json"
	"time"

	"letter-manage-backend/dao"
	"letter-manage-backend/model"
)

// ─── LetterSignoff 签收数据结构 ───

// LetterSignoff 信件的签收/办理/退回记录
type LetterSignoff struct {
	ID            uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	LetterNo      string    `json:"letter_no" gorm:"column:letter_no;index:idx_signoff_letter_no;size:64;not null"`
	Action        string    `json:"action" gorm:"column:action;size:32;not null"`     // dispatch / return_letter / handle_by_self / 市局下发 / submit_processing / 办案单位反馈
	FromUnit      string    `json:"from_unit" gorm:"column:from_unit;size:256"`       // 来源单位
	ToUnit        string    `json:"to_unit" gorm:"column:to_unit;size:256"`           // 目标单位
	Operator      string    `json:"operator" gorm:"column:operator;size:64"`          // 操作人
	OperatorID    uint      `json:"operator_id" gorm:"column:operator_id"`            // 操作人ID
	PrevStatus    string    `json:"prev_status" gorm:"column:prev_status;size:64"`    // 操作前状态
	CurrentStatus string    `json:"current_status" gorm:"column:current_status;size:64"` // 操作后状态/当前状态
	Remark        string    `json:"remark" gorm:"column:remark;type:text"`            // 备注
	RecordedAt    time.Time `json:"recorded_at" gorm:"column:recorded_at"`            // 操作时间
	CreatedAt     time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
}

// TableName 表名
func (LetterSignoff) TableName() string { return "letter_signoffs" }

// ─── 数据库迁移 ───

// EnsureSignoffTable 确保签收表存在
func EnsureSignoffTable() error {
	if dao.DB.Migrator().HasTable(&LetterSignoff{}) {
		return nil
	}
	return dao.DB.Migrator().CreateTable(&LetterSignoff{})
}

// ─── 从 letter_flows 提取签收数据 ───

// ExtractSignoffsFromFlows 从所有 letter_flows 的JSON中提取签收记录
// 运行时机：导出时自动执行，增量更新
func ExtractSignoffsFromFlows() (int, error) {
	if err := EnsureSignoffTable(); err != nil {
		return 0, err
	}

	// 获取所有有flow记录的信件
	var flows []model.LetterFlow
	if err := dao.DB.Find(&flows).Error; err != nil {
		return 0, err
	}

	count := 0
	for _, flow := range flows {
		if len(flow.FlowRecords) == 0 {
			continue
		}

		var records []map[string]interface{}
		if err := json.Unmarshal([]byte(flow.FlowRecords), &records); err != nil {
			continue
		}

		for _, r := range records {
			signoff := parseSignoffFromRecord(flow.LetterNo, r)
			if signoff == nil {
				continue
			}

			// upsert: 按 letter_no + action + recorded_at 去重
			var existing LetterSignoff
			err := dao.DB.Where("letter_no = ? AND action = ? AND recorded_at = ?",
				signoff.LetterNo, signoff.Action, signoff.RecordedAt).First(&existing).Error
			if err != nil {
				// 不存在，插入
				if err := dao.DB.Create(signoff).Error; err != nil {
					continue
				}
				count++
			}
		}
	}

	return count, nil
}

// parseSignoffFromRecord 从单条流转记录解析出结构化签收数据
func parseSignoffFromRecord(letterNo string, r map[string]interface{}) *LetterSignoff {
	s := &LetterSignoff{
		LetterNo: letterNo,
	}

	// 统一提取各字段（兼容新旧两种字段名格式）
	s.Action = getStrField(r, "action", "操作类型")
	s.FromUnit = getStrField(r, "from_unit", "操作前单位")
	s.ToUnit = getStrField(r, "to_unit", "目标单位")
	s.Operator = getStrField(r, "operator", "操作人姓名")
	s.Remark = getStrField(r, "remark", "备注")
	s.CurrentStatus = getStrField(r, "status", "操作后状态")
	s.PrevStatus = getStrField(r, "操作前状态", "")

	// operator_id
	if v, ok := r["operator_id"].(float64); ok {
		s.OperatorID = uint(v)
	}

	// 时间解析（兼容多种格式）
	timeStr := getStrField(r, "timestamp", "操作时间", "recorded_at")
	if timeStr != "" {
		formats := []string{
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05",
			"2006-01-02 15:04:05.000",
		}
		for _, f := range formats {
			if t, err := time.ParseInLocation(f, timeStr, time.Local); err == nil {
				s.RecordedAt = t
				break
			}
		}
	}

	// 如果 action 还是空，但有操作类型字段，尝试从操作类型映射
	if s.Action == "" {
		if opType := getStrField(r, "操作类型", ""); opType != "" {
			s.Action = mapActionType(opType)
		}
	}

	// 关键字段不能为空
	if s.Action == "" || s.RecordedAt.IsZero() {
		return nil
	}

	return s
}

// mapActionType 将中文操作类型映射为标准化action
func mapActionType(opType string) string {
	mapping := map[string]string{
		"生成":         "create",
		"市局下发":       "dispatch",
		"自行处理":       "handle_by_self",
		"办案单位反馈":     "feedback",
		"退回":         "return_letter",
		"签收":         "signoff",
		"核查":         "verify",
		"审批":         "approve",
	}
	if v, ok := mapping[opType]; ok {
		return v
	}
	return opType
}

// getStrField 从map中按多个key依次查找
func getStrField(r map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if k == "" {
			continue
		}
		if v, ok := r[k].(string); ok {
			return v
		}
	}
	return ""
}

// ─── 查询接口（供导出用） ───

// GetSignoffsByLetterNo 获取某封信件的签收记录
func GetSignoffsByLetterNo(letterNo string) ([]LetterSignoff, error) {
	var signoffs []LetterSignoff
	err := dao.DB.Where("letter_no = ?", letterNo).
		Order("recorded_at ASC").
		Find(&signoffs).Error
	return signoffs, err
}

// GetSignoffStats 获取签收统计（导出用）
type SignoffStats struct {
	TotalDispatches  int64 // 总下发次数
	TotalReturns     int64 // 总退回次数
	UnsignoffCount   int64 // 未签收数
	OverdueSignoff   int64 // 超时签收数
	AvgSignoffHours  float64 // 平均签收耗时(小时)
}

// GetUnitSignoffStats 按分县局统计签收情况
func GetUnitSignoffStats(unitLevel1 string, startTime, endTime time.Time) SignoffStats {
	var stats SignoffStats

	// 统计下发次数
	dao.DB.Model(&LetterSignoff{}).
		Where("action = ? AND recorded_at >= ? AND recorded_at < ?", "dispatch", startTime, endTime).
		Count(&stats.TotalDispatches)

	// 统计退回次数
	dao.DB.Model(&LetterSignoff{}).
		Where("action = ? AND recorded_at >= ? AND recorded_at < ?", "return_letter", startTime, endTime).
		Count(&stats.TotalReturns)

	// 统计未签收
	// 查找被下发但无后续签收/处理记录的信件
	_ = unitLevel1

	return stats
}

// ─── 初始化 ───

// InitSignoffExtraction 初始化签收数据提取（应用启动时调用）
func InitSignoffExtraction() {
	if err := EnsureSignoffTable(); err != nil {
		return
	}
	ExtractSignoffsFromFlows()
}
