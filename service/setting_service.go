package service

import (
	"encoding/json"
	"errors"
	"log"
	"strconv"

	"letter-manage-backend/dao"
	"letter-manage-backend/model"
)

func marshalJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func marshalJSONRaw(raw model.JSONRaw, dest interface{}) error {
	if len(raw) == 0 {
		return nil
	}
	return json.Unmarshal([]byte(raw), dest)
}

// Category

func GetCategoryList() ([]model.Category, error) {
	return dao.GetAllCategories()
}

func CreateCategory(args map[string]interface{}) error {
	cat := &model.Category{}
	if v, ok := args["level1"].(string); ok {
		cat.Level1 = v
	}
	if v, ok := args["level2"].(string); ok {
		cat.Level2 = v
	}
	if v, ok := args["level3"].(string); ok {
		cat.Level3 = v
	}
	if cat.Level1 == "" {
		return errors.New("level1 required")
	}
	return dao.CreateCategory(cat)
}

func UpdateCategory(args map[string]interface{}) error {
	idF, ok := args["id"].(float64)
	if !ok {
		return errors.New("id required")
	}
	cat, err := dao.GetCategoryByID(uint(idF))
	if err != nil {
		return err
	}
	if v, ok := args["level1"].(string); ok {
		cat.Level1 = v
	}
	if v, ok := args["level2"].(string); ok {
		cat.Level2 = v
	}
	if v, ok := args["level3"].(string); ok {
		cat.Level3 = v
	}
	return dao.UpdateCategory(cat)
}

func DeleteCategory(args map[string]interface{}) error {
	idF, ok := args["id"].(float64)
	if !ok {
		return errors.New("id required")
	}
	return dao.DeleteCategory(uint(idF))
}

// Units

func GetUnitList(args map[string]interface{}) (map[string]interface{}, error) {
	page := 1
	pageSize := 100
	if v, ok := args["page"].(float64); ok {
		page = int(v)
	}
	if v, ok := args["page_size"].(float64); ok {
		pageSize = int(v)
	}
	units, total, err := dao.GetUnitList(page, pageSize)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"list":      units,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}, nil
}

func GetAllUnits() ([]model.Unit, error) {
	return dao.GetAllUnits()
}

// GetUnitsWithFilter 获取单位列表，支持分页和筛选
func GetUnitsWithFilter(args map[string]interface{}) (map[string]interface{}, error) {
	// 调试日志：打印接收到的参数
	log.Printf("[GetUnitsWithFilter] args: %+v", args)
	
	// 解析分页参数
	page := 1
	if v, ok := args["page"]; ok {
		switch val := v.(type) {
		case float64:
			page = int(val)
		case int:
			page = val
		case string:
			if i, err := strconv.Atoi(val); err == nil {
				page = i
			}
		}
	}
	pageSize := 20
	if v, ok := args["page_size"]; ok {
		switch val := v.(type) {
		case float64:
			pageSize = int(val)
		case int:
			pageSize = val
		case string:
			if i, err := strconv.Atoi(val); err == nil {
				pageSize = i
			}
		}
	}
	
	log.Printf("[GetUnitsWithFilter] parsed: page=%d, pageSize=%d", page, pageSize)

	// 解析筛选参数
	searchKeyword := ""
	if v, ok := args["search_keyword"].(string); ok {
		searchKeyword = v
	}
	filterLevel1 := ""
	if v, ok := args["filter_level1"].(string); ok {
		filterLevel1 = v
	}
	filterLevel2 := ""
	if v, ok := args["filter_level2"].(string); ok {
		filterLevel2 = v
	}

	// 调用 DAO
	units, total, err := dao.GetUnitsWithFilter(page, pageSize, searchKeyword, filterLevel1, filterLevel2)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"list":      units,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}, nil
}

func CreateUnit(args map[string]interface{}) error {
	unit := &model.Unit{}
	if v, ok := args["level1"].(string); ok {
		unit.Level1 = v
	}
	if v, ok := args["level2"].(string); ok {
		unit.Level2 = v
	}
	if v, ok := args["level3"].(string); ok {
		unit.Level3 = v
	}
	if v, ok := args["system_code"].(string); ok {
		unit.SystemCode = v
	}
	if unit.SystemCode == "" {
		return errors.New("system_code required")
	}
	return dao.CreateUnit(unit)
}

func UpdateUnit(args map[string]interface{}) error {
	idF, ok := args["id"].(float64)
	if !ok {
		return errors.New("id required")
	}
	unit, err := dao.GetUnitByID(uint(idF))
	if err != nil {
		return err
	}
	if v, ok := args["level1"].(string); ok {
		unit.Level1 = v
	}
	if v, ok := args["level2"].(string); ok {
		unit.Level2 = v
	}
	if v, ok := args["level3"].(string); ok {
		unit.Level3 = v
	}
	if v, ok := args["system_code"].(string); ok {
		unit.SystemCode = v
	}
	return dao.UpdateUnit(unit)
}

func DeleteUnit(args map[string]interface{}) error {
	idF, ok := args["id"].(float64)
	if !ok {
		return errors.New("id required")
	}
	return dao.DeleteUnit(uint(idF))
}

func LevelRank(level string) int {
	switch level {
	case "CITY":
		return 3
	case "DISTRICT":
		return 2
	case "OFFICER":
		return 1
	default:
		return 0
	}
}

// Users

func GetUserList(args map[string]interface{}, currentUnitName string, permLevel string, currentIsAdmin bool, currentUnitID ...*uint) (map[string]interface{}, error) {
	page := 1
	pageSize := 20
	if v, ok := args["page"].(float64); ok {
		page = int(v)
	}
	if v, ok := args["page_size"].(float64); ok {
		pageSize = int(v)
	}
	// 权限数据隔离：根据用户级别限制可见的用户范围
	var unitFilter string
	var unitIDArg *uint
	if len(currentUnitID) > 0 {
		unitIDArg = currentUnitID[0]
	}
	switch permLevel {
	case "CITY":
		// 市局：可见所有用户
	case "DISTRICT":
		// 区县局：可见本单位及下属单位的用户
		subUnits := getSubordinateUnitNames(currentUnitName)
		if len(subUnits) > 0 {
			// 用户管理按单位过滤，可以传空字符串表示不过滤，但这里我们需要处理多单位
			// 简化处理：将单位名数组传给 DAO
		}
		unitFilter = currentUnitName
	default:
		unitFilter = currentUnitName
	}
	users, total, err := dao.GetUserList(page, pageSize, unitFilter, permLevel, currentIsAdmin, unitIDArg)
	if err != nil {
		return nil, err
	}
	// 映射字段为前端期望的键名
	mappedUsers := make([]map[string]interface{}, len(users))
	for i, user := range users {
		// 状态映射：is_active 布尔值 -> 显示文本
		statusText := "已禁用"
		if user.IsActive {
			statusText = "已激活"
		}
		// 手机号可见规则
		// - Admin看同级: 可见（手机号只读）
		// - Admin看下级: 可见（手机号可编辑，非管理员用户）
		// - 非admin看同级任何用户: 不可见
		// - 非admin看下级: 可见
		viewerRank := LevelRank(permLevel)
		targetRank := LevelRank(string(user.PermissionLevel))
		showPhone := viewerRank != targetRank || (viewerRank == targetRank && currentIsAdmin)
		phone := user.Phone
		if !showPhone {
			phone = ""
		}
		// 手机号是否可编辑
		// - 不同级别（查看者是上级）：可编辑（非管理员用户）
		// - 同级别且查看者是管理员：不可编辑（只读）
		// - 其他情况：不可编辑
		phoneEditable := false
		if showPhone {
			if viewerRank != targetRank {
				// 上级看下级：可编辑非管理员
				phoneEditable = currentIsAdmin || !user.IsAdmin
			} else {
				// 同级别：管理员可见但不可编辑
				phoneEditable = false
			}
		}
		mappedUsers[i] = map[string]interface{}{
			"id":               user.ID,
			"name":             user.Name,
			"police_number":    user.PoliceNumber,
			"unit_id":          user.UnitID,
			"role":             string(user.PermissionLevel),
			"permission_level": string(user.PermissionLevel),
			"status":           statusText,
			"is_active":        user.IsActive,
			"is_admin":         user.IsAdmin,
			"nickname":         user.Nickname,
			"phone":            phone,
			"phone_editable":   phoneEditable,
			"created_at":       user.CreatedAt,
			"last_login":       user.LastLogin,
		}
	}
	return map[string]interface{}{
		"list":      mappedUsers,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}, nil
}

func CreateUser(args map[string]interface{}) error {
	user := &model.PoliceUser{}
	if v, ok := args["name"].(string); ok {
		user.Name = v
	}
	if v, ok := args["nickname"].(string); ok {
		user.Nickname = v
	}
	if v, ok := args["police_number"].(string); ok {
		user.PoliceNumber = v
	}
	if v, ok := args["phone"].(string); ok {
		user.Phone = v
	}
	if v, ok := args["unit_name"].(string); ok {
		user.UnitName = dao.NormalizeUnitName(v)
	}
	// 如果传了 unit_id，通过 ID 查找单位
	if v, ok := args["unit_id"].(float64); ok {
		unit, err := dao.GetUnitByID(uint(v))
		if err == nil && unit != nil {
			u := uint(v)
			user.UnitID = &u
			if user.UnitName == "" {
				user.UnitName = unit.Level3
				if user.UnitName == "" {
					user.UnitName = unit.Level2
				}
				if user.UnitName == "" {
					user.UnitName = unit.Level1
				}
			}
		}
	}
	if v, ok := args["permission_level"].(string); ok {
		user.PermissionLevel = model.PermissionLevel(v)
	}
	password, ok := args["password"].(string)
	if !ok || password == "" {
		return errors.New("password required")
	}
	user.PasswordHash = HashPassword(password)
	user.IsActive = true
	if v, ok := args["is_admin"].(bool); ok {
		user.IsAdmin = v
	}
	if user.PoliceNumber == "" {
		return errors.New("police_number required")
	}
	if user.Name == "" {
		return errors.New("name required")
	}
	return dao.CreateUser(user)
}

func UpdateUser(args map[string]interface{}) error {
	idF, ok := args["id"].(float64)
	if !ok {
		return errors.New("id required")
	}
	user, err := dao.GetUserByID(uint(idF))
	if err != nil {
		return err
	}
	if v, ok := args["name"].(string); ok {
		user.Name = v
	}
	if v, ok := args["nickname"].(string); ok {
		user.Nickname = v
	}
	if v, ok := args["phone"].(string); ok {
		user.Phone = v
	}
	if v, ok := args["unit_name"].(string); ok {
		user.UnitName = dao.NormalizeUnitName(v)
	}
	// 如果传了 unit_id，更新 UnitID
	if v, ok := args["unit_id"].(float64); ok {
		u := uint(v)
		user.UnitID = &u
		// 如果未传 unit_name，从单位表中补充
		if _, hasUnitName := args["unit_name"]; !hasUnitName {
			if unit, err := dao.GetUnitByID(uint(v)); err == nil && unit != nil {
				name := unit.Level3
				if name == "" {
					name = unit.Level2
				}
				if name == "" {
					name = unit.Level1
				}
				user.UnitName = name
			}
		}
	}
	if v, ok := args["permission_level"].(string); ok {
		user.PermissionLevel = model.PermissionLevel(v)
	}
	if v, ok := args["is_active"].(bool); ok {
		user.IsActive = v
	}
	if v, ok := args["is_admin"].(bool); ok {
		user.IsAdmin = v
	}
	return dao.UpdateUser(user)
}

func DeleteUser(args map[string]interface{}) error {
	idF, ok := args["id"].(float64)
	if !ok {
		return errors.New("id required")
	}
	id := uint(idF)
	_ = dao.DeleteUserSessions(id)
	return dao.DeleteUser(id)
}

func ResetPassword(args map[string]interface{}) error {
	idF, ok := args["id"].(float64)
	if !ok {
		return errors.New("id required")
	}
	newPwd, ok := args["password"].(string)
	if !ok || newPwd == "" {
		return errors.New("password required")
	}
	user, err := dao.GetUserByID(uint(idF))
	if err != nil {
		return err
	}
	user.PasswordHash = HashPassword(newPwd)
	return dao.UpdateUser(user)
}

// DispatchPermissions

func GetDispatchPermissions() ([]model.DispatchPermission, error) {
	return dao.GetDispatchPermissions()
}

func CreateDispatchPermission(args map[string]interface{}) error {
	perm := &model.DispatchPermission{}
	// 如果传了 unit_id，通过 ID 查找单位名
	if v, ok := args["unit_id"].(float64); ok {
		u := uint(v)
		unit, err := dao.GetUnitByID(u)
		if err == nil && unit != nil {
			perm.UnitID = &u
			perm.UnitName = unit.Level3
			if perm.UnitName == "" {
				perm.UnitName = unit.Level2
			}
			if perm.UnitName == "" {
				perm.UnitName = unit.Level1
			}
		}
	}
	if perm.UnitName == "" {
		return errors.New("unit_id required")
	}
	if v, ok := args["dispatch_scope"]; ok {
		b, _ := marshalJSON(v)
		perm.DispatchScope = model.JSONRaw(b)
	} else {
		perm.DispatchScope = model.JSONRaw("[]")
	}
	return dao.CreateDispatchPermission(perm)
}

func UpdateDispatchPermission(args map[string]interface{}) error {
	idF, ok := args["id"].(float64)
	if !ok {
		return errors.New("id required")
	}
	perm, err := dao.GetDispatchPermissionByID(uint(idF))
	if err != nil {
		return err
	}
	// 如果传了 unit_id，通过 ID 查找单位名
	if v, ok := args["unit_id"].(float64); ok {
		u := uint(v)
		unit, err := dao.GetUnitByID(u)
		if err == nil && unit != nil {
			perm.UnitID = &u
			perm.UnitName = unit.Level3
			if perm.UnitName == "" {
				perm.UnitName = unit.Level2
			}
			if perm.UnitName == "" {
				perm.UnitName = unit.Level1
			}
		}
	}
	if v, ok := args["dispatch_scope"]; ok {
		b, _ := marshalJSON(v)
		perm.DispatchScope = model.JSONRaw(b)
	}
	return dao.UpdateDispatchPermission(perm)
}

func DeleteDispatchPermission(args map[string]interface{}) error {
	idF, ok := args["id"].(float64)
	if !ok {
		return errors.New("id required")
	}
	return dao.DeleteDispatchPermission(uint(idF))
}

func CheckDispatchPermissionAPI(args map[string]interface{}, operator *model.PoliceUser) (bool, error) {
	targetUnit, ok := args["target_unit"].(string)
	if !ok || targetUnit == "" {
		return false, errors.New("target_unit required")
	}
	return CheckDispatchPermission(operator, targetUnit)
}

// SpecialFocus

func GetSpecialFocusList() ([]model.SpecialFocus, error) {
	return dao.GetAllSpecialFocuses()
}

func CreateSpecialFocus(args map[string]interface{}) error {
	sf := &model.SpecialFocus{}
	if v, ok := args["tag_name"].(string); ok {
		sf.TagName = v
	}
	if v, ok := args["description"].(string); ok {
		sf.Description = v
	}
	if sf.TagName == "" {
		return errors.New("tag_name required")
	}
	return dao.CreateSpecialFocus(sf)
}

func UpdateSpecialFocus(args map[string]interface{}) error {
	idF, ok := args["id"].(float64)
	if !ok {
		return errors.New("id required")
	}
	sf, err := dao.GetSpecialFocusByID(uint(idF))
	if err != nil {
		return err
	}
	if v, ok := args["tag_name"].(string); ok {
		sf.TagName = v
	}
	if v, ok := args["description"].(string); ok {
		sf.Description = v
	}
	return dao.UpdateSpecialFocus(sf)
}

func DeleteSpecialFocus(args map[string]interface{}) error {
	idF, ok := args["id"].(float64)
	if !ok {
		return errors.New("id required")
	}
	return dao.DeleteSpecialFocus(uint(idF))
}

// GetDispatchUnits returns units visible for dispatch given operator's permission
func GetDispatchUnits(operator *model.PoliceUser) ([]model.Unit, error) {
	allUnits, err := dao.GetAllUnits()
	if err != nil {
		return nil, err
	}
	switch operator.PermissionLevel {
	case model.PermissionCity:
		return allUnits, nil
	case model.PermissionDistrict:
		var result []model.Unit
		for _, u := range allUnits {
			if u.Level1 == operator.UnitName || u.Level2 == operator.UnitName || u.Level3 == operator.UnitName {
				result = append(result, u)
			}
		}
		return result, nil
	default:
		perm, err := dao.GetDispatchPermissionByUnit(operator.UnitName)
		if err != nil || perm == nil {
			return nil, nil
		}
		var scope []string
		_ = marshalJSONRaw(perm.DispatchScope, &scope)
		var result []model.Unit
		for _, u := range allUnits {
			name := u.Level3
			if name == "" {
				name = u.Level2
			}
			if name == "" {
				name = u.Level1
			}
			for _, s := range scope {
				if s == name {
					result = append(result, u)
					break
				}
			}
		}
		return result, nil
	}
}
