package dao

import (
	"time"

	"letter-manage-backend/model"
)

func CreateOperationLog(log *model.OperationLog) error {
	return DB.Create(log).Error
}

// OperationLogFilter holds all optional filter parameters
type OperationLogFilter struct {
	Keyword   string
	Target    string
	TargetID  string
	Action    string
	UserName  string
	StartTime *time.Time
	EndTime   *time.Time
	Page      int
	PageSize  int
}

func GetOperationLogs(filter OperationLogFilter) ([]model.OperationLog, int64, error) {
	var logs []model.OperationLog
	var total int64
	query := DB.Model(&model.OperationLog{})

	// Keyword: fuzzy search across multiple fields
	if filter.Keyword != "" {
		like := "%" + filter.Keyword + "%"
		query = query.Where(
			"user_name LIKE ? OR police_number LIKE ? OR action LIKE ? OR target LIKE ? OR target_id LIKE ? OR detail LIKE ?",
			like, like, like, like, like, like,
		)
	}

	// Exact field filters
	if filter.Target != "" {
		query = query.Where("target = ?", filter.Target)
	}
	if filter.TargetID != "" {
		query = query.Where("target_id LIKE ?", "%"+filter.TargetID+"%")
	}
	if filter.Action != "" {
		query = query.Where("action = ?", filter.Action)
	}
	if filter.UserName != "" {
		query = query.Where("user_name LIKE ? OR police_number LIKE ?", "%"+filter.UserName+"%", "%"+filter.UserName+"%")
	}
	if filter.StartTime != nil {
		query = query.Where("created_at >= ?", filter.StartTime)
	}
	if filter.EndTime != nil {
		query = query.Where("created_at <= ?", filter.EndTime)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 20
	}
	offset := (filter.Page - 1) * filter.PageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(filter.PageSize).Find(&logs).Error
	return logs, total, err
}
