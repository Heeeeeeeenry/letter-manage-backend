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

func GetLetterList(args map[string]interface{}) (map[string]interface{}, error) {
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

func GetLetterDetail(letterNo string) (map[string]interface{}, error) {
	letter, err := dao.GetLetterByNo(letterNo)
	if err != nil {
		return nil, err
	}
	flow, _ := dao.GetFlowByLetterNo(letterNo)
	att, _ := dao.GetAttachmentByLetterNo(letterNo)
	feedbacks, _ := dao.GetFeedbacksByLetterNo(letterNo)
	// history letters by same phone
	var history []model.Letter
	if letter.Phone != "" {
		history, _ = dao.GetLettersByPhone(letter.Phone)
	}
	return map[string]interface{}{
		"letter":    letter,
		"flow":      flow,
		"files":     att,
		"feedbacks": feedbacks,
		"history":   history,
	}, nil
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
	return appendFlowRecord(letterNo, record)
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
	if err := dao.UpdateLetterStatus(letterNo, model.StatusReturned, ""); err != nil {
		return err
	}
	record := map[string]interface{}{
		"action":    "return_letter",
		"status":    model.StatusReturned,
		"remark":    remark,
		"operator":  operator.Name,
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
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

func GetStatistics() (map[string]interface{}, error) {
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
