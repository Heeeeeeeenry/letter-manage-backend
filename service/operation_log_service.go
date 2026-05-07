package service

import (
	"encoding/json"
	"time"

	"letter-manage-backend/dao"
	"letter-manage-backend/model"
)

// AddOperationLog 写入操作日志（便捷函数）
func AddOperationLog(userID uint, userName, policeNumber, action, target, targetID string, detail interface{}) {
	var detailStr string
	if detail != nil {
		if b, err := json.Marshal(detail); err == nil {
			detailStr = string(b)
		}
	}
	log := &model.OperationLog{
		UserID:       userID,
		UserName:     userName,
		PoliceNumber: policeNumber,
		Action:       action,
		Target:       target,
		TargetID:     targetID,
		Detail:       detailStr,
	}
	dao.CreateOperationLog(log)
}

// GetOperationLogs 分页+过滤查询操作日志
func GetOperationLogs(args map[string]interface{}) (map[string]interface{}, error) {
	filter := dao.OperationLogFilter{}

	if v, ok := args["page"].(float64); ok {
		filter.Page = int(v)
	}
	if v, ok := args["page_size"].(float64); ok {
		filter.PageSize = int(v)
	}
	if v, ok := args["keyword"].(string); ok {
		filter.Keyword = v
	}
	if v, ok := args["target"].(string); ok {
		filter.Target = v
	}
	if v, ok := args["target_id"].(string); ok {
		filter.TargetID = v
	}
	if v, ok := args["action"].(string); ok {
		filter.Action = v
	}
	if v, ok := args["user_name"].(string); ok {
		filter.UserName = v
	}
	if v, ok := args["start_time"].(string); ok && v != "" {
		t, err := time.ParseInLocation("2006-01-02", v, time.Local)
		if err == nil {
			filter.StartTime = &t
		}
	}
	if v, ok := args["end_time"].(string); ok && v != "" {
		t, err := time.ParseInLocation("2006-01-02", v, time.Local)
		if err == nil {
			t = t.Add(24*time.Hour - time.Second)
			filter.EndTime = &t
		}
	}

	logs, total, err := dao.GetOperationLogs(filter)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"list":      logs,
		"total":     total,
		"page":      filter.Page,
		"page_size": filter.PageSize,
	}, nil
}
