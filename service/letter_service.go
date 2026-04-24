package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"letter-manage-backend/dao"
	"letter-manage-backend/model"
)

func GenerateLetterNo() string {
	return fmt.Sprintf("XJ%d", time.Now().UnixNano()/int64(time.Millisecond))
}

func GetLetterList(args map[string]interface{}, unitName string, permLevel string) (map[string]interface{}, error) {
	// Remove order field from args to prevent SQL injection
	delete(args, "order")
	filter := dao.LetterFilter{}
	if v, ok := args["status"].(string); ok {
		filter.Status = v
	}
	if v, ok := args["category_l1"].(string); ok {
		filter.CategoryL1 = v
	}
	if v, ok := args["category_l2"].(string); ok {
		filter.CategoryL2 = v
	}
	if v, ok := args["category_l3"].(string); ok {
		filter.CategoryL3 = v
	}
	if v, ok := args["keyword"].(string); ok {
		filter.Keyword = v
	}
	if v, ok := args["letter_no"].(string); ok {
		filter.LetterNo = v
	}
	if v, ok := args["citizen_name"].(string); ok {
		filter.CitizenName = v
	}
	if v, ok := args["phone"].(string); ok {
		filter.Phone = v
	}
	if v, ok := args["id_card"].(string); ok {
		filter.IDCard = v
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
	if v, ok := args["unit_name"].(string); ok {
		filter.UnitName = v
	}
	filter.Page = 1
	filter.PageSize = 20
	if v, ok := args["page"].(float64); ok {
		filter.Page = int(v)
	}
	if v, ok := args["page_size"].(float64); ok {
		filter.PageSize = int(v)
	}

	// 权限数据隔离：根据用户权限级别自动添加单位过滤
	switch permLevel {
	case "CITY":
		// 市局：可见所有信件，不过滤
	case "DISTRICT":
		// 区县局：可见本单位及下属单位的信件
		subUnits := getSubordinateUnitNames(unitName)
		if len(subUnits) > 0 {
			filter.UnitNames = subUnits
		} else {
			filter.UnitName = unitName
		}
	default:
		// OFFICER：仅可见本单位信件
		filter.UnitName = unitName
	}

	letters, total, err := dao.GetLetterList(filter)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"list":      letters,
		"total":     total,
		"page":      filter.Page,
		"page_size": filter.PageSize,
	}, nil
}

// getSubordinateUnitNames 获取某单位及其下属所有单位的短名称列表
func getSubordinateUnitNames(unitName string) []string {
	return dao.GetSubordinateUnitNames(unitName)
}

func GetDispatchList(unitName string, permLevel string, args map[string]interface{}) (map[string]interface{}, error) {
	page := 1
	pageSize := 20
	if v, ok := args["page"].(float64); ok {
		page = int(v)
	}
	if v, ok := args["page_size"].(float64); ok {
		pageSize = int(v)
	}
	letters, total, err := dao.GetDispatchList(unitName, permLevel, page, pageSize)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"list":      letters,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}, nil
}

func GetProcessingList(unitName string, args map[string]interface{}) (map[string]interface{}, error) {
	page := 1
	pageSize := 20
	if v, ok := args["page"].(float64); ok {
		page = int(v)
	}
	if v, ok := args["page_size"].(float64); ok {
		pageSize = int(v)
	}
	letters, total, err := dao.GetProcessingList(unitName, page, pageSize)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"list":      letters,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}, nil
}

func GetAuditList(unitName string, permLevel string, args map[string]interface{}) (map[string]interface{}, error) {
	page := 1
	pageSize := 20
	if v, ok := args["page"].(float64); ok {
		page = int(v)
	}
	if v, ok := args["page_size"].(float64); ok {
		pageSize = int(v)
	}
	letters, total, err := dao.GetAuditList(unitName, permLevel, page, pageSize)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"list":      letters,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}, nil
}

// GetLettersByPhone 获取某手机号的所有信件（带权限过滤）
func GetLettersByPhone(phone, unitName, permLevel string) ([]model.Letter, error) {
	letters, err := dao.GetLettersByPhone(phone)
	if err != nil {
		return nil, err
	}
	return filterLettersByPermission(letters, unitName, permLevel), nil
}

// GetLettersByIDCard 获取某身份证的所有信件（带权限过滤）
func GetLettersByIDCard(idCard, unitName, permLevel string) ([]model.Letter, error) {
	letters, err := dao.GetLettersByIDCard(idCard)
	if err != nil {
		return nil, err
	}
	return filterLettersByPermission(letters, unitName, permLevel), nil
}

func GetLetterDetail(letterNo string, unitName string, permLevel string) (map[string]interface{}, error) {
	letter, err := dao.GetLetterByNo(letterNo)
	if err != nil {
		return nil, err
	}
	// 权限检查：验证用户是否有权访问该信件
	if !canAccessLetter(*letter, unitName, permLevel) {
		return nil, errors.New("无权访问该信件")
	}
	flow, _ := dao.GetFlowByLetterNo(letterNo)
	att, _ := dao.GetAttachmentByLetterNo(letterNo)
	feedbacks, _ := dao.GetFeedbacksByLetterNo(letterNo)
	// history letters by same phone
	var history []model.Letter
	if letter.Phone != "" {
		history, _ = dao.GetLettersByPhone(letter.Phone)
		// 历史信件也要做权限过滤
		history = filterLettersByPermission(history, unitName, permLevel)
	}
	return map[string]interface{}{
		"letter":    letter,
		"flow":      flow,
		"files":     att,
		"history":   history,
		"feedbacks": feedbacks,
	}, nil
}

// canAccessLetter 检查用户是否有权访问某封信件
func canAccessLetter(letter model.Letter, unitName string, permLevel string) bool {
	switch permLevel {
	case "CITY":
		return true
	case "DISTRICT":
		// DISTRICT 可以访问本单位及下属单位的信件
		subUnits := dao.GetSubordinateUnitNames(unitName)
		for _, u := range subUnits {
			if u == letter.CurrentUnit {
				return true
			}
		}
		return letter.CurrentUnit == unitName
	default:
		// OFFICER 只能访问本单位的信件
		return letter.CurrentUnit == unitName
	}
}

// filterLettersByPermission 根据权限过滤信件列表
func filterLettersByPermission(letters []model.Letter, unitName string, permLevel string) []model.Letter {
	var filtered []model.Letter
	for _, l := range letters {
		if canAccessLetter(l, unitName, permLevel) {
			filtered = append(filtered, l)
		}
	}
	return filtered
}

func CreateLetter(args map[string]interface{}) (*model.Letter, error) {
	letter := &model.Letter{}
	letter.LetterNo = GenerateLetterNo()
	if v, ok := args["citizen_name"].(string); ok {
		letter.CitizenName = v
	}
	if v, ok := args["phone"].(string); ok {
		letter.Phone = v
	}
	if v, ok := args["id_card"].(string); ok {
		letter.IDCard = v
	}
	if v, ok := args["channel"].(string); ok {
		letter.Channel = v
	}
	if v, ok := args["category_l1"].(string); ok {
		letter.CategoryL1 = v
	}
	if v, ok := args["category_l2"].(string); ok {
		letter.CategoryL2 = v
	}
	if v, ok := args["category_l3"].(string); ok {
		letter.CategoryL3 = v
	}
	if v, ok := args["content"].(string); ok {
		letter.Content = v
	}
	if v, ok := args["current_unit"].(string); ok {
		letter.CurrentUnit = v
	}
	letter.CurrentStatus = model.StatusPreProcess
	if v, ok := args["received_at"].(string); ok && v != "" {
		t, err := time.ParseInLocation("2006-01-02", v, time.Local)
		if err == nil {
			letter.ReceivedAt = t
		}
	} else {
		letter.ReceivedAt = time.Now()
	}
	if v, ok := args["special_tags"]; ok {
		b, _ := json.Marshal(v)
		letter.SpecialTags = model.JSONRaw(b)
	}
	if err := dao.CreateLetter(letter); err != nil {
		return nil, err
	}
	// create empty attachment record
	att := &model.LetterAttachment{
		LetterNo:              letter.LetterNo,
		CityDispatchFiles:     model.JSONRaw("[]"),
		DistrictDispatchFiles: model.JSONRaw("[]"),
		HandlerFeedbackFiles:  model.JSONRaw("[]"),
		DistrictFeedbackFiles: model.JSONRaw("[]"),
		CallRecordings:        model.JSONRaw("[]"),
	}
	dao.UpsertAttachment(att)
	return letter, nil
}

func UpdateLetter(args map[string]interface{}) error {
	letterNo, ok := args["letter_no"].(string)
	if !ok || letterNo == "" {
		return errors.New("letter_no required")
	}
	letter, err := dao.GetLetterByNo(letterNo)
	if err != nil {
		return err
	}
	if v, ok := args["citizen_name"].(string); ok {
		letter.CitizenName = v
	}
	if v, ok := args["phone"].(string); ok {
		letter.Phone = v
	}
	if v, ok := args["id_card"].(string); ok {
		letter.IDCard = v
	}
	if v, ok := args["channel"].(string); ok {
		letter.Channel = v
	}
	if v, ok := args["category_l1"].(string); ok {
		letter.CategoryL1 = v
	}
	if v, ok := args["category_l2"].(string); ok {
		letter.CategoryL2 = v
	}
	if v, ok := args["category_l3"].(string); ok {
		letter.CategoryL3 = v
	}
	if v, ok := args["content"].(string); ok {
		letter.Content = v
	}
	if v, ok := args["current_unit"].(string); ok {
		letter.CurrentUnit = v
	}
	if v, ok := args["current_status"].(string); ok {
		letter.CurrentStatus = v
	}
	if v, ok := args["received_at"].(string); ok && v != "" {
		t, err := time.ParseInLocation("2006-01-02", v, time.Local)
		if err == nil {
			letter.ReceivedAt = t
		}
	}
	if v, ok := args["special_tags"]; ok {
		b, _ := json.Marshal(v)
		letter.SpecialTags = model.JSONRaw(b)
	}
	return dao.UpdateLetter(letter)
}

func DeleteLetter(args map[string]interface{}) error {
	idF, ok := args["id"].(float64)
	if !ok {
		return errors.New("id required")
	}
	return dao.DeleteLetter(uint(idF))
}

func UpdateLetterStatus(args map[string]interface{}, operator *model.PoliceUser) error {
	letterNo, ok := args["letter_no"].(string)
	if !ok || letterNo == "" {
		return errors.New("letter_no required")
	}
	status, ok := args["status"].(string)
	if !ok || status == "" {
		return errors.New("status required")
	}
	unitName, _ := args["unit_name"].(string)
	remark, _ := args["remark"].(string)

	if err := dao.UpdateLetterStatus(letterNo, status, unitName); err != nil {
		return err
	}

	// append flow record
	flowRecord := map[string]interface{}{
		"status":    status,
		"unit":      unitName,
		"remark":    remark,
		"operator":  operator.Name,
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
	}
	return appendFlowRecord(letterNo, flowRecord)
}

func appendFlowRecord(letterNo string, record map[string]interface{}) error {
	flow, err := dao.GetFlowByLetterNo(letterNo)
	if err != nil {
		return err
	}
	var records []interface{}
	if flow != nil && len(flow.FlowRecords) > 0 {
		json.Unmarshal([]byte(flow.FlowRecords), &records)
	}
	records = append(records, record)
	b, _ := json.Marshal(records)
	return dao.UpsertLetterFlow(letterNo, b)
}

func DispatchLetter(args map[string]interface{}, operator *model.PoliceUser) error {
	letterNo, ok := args["letter_no"].(string)
	if !ok || letterNo == "" {
		return errors.New("letter_no required")
	}
	targetUnit, ok := args["target_unit"].(string)
	if !ok || targetUnit == "" {
		return errors.New("target_unit required")
	}
	remark, _ := args["remark"].(string)

	// permission check
	canDispatch, err := CheckDispatchPermission(operator, targetUnit)
	if err != nil {
		return err
	}
	if !canDispatch {
		return errors.New("没有向该单位下发的权限")
	}

	letter, err := dao.GetLetterByNo(letterNo)
	if err != nil {
		return err
	}

	var newStatus string
	switch operator.PermissionLevel {
	case model.PermissionCity:
		newStatus = model.StatusCityDispatched
	case model.PermissionDistrict:
		newStatus = model.StatusDispatched
	default:
		return errors.New("无下发权限")
	}

	if err := dao.UpdateLetterStatus(letterNo, newStatus, targetUnit); err != nil {
		return err
	}

	record := map[string]interface{}{
		"action":      "dispatch",
		"status":      newStatus,
		"from_unit":   letter.CurrentUnit,
		"to_unit":     targetUnit,
		"remark":      remark,
		"operator":    operator.Name,
		"operator_id": operator.ID,
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
	}
	if err := appendFlowRecord(letterNo, record); err != nil {
		return err
	}
	// 每次下发重新设置 30 分钟处理倒计时
	deadline := time.Now().Add(30 * time.Minute)
	return dao.UpdateLetterDeadline(letterNo, &deadline)
}

func CheckDispatchPermission(operator *model.PoliceUser, targetUnit string) (bool, error) {
	switch operator.PermissionLevel {
	case model.PermissionCity:
		return true, nil
	case model.PermissionDistrict:
		// district can dispatch to self or subordinates
		if operator.UnitName == targetUnit {
			return true, nil
		}
		// check units table: target unit's level1 or level2 matches operator's unit
		units, err := dao.GetAllUnits()
		if err != nil {
			return false, err
		}
		for _, u := range units {
			if u.Level2 == operator.UnitName || u.Level1 == operator.UnitName {
				fullName := u.Level1
				if u.Level2 != "" {
					fullName = u.Level2
				}
				if u.Level3 != "" {
					fullName = u.Level3
				}
				if fullName == targetUnit {
					return true, nil
				}
			}
		}
		return false, nil
	default:
		// check dispatch_permissions table
		perm, err := dao.GetDispatchPermissionByUnit(operator.UnitName)
		if err != nil {
			return false, err
		}
		if perm == nil {
			return false, nil
		}
		var scope []string
		if err := json.Unmarshal([]byte(perm.DispatchScope), &scope); err != nil {
			return false, nil
		}
		for _, s := range scope {
			if s == targetUnit {
				return true, nil
			}
		}
		return false, nil
	}
}

func MarkInvalid(args map[string]interface{}, operator *model.PoliceUser) error {
	letterNo, ok := args["letter_no"].(string)
	if !ok || letterNo == "" {
		return errors.New("letter_no required")
	}
	remark, _ := args["remark"].(string)
	if err := dao.UpdateLetterStatus(letterNo, model.StatusInvalid, ""); err != nil {
		return err
	}
	record := map[string]interface{}{
		"action":    "mark_invalid",
		"status":    model.StatusInvalid,
		"remark":    remark,
		"operator":  operator.Name,
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
	}
	return appendFlowRecord(letterNo, record)
}

func SubmitProcessing(args map[string]interface{}, operator *model.PoliceUser) error {
	letterNo, ok := args["letter_no"].(string)
	if !ok || letterNo == "" {
		return errors.New("letter_no required")
	}
	remark, _ := args["remark"].(string)
	if err := dao.UpdateLetterStatus(letterNo, model.StatusFeedback, ""); err != nil {
		return err
	}
	record := map[string]interface{}{
		"action":    "submit_processing",
		"status":    model.StatusFeedback,
		"remark":    remark,
		"operator":  operator.Name,
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
	}
	return appendFlowRecord(letterNo, record)
}

func HandleBySelf(args map[string]interface{}, operator *model.PoliceUser) error {
	letterNo, ok := args["letter_no"].(string)
	if !ok || letterNo == "" {
		return errors.New("letter_no required")
	}
	remark, _ := args["remark"].(string)
	if err := dao.UpdateLetterStatus(letterNo, model.StatusProcessing, operator.UnitName); err != nil {
		return err
	}
	record := map[string]interface{}{
		"action":    "handle_by_self",
		"status":    model.StatusProcessing,
		"unit":      operator.UnitName,
		"remark":    remark,
		"operator":  operator.Name,
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
	}
	return appendFlowRecord(letterNo, record)
}

func ReturnLetter(args map[string]interface{}, operator *model.PoliceUser) error {
	letterNo, ok := args["letter_no"].(string)
	if !ok || letterNo == "" {
		return errors.New("letter_no required")
	}
	remark, _ := args["remark"].(string)

	// 获取当前信件
	letter, err := dao.GetLetterByNo(letterNo)
	if err != nil {
		return err
	}
	if letter == nil {
		return errors.New("letter not found")
	}

	// 获取流转记录，追溯上一个状态/人/单位
	var prevUnit, prevStatus, prevOperator string
	flow, err := dao.GetFlowByLetterNo(letterNo)
	if err != nil {
		return err
	}
	if flow != nil {
		var records []map[string]interface{}
		if err := json.Unmarshal([]byte(flow.FlowRecords), &records); err == nil {
			// 倒序查找最后一次 dispatch 操作
			for i := len(records) - 1; i >= 0; i-- {
				r := records[i]
				action, _ := r["action"].(string)
			if action == "dispatch" {
					prevUnit, _ = r["from_unit"].(string)
					prevOperator, _ = r["operator"].(string)
					// 看 dispatch 记录之前的记录状态
					if i > 0 {
						prevRec := records[i-1]
						prevStatus, _ = prevRec["status"].(string)
						if prevStatus == "" {
							prevStatus, _ = prevRec["操作后状态"].(string)
						}
					}
					break
				}
			}
		}
	}
	if prevStatus == "" {
		prevStatus = model.StatusReturned
	}

	// 更新信件为上一个状态
	updates := map[string]interface{}{
		"current_status":   prevStatus,
		"current_unit":     prevUnit,
		"current_operator": prevOperator,
	}
	if err := dao.UpdateLetterFields(letterNo, updates); err != nil {
		return err
	}

	// 清除 deadline（等待下次下发重新计时）
	if err := dao.UpdateLetterDeadline(letterNo, nil); err != nil {
		return err
	}

	// 追加退回记录（保留完整历史）
	record := map[string]interface{}{
		"action":        "return_letter",
		"status":        prevStatus,
		"from_unit":     letter.CurrentUnit,
		"to_unit":       prevUnit,
		"from_operator": letter.CurrentOperator,
		"to_operator":   prevOperator,
		"remark":        remark,
		"operator":      operator.Name,
		"operator_id":   operator.ID,
		"timestamp":     time.Now().Format("2006-01-02 15:04:05"),
	}
	return appendFlowRecord(letterNo, record)
}

func AuditApprove(args map[string]interface{}, operator *model.PoliceUser) error {
	letterNo, ok := args["letter_no"].(string)
	if !ok || letterNo == "" {
		return errors.New("letter_no required")
	}
	remark, _ := args["remark"].(string)
	if err := dao.UpdateLetterStatus(letterNo, model.StatusDone, ""); err != nil {
		return err
	}
	record := map[string]interface{}{
		"action":    "audit_approve",
		"status":    model.StatusDone,
		"remark":    remark,
		"operator":  operator.Name,
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
	}
	return appendFlowRecord(letterNo, record)
}

func AuditReject(args map[string]interface{}, operator *model.PoliceUser) error {
	letterNo, ok := args["letter_no"].(string)
	if !ok || letterNo == "" {
		return errors.New("letter_no required")
	}
	remark, _ := args["remark"].(string)
	if err := dao.UpdateLetterStatus(letterNo, model.StatusProcessing, ""); err != nil {
		return err
	}
	record := map[string]interface{}{
		"action":    "audit_reject",
		"status":    model.StatusProcessing,
		"remark":    remark,
		"operator":  operator.Name,
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
	}
	return appendFlowRecord(letterNo, record)
}

func GetStatistics(unitName string, permLevel string) (map[string]interface{}, error) {
	statusStats, err := dao.GetLetterStatusStats()
	if err != nil {
		return nil, err
	}
	channelStats, err := dao.GetLetterChannelStats()
	if err != nil {
		return nil, err
	}
	monthStats, err := dao.GetLetterMonthStats()
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"status_stats":  statusStats,
		"channel_stats": channelStats,
		"month_stats":   monthStats,
	}, nil
}

func GetAttachments(letterNo string) (*model.LetterAttachment, error) {
	att, err := dao.GetAttachmentByLetterNo(letterNo)
	if err != nil {
		return nil, err
	}
	if att == nil {
		att = &model.LetterAttachment{
			LetterNo:              letterNo,
			CityDispatchFiles:     model.JSONRaw("[]"),
			DistrictDispatchFiles: model.JSONRaw("[]"),
			HandlerFeedbackFiles:  model.JSONRaw("[]"),
			DistrictFeedbackFiles: model.JSONRaw("[]"),
			CallRecordings:        model.JSONRaw("[]"),
		}
	}
	return att, nil
}

func UpdateAttachments(args map[string]interface{}) error {
	letterNo, ok := args["letter_no"].(string)
	if !ok || letterNo == "" {
		return errors.New("letter_no required")
	}
	att := &model.LetterAttachment{LetterNo: letterNo}
	if v, ok := args["city_dispatch_files"]; ok {
		b, _ := json.Marshal(v)
		att.CityDispatchFiles = model.JSONRaw(b)
	} else {
		att.CityDispatchFiles = model.JSONRaw("[]")
	}
	if v, ok := args["district_dispatch_files"]; ok {
		b, _ := json.Marshal(v)
		att.DistrictDispatchFiles = model.JSONRaw(b)
	} else {
		att.DistrictDispatchFiles = model.JSONRaw("[]")
	}
	if v, ok := args["handler_feedback_files"]; ok {
		b, _ := json.Marshal(v)
		att.HandlerFeedbackFiles = model.JSONRaw(b)
	} else {
		att.HandlerFeedbackFiles = model.JSONRaw("[]")
	}
	if v, ok := args["district_feedback_files"]; ok {
		b, _ := json.Marshal(v)
		att.DistrictFeedbackFiles = model.JSONRaw(b)
	} else {
		att.DistrictFeedbackFiles = model.JSONRaw("[]")
	}
	if v, ok := args["call_recordings"]; ok {
		b, _ := json.Marshal(v)
		att.CallRecordings = model.JSONRaw(b)
	} else {
		att.CallRecordings = model.JSONRaw("[]")
	}
	return dao.UpsertAttachment(att)
}

func GetCategories() ([]map[string]interface{}, error) {
	cats, err := dao.GetAllCategories()
	if err != nil {
		return nil, err
	}
	// build tree
	l1Map := map[string]map[string]interface{}{}
	var result []map[string]interface{}
	for _, c := range cats {
		if _, ok := l1Map[c.Level1]; !ok {
			node := map[string]interface{}{
				"name":     c.Level1,
				"children": []map[string]interface{}{},
			}
			l1Map[c.Level1] = node
			result = append(result, node)
		}
		if c.Level2 != "" {
			l1 := l1Map[c.Level1]
			children := l1["children"].([]map[string]interface{})
			var l2Node map[string]interface{}
			for _, ch := range children {
				if ch["name"] == c.Level2 {
					l2Node = ch
					break
				}
			}
			if l2Node == nil {
				l2Node = map[string]interface{}{
					"name":     c.Level2,
					"children": []map[string]interface{}{},
				}
				children = append(children, l2Node)
				l1["children"] = children
			}
			if c.Level3 != "" {
				l2Children := l2Node["children"].([]map[string]interface{})
				l2Children = append(l2Children, map[string]interface{}{"name": c.Level3})
				l2Node["children"] = l2Children
			}
		}
	}
	if result == nil {
		result = []map[string]interface{}{}
	}
	return result, nil
}

// AnalyzeLetterForDispatch uses LLM to analyze a letter and suggest dispatch target & category
func AnalyzeLetterForDispatch(letterNo string) (map[string]interface{}, error) {
	letter, err := dao.GetLetterByNo(letterNo)
	if err != nil {
		return nil, err
	}

	// Get available categories for context
	categories, _ := GetCategories()
	catJSON, _ := json.Marshal(categories)

	// Get available units for context
	allUnits, _ := dao.GetAllUnits()
	var unitNames []string
	for _, u := range allUnits {
		name := u.Level3
		if name == "" {
			name = u.Level2
		}
		if name == "" {
			name = u.Level1
		}
		unitNames = append(unitNames, name)
	}
	unitsJSON, _ := json.Marshal(unitNames)

	// Build the LLM prompt
	prompt := fmt.Sprintf(`你是一个公安信访信件处理专家。请分析以下信件内容，给出专业的处理建议。

信件编号：%s
来信人：%s
手机号：%s
诉求内容：%s

可用的信件分类（三级分类）：
%s

可用的下发单位：
%s

请以JSON格式返回分析结果，包含以下字段：
1. summary: 信件内容摘要（50字以内）
2. sentiment: 情绪分析（积极/中性/消极/紧急）
3. suggested_category_l1: 建议的一级分类
4. suggested_category_l2: 建议的二级分类
5. suggested_category_l3: 建议的三级分类
6. suggested_unit: 建议下发到的单位
7. urgency: 紧急程度（1-5，5最紧急）
8. reason: 建议理由

只返回JSON，不要其他说明文字。`,
		letter.LetterNo, letter.CitizenName, letter.Phone, letter.Content,
		string(catJSON), string(unitsJSON))

	messages := []LLMMessage{
		{Role: "system", Content: "你是一个专业的公安信访信件分析助手。请严格按JSON格式输出分析结果。"},
		{Role: "user", Content: prompt},
	}

	result, err := Chat(messages)
	if err != nil {
		return nil, fmt.Errorf("LLM analysis failed: %w", err)
	}

	// Try to parse the result as JSON
	var analysis map[string]interface{}
	if err := json.Unmarshal([]byte(result), &analysis); err != nil {
		// Clean up potential markdown code block wrapping
		cleaned := result
		// Remove ```json and ``` if present
		if len(cleaned) > 7 && cleaned[:7] == "```json" {
			cleaned = cleaned[7:]
		}
		if len(cleaned) > 3 && cleaned[len(cleaned)-3:] == "```" {
			cleaned = cleaned[:len(cleaned)-3]
		}
		cleaned = trimWhitespace(cleaned)
		if err2 := json.Unmarshal([]byte(cleaned), &analysis); err2 != nil {
			return nil, fmt.Errorf("failed to parse LLM response as JSON: %s (raw: %s)", err2.Error(), result)
		}
	}

	analysis["letter_no"] = letterNo
	analysis["citizen_name"] = letter.CitizenName
	analysis["content"] = letter.Content
	return analysis, nil
}

// AutoDispatchLetter automatically dispatches a letter using AI analysis
func AutoDispatchLetter(args map[string]interface{}, operator *model.PoliceUser) (map[string]interface{}, error) {
	letterNo, ok := args["letter_no"].(string)
	if !ok || letterNo == "" {
		return nil, errors.New("letter_no required")
	}

	// If target_unit is provided, use it directly
	targetUnit, hasTarget := args["target_unit"].(string)

	if !hasTarget || targetUnit == "" {
		// Use AI to determine the best dispatch target
		analysis, err := AnalyzeLetterForDispatch(letterNo)
		if err != nil {
			return nil, fmt.Errorf("AI analysis failed: %w", err)
		}

		suggestedUnit, _ := analysis["suggested_unit"].(string)
		if suggestedUnit == "" {
			return nil, errors.New("AI未能确定下发目标单位")
		}
		targetUnit = suggestedUnit

		// Dispatch with AI suggestions
		dispatchArgs := map[string]interface{}{
			"letter_no":   letterNo,
			"target_unit": targetUnit,
			"remark":      fmt.Sprintf("AI自动下发：%s", analysis["reason"]),
		}

		if err := DispatchLetter(dispatchArgs, operator); err != nil {
			return nil, err
		}

		analysis["dispatched_to"] = targetUnit
		return analysis, nil
	}

	// Dispatch to specified target
	dispatchArgs := map[string]interface{}{
		"letter_no":   letterNo,
		"target_unit": targetUnit,
		"remark":      "自动下发",
	}
	if v, ok := args["remark"].(string); ok {
		dispatchArgs["remark"] = v
	}

	if err := DispatchLetter(dispatchArgs, operator); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"letter_no":      letterNo,
		"dispatched_to":  targetUnit,
		"auto_dispatched": true,
	}, nil
}

func trimWhitespace(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	if start > 0 || end < len(s) {
		return s[start:end]
	}
	return s
}
