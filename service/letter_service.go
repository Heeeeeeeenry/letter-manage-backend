package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"letter-manage-backend/dao"
	"letter-manage-backend/model"

	"github.com/xuri/excelize/v2"
)

func GenerateLetterNo() string {
	return fmt.Sprintf("XJ%d", time.Now().UnixNano()/int64(time.Millisecond))
}

func GetLetterList(args map[string]interface{}, user *model.PoliceUser) (map[string]interface{}, error) {
	permLevel := string(user.PermissionLevel)
	// Remove order field from args to prevent SQL injection
	delete(args, "order")
	filter := dao.LetterFilter{}
	if v, ok := args["status"].(string); ok {
		filter.Status = v
	}
	if v, ok := args["category_id"].(float64); ok && v > 0 {
		cid := uint(v)
		filter.CategoryID = &cid
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
	// 查看模式：由后端从 session 中获取用户ID，避免前端传错
	viewMode, _ := args["view_mode"].(string)
	if viewMode == "personal" || permLevel == "OFFICER" {
		// 个人模式：仅过滤处理人，不加单位过滤
		if user.ID > 0 {
			filter.HandlerUserID = &user.ID
		}
	} else {
		// 单位/全部模式：按 handler 所属单位过滤
		switch permLevel {
		case "CITY":
			// 市局：可见所有信件，不过滤
		case "DISTRICT":
			// 区县局：handler_unit_id + current_unit_id 双过滤，覆盖处理中+待下发的信件
			if user.UnitID != nil {
				unitIDs := dao.GetSubordinateUnitIDs(*user.UnitID)
				if len(unitIDs) > 0 {
					filter.HandlerUnitIDs = unitIDs
					filter.AllUnitIDs = unitIDs
				} else {
					uid := *user.UnitID
					filter.HandlerUnitID = &uid
					filter.AllUnitID = &uid
				}
			}
		default:
			// OFFICER 不会进入此分支（permLevel == OFFICER 已走个人模式）
		}
	}

	filter.Page = 1
	filter.PageSize = 20
	if v, ok := args["page"].(float64); ok {
		filter.Page = int(v)
	}
	if v, ok := args["page_size"].(float64); ok {
		filter.PageSize = int(v)
	}
	// 排序参数
	if v, ok := args["order_by"].(string); ok {
		filter.OrderBy = v
	}
	if v, ok := args["order_desc"].(bool); ok {
		filter.OrderDesc = v
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

// normalizeUnitName 将全路径单位名转为短名
// "分局 / 桃城分局 / 民意智感中心" → "民意智感中心"
func normalizeUnitName(name string) string {
	parts := strings.Split(name, " / ")
	return strings.TrimSpace(parts[len(parts)-1])
}

// getSubordinateUnitNames 获取某单位及其下属所有单位的短名称列表
func getSubordinateUnitNames(unitName string) []string {
	return dao.GetSubordinateUnitNames(unitName)
}

func GetDispatchList(unitID *uint, permLevel string, args map[string]interface{}) (map[string]interface{}, error) {
	page := 1
	pageSize := 20
	if v, ok := args["page"].(float64); ok {
		page = int(v)
	}
	if v, ok := args["page_size"].(float64); ok {
		pageSize = int(v)
	}
	letters, total, err := dao.GetDispatchList(unitID, permLevel, page, pageSize)
	if err != nil {
		return nil, err
	}
	// 批量注入 focus_id
	if len(letters) > 0 {
		letterNos := make([]string, len(letters))
		for i, l := range letters {
			letterNos[i] = l.LetterNo
		}
		focusMap, _ := dao.GetFocusIDsByLetterNos(letterNos)
		for i, l := range letters {
			if fid, ok := focusMap[l.LetterNo]; ok {
				letters[i].FocusID = &fid
			}
		}
	}
	return map[string]interface{}{
		"list":      letters,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}, nil
}

func GetProcessingList(unitID *uint, permLevel string, args map[string]interface{}, userID ...uint) (map[string]interface{}, error) {
	page := 1
	pageSize := 20
	if v, ok := args["page"].(float64); ok {
		page = int(v)
	}
	if v, ok := args["page_size"].(float64); ok {
		pageSize = int(v)
	}
	var handlerID uint
	if len(userID) > 0 {
		handlerID = userID[0]
	}
	letters, total, err := dao.GetProcessingList(unitID, permLevel, page, pageSize, handlerID)
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

func GetAuditList(unitID *uint, permLevel string, args map[string]interface{}) (map[string]interface{}, error) {
	page := 1
	pageSize := 20
	if v, ok := args["page"].(float64); ok {
		page = int(v)
	}
	if v, ok := args["page_size"].(float64); ok {
		pageSize = int(v)
	}
	letters, total, err := dao.GetAuditList(unitID, permLevel, page, pageSize)
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
func GetLettersByPhone(phone, permLevel string, unitID *uint) ([]model.Letter, error) {
	letters, err := dao.GetLettersByPhone(phone)
	if err != nil {
		return nil, err
	}
	return filterLettersByPermission(letters, permLevel, unitID), nil
}

// GetLettersByIDCard 获取某身份证的所有信件（带权限过滤）
func GetLettersByIDCard(idCard, permLevel string, unitID *uint) ([]model.Letter, error) {
	letters, err := dao.GetLettersByIDCard(idCard)
	if err != nil {
		return nil, err
	}
	return filterLettersByPermission(letters, permLevel, unitID), nil
}

func GetLetterDetail(letterNo string, permLevel string, userUnitID *uint) (map[string]interface{}, error) {
	letter, err := dao.GetLetterByNo(letterNo)
	if err != nil {
		return nil, err
	}
	// 权限检查：验证用户是否有权访问该信件
	if !canAccessLetter(*letter, permLevel, userUnitID) {
		return nil, errors.New("无权访问该信件")
	}
	flow, _ := dao.GetFlowByLetterNo(letterNo)
	att, _ := dao.GetAttachmentByLetterNo(letterNo)
	feedbacks, _ := dao.GetFeedbacksByLetterNo(letterNo)
	// 注入 focus_id
	if focusIDs, err := dao.GetFocusIDsByLetterNos([]string{letterNo}); err == nil {
		if fid, ok := focusIDs[letterNo]; ok {
			letter.FocusID = &fid
		}
	}
	// history letters by same phone
	var history []model.Letter
	if letter.Phone != "" {
		history, _ = dao.GetLettersByPhone(letter.Phone)
		// 历史信件也要做权限过滤
		history = filterLettersByPermission(history, permLevel, userUnitID)
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
func canAccessLetter(letter model.Letter, permLevel string, userUnitID *uint) bool {
	// 归一化单位名
	normalizedUnit := dao.GetUnitNameByID(userUnitID)
	// unit_id 检查：如果 letter 和 user 都有 unit_id 且相同，直接通过
	if userUnitID != nil && letter.CurrentUnitID != nil && *letter.CurrentUnitID == *userUnitID {
		return true
	}
	switch permLevel {
	case "CITY":
		return true
	case "DISTRICT":
		// DISTRICT 可以访问本单位及下属单位的信件
		// 如果有 unitID，使用 ID 判断
		if userUnitID != nil {
			subIDs := dao.GetSubordinateUnitIDs(*userUnitID)
			if letter.CurrentUnitID != nil {
				for _, sid := range subIDs {
					if sid == *letter.CurrentUnitID {
						return true
					}
				}
			}
		}
		// fallback: 通过 CurrentUnitID 查询单位名称做字符串比较
		letterUnitName := getUnitNameFromID(letter.CurrentUnitID)
		// 全路径转为短名，和 subordinate names 做比较
		letterShortName := dao.NormalizeUnitName(letterUnitName)
		subUnits := dao.GetSubordinateUnitNames(dao.GetUnitNameByID(userUnitID))
		for _, u := range subUnits {
			if u == letterShortName {
				return true
			}
		}
		return letterShortName == normalizedUnit
	default:
		// OFFICER 只能访问本单位的信件
		// 如果有 unitID，直接比较 unit_id
		if userUnitID != nil && letter.CurrentUnitID != nil && *letter.CurrentUnitID == *userUnitID {
			return true
		}
		// fallback: 通过名称比较
		letterUnitName := getUnitNameFromID(letter.CurrentUnitID)
		letterShortName := dao.NormalizeUnitName(letterUnitName)
		return letterShortName == normalizedUnit
	}
}

// getUnitNameFromObj 从 Unit 对象中获取最后一级单位名称
func getUnitNameFromObj(unit *model.Unit) string {
	if unit == nil {
		return ""
	}
	if unit.Level3 != "" {
		return unit.Level3
	}
	if unit.Level2 != "" {
		return unit.Level2
	}
	return unit.Level1
}

// getUnitNameFromID 从 unit ID 获取全路径单位名称（用于 flow record）
func getUnitNameFromID(unitID *uint) string {
	if unitID == nil {
		return ""
	}
	unit, err := dao.GetUnitByID(*unitID)
	if err != nil || unit == nil {
		return ""
	}
	var parts []string
	if unit.Level1 != "" {
		parts = append(parts, unit.Level1)
	}
	if unit.Level2 != "" {
		parts = append(parts, unit.Level2)
	}
	if unit.Level3 != "" {
		parts = append(parts, unit.Level3)
	}
	return strings.Join(parts, " / ")
}

// filterLettersByPermission 根据权限过滤信件列表
func filterLettersByPermission(letters []model.Letter, permLevel string, userUnitID *uint) []model.Letter {
	var filtered []model.Letter
	for _, l := range letters {
		if canAccessLetter(l, permLevel, userUnitID) {
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
		letter.Channel = model.ChannelNameToCode[v]
	}
	if v, ok := args["category_id"].(float64); ok {
		catID := uint(v)
		letter.CategoryID = &catID
	}
	if v, ok := args["content"].(string); ok {
		letter.Content = v
	}
	// 如果传了 current_unit_id，通过 ID 查找单位
	if v, ok := args["current_unit_id"].(float64); ok {
		u := uint(v)
		letter.CurrentUnitID = &u
		unit, err := dao.GetUnitByID(u)
		if err == nil && unit != nil {
			letter.CurrentUnitObj = unit
		}
	}
	letter.CurrentStatus = model.StatusCodePreProcess
	if v, ok := args["received_at"].(string); ok && v != "" {
		t, err := time.ParseInLocation("2006-01-02", v, time.Local)
		if err == nil {
			letter.ReceivedAt = t
		}
	} else {
		letter.ReceivedAt = time.Now()
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
		letter.Channel = model.ChannelNameToCode[v]
	}
	if v, ok := args["category_id"].(float64); ok {
		catID := uint(v)
		letter.CategoryID = &catID
	}
	if v, ok := args["content"].(string); ok {
		letter.Content = v
	}
	// 如果传了 current_unit_id，更新 CurrentUnitID
	if v, ok := args["current_unit_id"].(float64); ok {
		u := uint(v)
		letter.CurrentUnitID = &u
		unit, err := dao.GetUnitByID(u)
		if err == nil && unit != nil {
			letter.CurrentUnitObj = unit
		}
	}
	if v, ok := args["current_status"].(string); ok {
		letter.CurrentStatus = model.StatusNameToCode[v]
	}
	if v, ok := args["received_at"].(string); ok && v != "" {
		t, err := time.ParseInLocation("2006-01-02", v, time.Local)
		if err == nil {
			letter.ReceivedAt = t
		}
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

	if err := dao.UpdateLetterStatus(letterNo, status); err != nil {
		return err
	}

	// append flow record
	flowRecord := map[string]interface{}{
		"status":        status,
		"unit":          unitName,
		"remark":        remark,
		"operator":      operator.Name,
		"operator_id":   operator.PoliceNumber,
		"operator_unit": dao.GetUnitFullNameByID(operator.UnitID),
		"timestamp":     time.Now().Format("2006-01-02 15:04:05"),
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
	canDispatch, err := CheckDispatchPermission(operator, targetUnit, args)
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

	// 查询 targetUnit 对应的 unit ID，并确定下发级别
	var targetUnitID *uint
	var targetHasDispatchLevel bool

	// 优先使用前端传来的 unit_id
	if unitIDF, ok := args["unit_id"].(float64); ok && unitIDF > 0 {
		uid := uint(unitIDF)
		targetUnitID = &uid
		targetHasDispatchLevel = dao.HasDispatchLevelUsersInUnit(uid)
	} else {
		// 没有 unit_id 时通过名称匹配
		allUnits, err := dao.GetAllUnits()
		if err == nil {
			for _, u := range allUnits {
				// 构建全路径用于匹配
				var names []string
				if u.Level1 != "" {
					names = append(names, u.Level1)
				}
				if u.Level2 != "" {
					names = append(names, u.Level2)
				}
				if u.Level3 != "" {
					names = append(names, u.Level3)
				}
				fullPath := strings.Join(names, " / ")
				if fullPath == targetUnit {
					targetUnitID = &u.ID
					targetHasDispatchLevel = dao.HasDispatchLevelUsersInUnit(u.ID)
					break
				}
			}
			// 如果全路径没匹配到，退化为用单级名称匹配
			if targetUnitID == nil {
				for _, u := range allUnits {
					name := u.Level3
					if name == "" {
						name = u.Level2
					}
					if name == "" {
						name = u.Level1
					}
					if name == targetUnit {
						targetUnitID = &u.ID
						targetHasDispatchLevel = dao.HasDispatchLevelUsersInUnit(u.ID)
						break
					}
				}
			}
		}
	}

	// 判断是否指定了处理人
	_, hasHandler := args["handler_user_id"].(float64)

	switch operator.PermissionLevel {
	case model.PermissionCity:
		if hasHandler && targetUnitID != nil {
			// 指定了处理人→直接下发给处理人，跳过区县局下发步骤
			newStatus = model.StatusDispatched
		} else if targetHasDispatchLevel {
			// 未指定处理人，目标单位有区县局用户→待区县局下发
			newStatus = model.StatusPendingDistrictDispatch
		} else {
			// 未指定处理人且目标单位无区县局用户→越级下发
			newStatus = model.StatusCityDirectDispatch
		}
	case model.PermissionDistrict:
		newStatus = model.StatusDispatched
	default:
		return errors.New("无下发权限")
	}

	if err := dao.UpdateLetterStatus(letterNo, newStatus, targetUnitID); err != nil {
		return err
	}
	// 如果指定了处理人，设置 handler_user_id 和 handler_unit_id
	if handlerID, ok := args["handler_user_id"].(float64); ok && handlerID > 0 {
		uid := uint(handlerID)
		updates := map[string]interface{}{
			"handler_user_id": uid,
		}
		// 查找处理人的 unit_id 并同步设置 handler_unit_id
		if handlerUnitID := dao.GetUserUnitID(uid); handlerUnitID != nil {
			updates["handler_unit_id"] = *handlerUnitID
		}
		if err := dao.UpdateLetterFields(letterNo, updates); err != nil {
			return err
		}
	}

	record := map[string]interface{}{
		"action":         "dispatch",
		"status":         newStatus,
		"from_unit":      getUnitNameFromID(letter.CurrentUnitID),
		"to_unit":        targetUnit,
		"remark":         remark,
		"operator":       operator.Name,
		"operator_id":    operator.PoliceNumber,
		"operator_unit":  dao.GetUnitFullNameByID(operator.UnitID),
		"timestamp":      time.Now().Format("2006-01-02 15:04:05"),
	}
	if err := appendFlowRecord(letterNo, record); err != nil {
		return err
	}
	// 保存专项关注绑定关系
	if focusIDF, ok := args["focus_id"].(float64); ok && focusIDF > 0 {
		// 先清除旧绑定，再添加新绑定
		dao.RemoveLetterSpecialFocusesByLetterNo(letterNo)
		dao.AddLetterSpecialFocus(letterNo, uint(focusIDF))
	}
	// 每次下发重新设置 4 个工作日处理倒计时（扣除节假日）
	deadline := GetWorkdayDeadline(time.Now(), 4)
	return dao.UpdateLetterDeadline(letterNo, &deadline)
}

func CheckDispatchPermission(operator *model.PoliceUser, targetUnit string, args ...map[string]interface{}) (bool, error) {
	// 如果传入了 args 且有 unit_id，直接用 ID 检查
	if len(args) > 0 {
		if unitIDF, ok := args[0]["unit_id"].(float64); ok && unitIDF > 0 {
			uid := uint(unitIDF)
			switch operator.PermissionLevel {
			case model.PermissionCity:
				return true, nil
			case model.PermissionDistrict:
				if operator.UnitID != nil {
					if uid == *operator.UnitID {
						return true, nil
					}
					subIDs := dao.GetSubordinateUnitIDs(*operator.UnitID)
					for _, subID := range subIDs {
						if subID == uid {
							return true, nil
						}
					}
				}
				return false, nil
			default:
				// OFFICER: check dispatch_targets + 民意智感中心 fallback
				return checkOfficerDispatch(operator, uid)
			}
		}
	}
	// String-based: resolve targetUnit → unitIDs
	targetIDs := dao.UnitNameToIDs(targetUnit)
	if len(targetIDs) == 0 {
		return false, nil
	}
	switch operator.PermissionLevel {
	case model.PermissionCity:
		return true, nil
	case model.PermissionDistrict:
		opUnit, _ := dao.GetUnitByID(*operator.UnitID)
		if opUnit != nil {
			for _, tid := range targetIDs {
				tu, _ := dao.GetUnitByID(tid)
				if tu != nil && tu.Level1 == opUnit.Level1 && tu.Level2 == opUnit.Level2 {
					return true, nil
				}
			}
		}
		return false, nil
	default:
		// OFFICER
		for _, tid := range targetIDs {
			ok, _ := checkOfficerDispatch(operator, tid)
			if ok {
				return true, nil
			}
		}
		return false, nil
	}
}

// checkOfficerDispatch checks if OFFICER can dispatch to target unit ID
func checkOfficerDispatch(operator *model.PoliceUser, targetUnitID uint) (bool, error) {
	if operator.UnitID == nil {
		return false, nil
	}
	// 1) Check dispatch_targets table
	ok, _ := dao.CheckDispatchPermissionByUnitID(*operator.UnitID, targetUnitID)
	if ok {
		return true, nil
	}
	// 2) 民意智感中心 fallback: same-branch units
	opUnit, _ := dao.GetUnitByID(*operator.UnitID)
	if opUnit != nil && opUnit.Level3 == "民意智感中心" {
		tu, _ := dao.GetUnitByID(targetUnitID)
		if tu != nil && tu.Level1 == opUnit.Level1 && tu.Level2 == opUnit.Level2 {
			return true, nil
		}
	}
	return false, nil
}

func MarkInvalid(args map[string]interface{}, operator *model.PoliceUser) error {
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

	// 找上级单位：OFFICER→DISTRICT, DISTRICT→CITY
	var targetUnitID *uint
	if operator.UnitID != nil {
		targetUnitID = dao.GetParentUnitID(*operator.UnitID)
	}
	// 如果找不到上级，使用当前单位
	if targetUnitID == nil {
		targetUnitID = operator.UnitID
	}

	// 更新状态为待核查，流转到上级单位，清除处理人
	updates := map[string]interface{}{
		"current_status":  model.StatusCodePendingVerification,
		"current_unit_id": targetUnitID,
		"handler_user_id": nil,
		"handler_unit_id": nil,
	}
	if err := dao.UpdateLetterFields(letterNo, updates); err != nil {
		return err
	}

	record := map[string]interface{}{
		"action":         "mark_invalid",
		"status":         model.StatusPendingVerification,
		"remark":         remark,
		"from_unit":      getUnitNameFromID(letter.CurrentUnitID),
		"to_unit":        getUnitNameFromID(targetUnitID),
		"operator":       operator.Name,
		"operator_id":    operator.PoliceNumber,
		"operator_unit":  dao.GetUnitFullNameByID(operator.UnitID),
		"timestamp":      time.Now().Format("2006-01-02 15:04:05"),
	}
	return appendFlowRecord(letterNo, record)
}

func SubmitProcessing(args map[string]interface{}, operator *model.PoliceUser) error {
	letterNo, ok := args["letter_no"].(string)
	if !ok || letterNo == "" {
		return errors.New("letter_no required")
	}
	remark, _ := args["remark"].(string)
	contactFeedback, _ := args["contact_feedback"].(string)

	// 获取流转记录，追溯上次下发的来源单位（上级审核单位）
	letter, err := dao.GetLetterByNo(letterNo)
	if err != nil {
		return err
	}
	if letter == nil {
		return errors.New("letter not found")
	}

	var parentUnit string
	var parentUnitID *uint
	flow, err := dao.GetFlowByLetterNo(letterNo)
	if err != nil {
		return err
	}
	if flow != nil {
		var records []map[string]interface{}
		if err := json.Unmarshal([]byte(flow.FlowRecords), &records); err == nil {
			// 倒序查找最后一次 dispatch 操作，获取来源单位
			for i := len(records) - 1; i >= 0; i-- {
				r := records[i]
				action, _ := r["action"].(string)
				if action == "dispatch" {
					parentUnit, _ = r["from_unit"].(string)
					break
				}
			}
		}
	}

	// 如果 parentUnit 不为空，查询对应的 ID
	if parentUnit != "" {
		allUnits, err := dao.GetAllUnits()
		if err == nil {
			for _, u := range allUnits {
				name := u.Level3
				if name == "" {
					name = u.Level2
				}
				if name == "" {
					name = u.Level1
				}
				if name == parentUnit {
					parentUnitID = &u.ID
					break
				}
			}
		}
	}

	// 更新状态为"待核查"，同时将 current_unit_id 设为上级审核单位
	updates := map[string]interface{}{
		"current_status": model.StatusCodePendingVerification,
	}
	if parentUnitID != nil {
		updates["current_unit_id"] = *parentUnitID
	}
	if err := dao.UpdateLetterFields(letterNo, updates); err != nil {
		return err
	}

	// 不要修改 deadline_at — 保持下发的 4 个工作日倒计时不变

	// 保存联系反馈信息到 feedbacks 表
	if contactFeedback != "" {
		fbInfo, _ := json.Marshal(map[string]interface{}{
			"type":    "contact_feedback",
			"content": contactFeedback,
		})
		fb := &model.Feedback{
			LetterNo:     letterNo,
			FeedbackInfo: model.JSONRaw(fbInfo),
		}
		dao.CreateFeedback(fb)
	}

	record := map[string]interface{}{
		"action":         "submit_processing",
		"status":         model.StatusPendingVerification,
		"remark":         remark,
		"operator":       operator.Name,
		"operator_id":    operator.PoliceNumber,
		"operator_unit":  dao.GetUnitFullNameByID(operator.UnitID),
		"from_unit":      getUnitNameFromObj(letter.CurrentUnitObj),
		"to_unit":        parentUnit,
		"timestamp":      time.Now().Format("2006-01-02 15:04:05"),
	}
	return appendFlowRecord(letterNo, record)
}

func HandleBySelf(args map[string]interface{}, operator *model.PoliceUser) error {
	letterNo, ok := args["letter_no"].(string)
	if !ok || letterNo == "" {
		return errors.New("letter_no required")
	}
	remark, _ := args["remark"].(string)
	if err := dao.UpdateLetterStatus(letterNo, model.StatusProcessing, operator.UnitID); err != nil {
		return err
	}
	// 设置当前用户为处理人
	if err := dao.UpdateLetterFields(letterNo, map[string]interface{}{
		"handler_user_id": operator.ID,
		"handler_unit_id": operator.UnitID,
	}); err != nil {
		return err
	}
	record := map[string]interface{}{
		"action":         "handle_by_self",
		"status":         model.StatusProcessing,
		"unit":           dao.GetUnitFullNameByID(operator.UnitID),
		"remark":         remark,
		"operator":       operator.Name,
		"operator_id":    operator.PoliceNumber,
		"operator_unit":  dao.GetUnitFullNameByID(operator.UnitID),
		"timestamp":      time.Now().Format("2006-01-02 15:04:05"),
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
	var prevUnitID *uint
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
					// 如果 from_unit 为空，尝试从 to_unit 反推
					if prevUnit == "" {
						prevUnit, _ = r["to_unit"].(string)
					}
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
			// 如果没找到 dispatch 记录，尝试从第一条记录获取原始单位
			if prevUnit == "" && len(records) > 0 {
				first := records[0]
				if opUnit, ok := first["操作后单位"].(string); ok {
					prevUnit = opUnit
				}
				if prevStatus == "" {
					if s, ok := first["操作后状态"].(string); ok {
						prevStatus = s
					}
				}
			}
		}
	}
	// 根据退回者身份决定退回状态和退回单位
	// DISTRICT退回→回到区县局下发工作台（状态=待区县局下发，单位=操作人单位）
	// OFFICER退回→回到区县局下发工作台（状态=待区县局下发，单位=上级单位）
	switch operator.PermissionLevel {
	case model.PermissionDistrict:
		if prevStatus == "" {
			prevStatus = model.StatusPendingDistrictDispatch
		}
		// DISTRICT 退回：信件回到自己的单位，而非 CITY 的 from_unit
		prevUnitID = operator.UnitID
	case model.PermissionOfficer:
		if prevStatus == "" || prevStatus == model.StatusDispatched || prevStatus == model.StatusProcessing {
			prevStatus = model.StatusPendingDistrictDispatch
		}
	}
	if prevStatus == "" {
		prevStatus = model.StatusReturned
	}
	if prevUnit == "" {
		prevUnit = getUnitNameFromID(letter.CurrentUnitID)
	}

	// 通过全路径名称匹配上级单位 ID（仅当 prevUnitID 未设置时）
	if prevUnitID == nil {
		allUnits, err := dao.GetAllUnits()
		if err == nil {
			for _, u := range allUnits {
				var parts []string
				if u.Level1 != "" {
					parts = append(parts, u.Level1)
				}
				if u.Level2 != "" {
					parts = append(parts, u.Level2)
				}
				if u.Level3 != "" {
					parts = append(parts, u.Level3)
				}
				fullPath := strings.Join(parts, " / ")
				if fullPath == prevUnit {
					prevUnitID = &u.ID
					break
				}
			}
			// 如果全路径没匹配到，退化用短名匹配
			if prevUnitID == nil {
				for _, u := range allUnits {
					shortName := u.Level3
					if shortName == "" {
						shortName = u.Level2
					}
					if shortName == "" {
						shortName = u.Level1
					}
					if shortName == prevUnit {
						prevUnitID = &u.ID
						break
					}
				}
			}
		}
	}
	// 更新信件为上一个状态
	statusCode := model.StatusCodeReturned
	if code, ok := model.StatusNameToCode[prevStatus]; ok {
		statusCode = code
	}
	updates := map[string]interface{}{
		"current_status":   statusCode,
		"current_operator": prevOperator,
	}
	if prevUnitID != nil {
		updates["current_unit_id"] = *prevUnitID
	}
	// 退回后清除处理人及处理单位
	updates["handler_user_id"] = nil
	updates["handler_unit_id"] = nil
	if err := dao.UpdateLetterFields(letterNo, updates); err != nil {
		return err
	}

	// 清除 deadline（等待下次下发重新计时）
	if err := dao.UpdateLetterDeadline(letterNo, nil); err != nil {
		return err
	}

	// 追加退回记录（保留完整历史）
	record := map[string]interface{}{
		"action":         "return_letter",
		"status":         prevStatus,
		"from_unit":      getUnitNameFromID(letter.CurrentUnitID),
		"to_unit":        prevUnit,
		"from_operator":  letter.CurrentOperator,
		"to_operator":    prevOperator,
		"remark":         remark,
		"operator":       operator.Name,
		"operator_id":    operator.PoliceNumber,
		"operator_unit":  dao.GetUnitFullNameByID(operator.UnitID),
		"timestamp":      time.Now().Format("2006-01-02 15:04:05"),
	}
	return appendFlowRecord(letterNo, record)
}

func AuditApprove(args map[string]interface{}, operator *model.PoliceUser) error {
	letterNo, ok := args["letter_no"].(string)
	if !ok || letterNo == "" {
		return errors.New("letter_no required")
	}
	remark, _ := args["remark"].(string)

	var newStatus string
	switch operator.PermissionLevel {
	case model.PermissionDistrict:
		// 分县局审核通过 → 上报市局审批
		newStatus = model.StatusPendingCityAudit
	case model.PermissionCity:
		// 市局审核通过 → 办结
		newStatus = model.StatusDone
	default:
		return errors.New("无审核权限")
	}

	if err := dao.UpdateLetterStatus(letterNo, newStatus); err != nil {
		return err
	}
	record := map[string]interface{}{
		"action":    "audit_approve",
		"status":    newStatus,
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

	// 核查不通过：追溯流转记录找到原始处理单位（from_unit 或处理民警单位）
	letter, err := dao.GetLetterByNo(letterNo)
	if err != nil {
		return err
	}
	if letter == nil {
		return errors.New("letter not found")
	}

	// 从流转记录中追溯上次 submit_processing 操作的 from_unit
	var processingUnit string
	var processingUnitID *uint
	flow, err := dao.GetFlowByLetterNo(letterNo)
	if err != nil {
		return err
	}
	if flow != nil {
		var records []map[string]interface{}
		if err := json.Unmarshal([]byte(flow.FlowRecords), &records); err == nil {
			// 倒序查找最后一次 submit_processing 操作，获取 from_unit（原始处理单位）
			for i := len(records) - 1; i >= 0; i-- {
				r := records[i]
				action, _ := r["action"].(string)
				if action == "submit_processing" {
					processingUnit, _ = r["from_unit"].(string)
					// 如果 from_unit 为空，尝试从 operator_unit 获取
					if processingUnit == "" {
						if opUnit, ok := r["operator_unit"].(string); ok {
							processingUnit = opUnit
						}
					}
					break
				}
			}
			// 如果仍未找到，从最近的 handle_by_self 记录获取处理单位
			if processingUnit == "" {
				for i := len(records) - 1; i >= 0; i-- {
					r := records[i]
					action, _ := r["action"].(string)
					if action == "handle_by_self" || action == "dispatch" {
						if unit, ok := r["unit"].(string); ok && unit != "" {
							processingUnit = unit
						} else if opUnit, ok := r["operator_unit"].(string); ok && opUnit != "" {
							processingUnit = opUnit
						} else if toUnit, ok := r["to_unit"].(string); ok && toUnit != "" {
							processingUnit = toUnit
						}
						break
					}
				}
			}
		}
	}

	// 如果追溯不到，使用当前状态前的单位
	if processingUnit == "" {
		processingUnit = getUnitNameFromObj(letter.CurrentUnitObj)
	}

	// 退回处理单位，状态恢复为"处理中"
	updates := map[string]interface{}{
		"current_status": model.StatusCodeProcessing,
	}
	// 如果 processingUnit 不为空，查询对应的 ID
	if processingUnit != "" && processingUnitID == nil {
		allUnits, err := dao.GetAllUnits()
		if err == nil {
			for _, u := range allUnits {
				name := u.Level3
				if name == "" {
					name = u.Level2
				}
				if name == "" {
					name = u.Level1
				}
				if name == processingUnit {
					processingUnitID = &u.ID
					break
				}
			}
		}
	}
	if processingUnitID != nil {
		updates["current_unit_id"] = *processingUnitID
	}
	if err := dao.UpdateLetterFields(letterNo, updates); err != nil {
		return err
	}

	// 重新设置 deadline 为 4 个工作日（从退回时重新计时）
	deadline := GetWorkdayDeadline(time.Now(), 4)
	if err := dao.UpdateLetterDeadline(letterNo, &deadline); err != nil {
		return err
	}

	record := map[string]interface{}{
		"action":         "audit_reject",
		"status":         model.StatusProcessing,
		"remark":         remark,
		"operator":       operator.Name,
		"operator_id":    operator.PoliceNumber,
		"operator_unit":  dao.GetUnitFullNameByID(operator.UnitID),
		"from_unit":      getUnitNameFromObj(letter.CurrentUnitObj),
		"to_unit":        processingUnit,
		"timestamp":      time.Now().Format("2006-01-02 15:04:05"),
	}
	return appendFlowRecord(letterNo, record)
}

func GetStatistics(permLevel string, period string, unitID *uint, handlerUserID uint, viewMode string) (map[string]interface{}, error) {
	// 根据 period 计算滚动时间窗口
	var startTime, endTime *time.Time
	now := time.Now()
	switch period {
	case "day":
		t := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		startTime = &t
	case "week":
		t := now.AddDate(0, 0, -6)
		t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, now.Location())
		startTime = &t
	case "month":
		t := now.AddDate(0, 0, -29)
		t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, now.Location())
		startTime = &t
	case "year":
		t := now.AddDate(0, 0, -364)
		t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, now.Location())
		startTime = &t
	}
	if startTime != nil {
		t := now
		endTime = &t
	}

	// 根据权限计算可访问的单位列表（用于 handler_unit_id 过滤）
	var unitIDs []uint

	// 如果传了 unitID，优先使用 unitID 路径
	hasUnitID := unitID != nil

	switch permLevel {
	case "CITY":
		// 市局：可见所有数据，不过滤
	case "DISTRICT":
		if hasUnitID {
			unitIDs = dao.GetSubordinateUnitIDs(*unitID)
			if len(unitIDs) == 0 {
				unitIDs = []uint{*unitID}
			}
		}
	default:
		// OFFICER：仅可见本单位处理人的数据
		if hasUnitID {
			unitIDs = []uint{*unitID}
		}
	}

	var statusStats []dao.StatusCount
	var channelStats []dao.ChannelCount
	var catStats []dao.CategoryCount
	var err error

	// 确定是否个人模式
	isPersonal := viewMode == "personal" || permLevel == "OFFICER"

	// 如果有 unitIDs，使用 ByUnitIDs 函数（handler_unit_id 过滤）
	// 无 unitIDs 时（如 CITY），不按 handler 单位过滤
	if isPersonal && handlerUserID > 0 {
		statusStats, err = dao.GetLetterStatusStatsByUnitIDs(startTime, endTime, unitIDs, handlerUserID)
	} else {
		statusStats, err = dao.GetLetterStatusStatsByUnitIDs(startTime, endTime, unitIDs)
	}
	if err != nil {
		return nil, err
	}
	if isPersonal && handlerUserID > 0 {
		channelStats, err = dao.GetLetterChannelStatsByUnitIDs(unitIDs, handlerUserID)
	} else {
		channelStats, err = dao.GetLetterChannelStatsByUnitIDs(unitIDs)
	}
	if err != nil {
		return nil, err
	}
	// 趋势数据：根据 period 选择粒度
	trendGranularity := periodToGranularity(period)
	var trendPoints []dao.TrendPoint
	if isPersonal && handlerUserID > 0 {
		trendPoints, err = dao.GetLetterTrend(unitIDs, handlerUserID, trendGranularity, startTime, endTime)
	} else {
		trendPoints, err = dao.GetLetterTrend(unitIDs, 0, trendGranularity, startTime, endTime)
	}
	if err != nil {
		return nil, err
	}
	if isPersonal && handlerUserID > 0 {
		catStats, err = dao.GetLetterCategoryStatsByUnitIDs(startTime, endTime, unitIDs, handlerUserID)
	} else {
		catStats, err = dao.GetLetterCategoryStatsByUnitIDs(startTime, endTime, unitIDs)
	}
	if err != nil {
		return nil, err
	}

	// 构建 summary 统计
	var total int64
	var preprocessCount, processingCount, doneCount, districtAuditCount int64
	for _, s := range statusStats {
		total += s.Count
		switch s.Status {
		case model.StatusCodePreProcess:
			preprocessCount = s.Count
		case model.StatusCodePendingVerification:
			// 待核查：分局审核层面，市局不应计入
			if permLevel == "CITY" {
				processingCount += s.Count
			} else {
				districtAuditCount += s.Count
			}
		case model.StatusCodePendingDistrictAudit:
			districtAuditCount += s.Count
		case model.StatusCodePendingCityAudit:
			// 待市局审核：只有市局可见
			if permLevel == "CITY" {
				districtAuditCount += s.Count
			} else {
				processingCount += s.Count
			}
		case model.StatusCodeDispatched, model.StatusCodeProcessing, model.StatusCodeCityDirectDispatch:
			processingCount += s.Count
		case model.StatusCodeDone:
			doneCount = s.Count
		case model.StatusCodeInvalid:
			// skip
		default:
			processingCount += s.Count
		}
	}

	// 构建状态分布数组
	statusDistribution := []map[string]interface{}{}
	for _, s := range statusStats {
		if s.Status == model.StatusCodeInvalid {
			statusDistribution = append(statusDistribution, map[string]interface{}{
				"name":  "已无效",
				"value": s.Count,
			})
		} else {
			name := model.StatusCodeToName[s.Status]
			if name == "" {
				name = fmt.Sprintf("%d", s.Status)
			}
			statusDistribution = append(statusDistribution, map[string]interface{}{
				"name":  name,
				"value": s.Count,
			})
		}
	}

	// 构建趋势数据
	trendDates := []string{}
	trendValues := []int64{}
	for _, tp := range trendPoints {
		trendDates = append(trendDates, tp.Label)
		trendValues = append(trendValues, tp.Count)
	}

	// 构建分类统计
	categories := []string{}
	catValues := []int64{}
	for _, c := range catStats {
		categories = append(categories, c.Category)
		catValues = append(catValues, c.Count)
	}

	// 构建来源分布
	sourceDistribution := []map[string]interface{}{}
	for _, ch := range channelStats {
		chName := model.ChannelToName[ch.Channel]
		if chName == "" {
			chName = fmt.Sprintf("%d", ch.Channel)
		}
		sourceDistribution = append(sourceDistribution, map[string]interface{}{
			"name":  chName,
			"value": ch.Count,
		})
	}

	return map[string]interface{}{
		"信件总量":   total,
		"预处理":    preprocessCount,
		"处理中":    processingCount,
		"已完成":    doneCount,
		"待分县局/支队审核": districtAuditCount,
		"状态分布":   statusDistribution,
		"趋势":     map[string]interface{}{"dates": trendDates, "values": trendValues},
		"分类统计":   map[string]interface{}{"categories": categories, "values": catValues},
		"来源分布":   sourceDistribution,
		// 环比对比（暂不计算，前端显示-）
		"comparison": nil,
		// 保留原始数据以备后用
		"status_stats":  statusStats,
		"channel_stats": channelStats,
	}, nil
}

func GetAttachments(letterNo string) (*model.LetterAttachment, error) {
	att, err := dao.GetAttachmentByLetterNo(letterNo)
	if err != nil {
		return nil, err
	}
	if att == nil {
		now := time.Now()
		att = &model.LetterAttachment{
			LetterNo:              letterNo,
			CityDispatchFiles:     model.JSONRaw("[]"),
			DistrictDispatchFiles: model.JSONRaw("[]"),
			HandlerFeedbackFiles:  model.JSONRaw("[]"),
			DistrictFeedbackFiles: model.JSONRaw("[]"),
			CallRecordings:        model.JSONRaw("[]"),
			CreatedAt:             now,
			UpdatedAt:             now,
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

// AppendAttachment 向 letter_attachments 的指定字段追加一条文件记录
func AppendAttachment(letterNo, fileType, url, fileName string) error {
	att, err := dao.GetAttachmentByLetterNo(letterNo)
	if err != nil {
		return err
	}
	if att == nil {
		now := time.Now()
		att = &model.LetterAttachment{
			LetterNo:              letterNo,
			CityDispatchFiles:     model.JSONRaw("[]"),
			DistrictDispatchFiles: model.JSONRaw("[]"),
			HandlerFeedbackFiles:  model.JSONRaw("[]"),
			DistrictFeedbackFiles: model.JSONRaw("[]"),
			CallRecordings:        model.JSONRaw("[]"),
			CreatedAt:             now,
			UpdatedAt:             now,
		}
	}

	entry := map[string]string{"url": url, "name": fileName}

	var currentJSON model.JSONRaw
	switch fileType {
	case "city_dispatch_files":
		currentJSON = att.CityDispatchFiles
	case "district_dispatch_files":
		currentJSON = att.DistrictDispatchFiles
	case "handler_feedback_files":
		currentJSON = att.HandlerFeedbackFiles
	case "district_feedback_files":
		currentJSON = att.DistrictFeedbackFiles
	default:
		currentJSON = att.CallRecordings
	}

	var list []map[string]string
	if len(currentJSON) > 0 {
		json.Unmarshal(currentJSON, &list)
	}
	list = append(list, entry)
	b, _ := json.Marshal(list)

	switch fileType {
	case "city_dispatch_files":
		att.CityDispatchFiles = model.JSONRaw(b)
	case "district_dispatch_files":
		att.DistrictDispatchFiles = model.JSONRaw(b)
	case "handler_feedback_files":
		att.HandlerFeedbackFiles = model.JSONRaw(b)
	case "district_feedback_files":
		att.DistrictFeedbackFiles = model.JSONRaw(b)
	default:
		att.CallRecordings = model.JSONRaw(b)
	}

	return dao.UpsertAttachment(att)
}

func GetCategories() ([]map[string]interface{}, error) {
	cats, err := dao.GetAllCategories()
	if err != nil {
		return nil, err
	}
	// build tree (with IDs for frontend to use in filtering)
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
				l2Children = append(l2Children, map[string]interface{}{
					"name": c.Level3,
					"id":   c.ID,
				})
				l2Node["children"] = l2Children
			} else {
				// Level2 has no Level3 — store the category ID on Level2 node itself
				// Only set if there's truly no Level3 (the category entry is level2-only)
				// But since each row in categories has its own ID, we need to track
				// which ID corresponds to the (level1,level2) pair without level3
				if _, hasID := l2Node["id"]; !hasID {
					l2Node["id"] = c.ID
				}
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

// isStationLevelUnit 检查目标单位是否为基层科室所队（Level3）
// 用于判断市局下发是否为越级下发
func isStationLevelUnit(targetUnit string) bool {
	units, err := dao.GetAllUnits()
	if err != nil {
		return false
	}
	for _, u := range units {
		if u.Level3 == targetUnit {
			return true
		}
	}
	return false
}

func periodToGranularity(period string) string {
	switch period {
	case "day":
		return "hour"
	case "week":
		return "day"
	case "month":
		return "day"
	case "year":
		return "month"
	default:
		return "month"
	}
}

// ExportLetters 导出信件为 Excel，返回文件路径
func ExportLetters(permLevel string, unitID *uint, handlerUserID uint, args map[string]interface{}) (string, error) {
	filter := dao.LetterFilter{}
	now := time.Now()

	// 时间范围：优先用 start_time/end_time，其次用 period
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
	// 如果没有精确时间范围，使用 period 兜底
	if filter.StartTime == nil {
		now := time.Now()
		period, _ := args["period"].(string)
		switch period {
		case "day":
			t := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			filter.StartTime = &t
		case "week":
			t := now.AddDate(0, 0, -6)
			t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, now.Location())
			filter.StartTime = &t
		case "month":
			t := now.AddDate(0, 0, -29)
			t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, now.Location())
			filter.StartTime = &t
		case "year":
			t := now.AddDate(0, 0, -364)
			t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, now.Location())
			filter.StartTime = &t
		}
		if filter.StartTime != nil {
			t := now
			filter.EndTime = &t
		}
	}

	// 筛选字段
	if v, ok := args["status"].(string); ok {
		filter.Status = v
	}
	if v, ok := args["category_id"].(float64); ok && v > 0 {
		cid := uint(v)
		filter.CategoryID = &cid
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

	// 权限过滤
	viewMode, _ := args["view_mode"].(string)
	switch permLevel {
	case "CITY":
	case "DISTRICT":
		if unitID != nil {
			unitIDs := dao.GetSubordinateUnitIDs(*unitID)
			if len(unitIDs) > 0 {
				filter.AllUnitIDs = unitIDs
			} else {
				filter.AllUnitID = unitID
			}
		}
	default:
		if unitID != nil {
			filter.AllUnitID = unitID
		}
	}
	if viewMode == "personal" || permLevel == "OFFICER" {
		if handlerUserID > 0 {
			filter.HandlerUserID = &handlerUserID
		}
	}

	letters, err := dao.GetLettersForExport(filter)
	if err != nil {
		return "", err
	}

	f := excelize.NewFile()
	sheet := "信件数据"
	f.SetSheetName("Sheet1", sheet)

	// 标题行样式
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 11, Family: "微软雅黑"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#4472C4"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})

	cellStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})

	// 表头
	headers := []string{
		"序号", "信件编号", "信件状态", "来信时间", "来信渠道",
		"群众姓名", "手机号码", "信件类别", "信件细类", "简要诉求",
		"分县局", "主办单位", "是否逾期", "是否退回",
	}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
		f.SetCellStyle(sheet, cell, cell, headerStyle)
	}

	// 数据行
	for i, l := range letters {
		row := i + 2
		f.SetCellValue(sheet, cellName(1, row), i+1)
		f.SetCellValue(sheet, cellName(2, row), l.LetterNo)
		f.SetCellValue(sheet, cellName(3, row), model.StatusCodeToName[l.CurrentStatus])
		f.SetCellValue(sheet, cellName(4, row), l.ReceivedAt.Format("2006-01-02 15:04:05"))
		f.SetCellValue(sheet, cellName(5, row), model.ChannelToName[l.Channel])
		f.SetCellValue(sheet, cellName(6, row), l.CitizenName)
		f.SetCellValue(sheet, cellName(7, row), l.Phone)
		// 分类
		catStr := ""
		if l.Category != nil {
			catStr = l.Category.Level1
		}
		f.SetCellValue(sheet, cellName(8, row), catStr)
		catDetail := ""
		if l.Category != nil {
			parts := []string{}
			if l.Category.Level2 != "" {
				parts = append(parts, l.Category.Level2)
			}
			if l.Category.Level3 != "" {
				parts = append(parts, l.Category.Level3)
			}
			catDetail = strings.Join(parts, " / ")
		}
		f.SetCellValue(sheet, cellName(9, row), catDetail)
		f.SetCellValue(sheet, cellName(10, row), l.Content)
		f.SetCellValue(sheet, cellName(11, row), getUnitFullName(l.CurrentUnitID))
		f.SetCellValue(sheet, cellName(12, row), getUnitFullName(l.HandlerUnitID))
		// 是否逾期
		overdue := "否"
		if l.DeadlineAt != nil && now.After(*l.DeadlineAt) && l.CurrentStatus != model.StatusCodeDone {
			overdue = "是"
		}
		f.SetCellValue(sheet, cellName(13, row), overdue)
		// 是否退回
		returned := "否"
		if l.CurrentStatus == model.StatusCodeReturned {
			returned = "是"
		}
		f.SetCellValue(sheet, cellName(14, row), returned)

		// 应用样式
		for c := 1; c <= len(headers); c++ {
			cell, _ := excelize.CoordinatesToCellName(c, row)
			f.SetCellStyle(sheet, cell, cell, cellStyle)
		}
	}

	// 列宽
	f.SetColWidth(sheet, "A", "A", 6)
	f.SetColWidth(sheet, "B", "B", 22)
	f.SetColWidth(sheet, "C", "C", 12)
	f.SetColWidth(sheet, "D", "D", 18)
	f.SetColWidth(sheet, "E", "E", 10)
	f.SetColWidth(sheet, "F", "F", 10)
	f.SetColWidth(sheet, "G", "G", 13)
	f.SetColWidth(sheet, "H", "H", 14)
	f.SetColWidth(sheet, "I", "I", 20)
	f.SetColWidth(sheet, "J", "J", 40)
	f.SetColWidth(sheet, "K", "K", 16)
	f.SetColWidth(sheet, "L", "L", 16)
	f.SetColWidth(sheet, "M", "M", 8)
	f.SetColWidth(sheet, "N", "N", 8)

	// 保存
	tmpDir := os.TempDir()
	filename := fmt.Sprintf("信件导出_%s.xlsx", now.Format("20060102_150405"))
	filePath := filepath.Join(tmpDir, filename)
	if err := f.SaveAs(filePath); err != nil {
		return "", err
	}
	return filePath, nil
}

func cellName(col, row int) string {
	name, _ := excelize.CoordinatesToCellName(col, row)
	return name
}

func getUnitFullName(unitID *uint) string {
	if unitID == nil {
		return ""
	}
	return dao.GetUnitFullNameByID(unitID)
}

// SetLetterSpecialFocus 设置信件的专项关注（独立于下发动作）
func SetLetterSpecialFocus(args map[string]interface{}) error {
	letterNo, ok := args["letter_no"].(string)
	if !ok || letterNo == "" {
		return errors.New("letter_no required")
	}
	focusIDF, ok := args["focus_id"].(float64)
	if !ok {
		return errors.New("focus_id required")
	}
	// 先清除旧绑定，再添加新绑定
	dao.RemoveLetterSpecialFocusesByLetterNo(letterNo)
	return dao.AddLetterSpecialFocus(letterNo, uint(focusIDF))
}

// GetLetterSpecialFocus 获取信件的专项关注
func GetLetterSpecialFocus(letterNo string) (uint, string, error) {
	ids, err := dao.GetFocusIDsByLetterNo(letterNo)
	if err != nil {
		return 0, "", err
	}
	if len(ids) == 0 {
		return 0, "", nil
	}
	focusID := ids[len(ids)-1] // 取最近一条
	sf, err := dao.GetSpecialFocusByID(focusID)
	if err != nil {
		return focusID, "", nil
	}
	if sf == nil {
		return focusID, "", nil
	}
	return sf.ID, sf.Name, nil
}
