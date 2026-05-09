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
	filterLevel3 := ""
	if v, ok := args["filter_level3"].(string); ok {
		filterLevel3 = v
	}

	// 调用 DAO
	units, total, err := dao.GetUnitsWithFilter(page, pageSize, searchKeyword, filterLevel1, filterLevel2, filterLevel3)
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

func GetUserList(args map[string]interface{}, permLevel string, currentIsAdmin bool, currentUnitID *uint, currentUserID uint) (map[string]interface{}, error) {
	page := 1
	pageSize := 20
	if v, ok := args["page"].(float64); ok {
		page = int(v)
	}
	if v, ok := args["page_size"].(float64); ok {
		pageSize = int(v)
	}
	// 解析筛选参数
	keyword := ""
	if v, ok := args["keyword"].(string); ok {
		keyword = v
	}
	permLevelFilter := ""
	if v, ok := args["perm_level"].(string); ok {
		permLevelFilter = v
	}
	var isActiveFilter *bool
	if v, ok := args["is_active"]; ok {
		switch val := v.(type) {
		case bool:
			isActiveFilter = &val
		case float64:
			b := val != 0
			isActiveFilter = &b
		}
	}
	// 权限数据隔离：根据用户级别限制可见的用户范围
	var unitIDSlice []*uint
	if currentUnitID != nil {
		unitIDSlice = append(unitIDSlice, currentUnitID)
	}
	switch permLevel {
	case "CITY":
		// 市局：可见所有用户，DAO 中处理
	case "DISTRICT":
		// 区县局：DAO 中根据 unitID 和 permLevel 处理权限
	default:
		// OFFICER 无权访问用户管理，DAO 中拦截
	}
	users, total, err := dao.GetUserList(page, pageSize, "", permLevel, currentIsAdmin, currentUserID, unitIDSlice, keyword, permLevelFilter, isActiveFilter)
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
	// 如果传了 unit_id，通过 ID 查找单位
	if v, ok := args["unit_id"].(float64); ok {
		unit, err := dao.GetUnitByID(uint(v))
		if err == nil && unit != nil {
			u := uint(v)
			user.UnitID = &u
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
	if v, ok := args["unit_id"].(float64); ok {
		u := uint(v)
		user.UnitID = &u
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

// GetUsersInUnit 获取某单位下的活跃用户列表（用于下发选人）
func GetUsersInUnit(args map[string]interface{}) ([]model.PoliceUser, error) {
	idF, ok := args["unit_id"].(float64)
	if !ok {
		return nil, errors.New("unit_id required")
	}
	return dao.GetActiveUsersByUnitID(uint(idF))
}

// DispatchPermissions

func GetDispatchPermissions() ([]model.DispatchPermission, error) {
	return dao.GetDispatchPermissions()
}

func CreateDispatchPermission(args map[string]interface{}) error {
	perm := &model.DispatchPermission{}
	// 优先用 unit_id，没传则通过 unit_name 查找
	if v, ok := args["unit_id"].(float64); ok {
		u := uint(v)
		unit, err := dao.GetUnitByID(u)
		if err == nil && unit != nil {
			perm.UnitID = &u
			perm.UnitName = pickUnitName(unit)
		}
	}
	if perm.UnitName == "" {
		if v, ok := args["unit_name"].(string); ok && v != "" {
			perm.UnitName = v
			// 尝试通过全路径名查找 unit_id
			u, err := dao.GetUnitByFullName(v)
			if err == nil && u != nil {
				perm.UnitID = &u.ID
			}
		}
	}
	if perm.UnitName == "" {
		return errors.New("unit_id or unit_name required")
	}
	if v, ok := args["dispatch_scope"]; ok {
		b, _ := marshalJSON(v)
		perm.CanDispatchTo = string(b)
	} else {
		perm.CanDispatchTo = "[]"
	}
	return dao.CreateDispatchPermission(perm)
}

func pickUnitName(unit *model.Unit) string {
	if unit.Level3 != "" {
		return unit.Level3
	}
	if unit.Level2 != "" {
		return unit.Level2
	}
	return unit.Level1
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
	// 如果传了 unit_id，通过 ID 更新单位名
	if v, ok := args["unit_id"].(float64); ok {
		u := uint(v)
		unit, err := dao.GetUnitByID(u)
		if err == nil && unit != nil {
			perm.UnitID = &u
			perm.UnitName = pickUnitName(unit)
		}
	}
	if perm.UnitName == "" {
		if v, ok := args["unit_name"].(string); ok && v != "" {
			perm.UnitName = v
			u, err := dao.GetUnitByFullName(v)
			if err == nil && u != nil {
				perm.UnitID = &u.ID
			}
		}
	}
	if v, ok := args["dispatch_scope"]; ok {
		b, _ := marshalJSON(v)
		perm.CanDispatchTo = string(b)
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

func GetSpecialFocusList() ([]map[string]interface{}, error) {
	sfs, err := dao.GetAllSpecialFocuses()
	if err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, len(sfs))
	for i, sf := range sfs {
		count := dao.CountLettersByFocusID(sf.ID)
		result[i] = map[string]interface{}{
			"id":           sf.ID,
			"name":         sf.Name,
			"description":  sf.Description,
			"letter_count": count,
			"created_at":   sf.CreatedAt,
		}
	}
	return result, nil
}

func CreateSpecialFocus(args map[string]interface{}) error {
	sf := &model.SpecialFocus{}
	if v, ok := args["tag_name"].(string); ok {
		sf.Name = v
	}
	if v, ok := args["description"].(string); ok {
		sf.Description = v
	}
	if sf.Name == "" {
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
		sf.Name = v
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
	opUnit, _ := dao.GetUnitByID(*operator.UnitID)
	switch operator.PermissionLevel {
	case model.PermissionCity:
		return allUnits, nil
	case model.PermissionDistrict:
		// 区县局：只能下发到本级及下级单位
		var result []model.Unit
		for _, u := range allUnits {
			// 同分局 (Level1+Level2 相同)，排除市局
			if opUnit != nil && u.Level1 == opUnit.Level1 && u.Level2 == opUnit.Level2 {
				result = append(result, u)
			}
		}
		return result, nil
	default:
		// OFFICER：优先查下发权限表；若未配置且单位为"民意智感中心"，默认可下发到同级单位
		shortName := dao.GetUnitNameByID(operator.UnitID)
		perm, _ := dao.GetDispatchPermissionByUnit(shortName)
		if perm != nil {
			var scope []string
			json.Unmarshal([]byte(perm.CanDispatchTo), &scope)
			var result []model.Unit
			for _, u := range allUnits {
				name := pickName(u)
				for _, s := range scope {
					if s == name {
						result = append(result, u)
						break
					}
				}
			}
			return result, nil
		}
		// 无下发权限配置：民意智感中心默认可下发到同分局所有单位
		if shortName == "民意智感中心" && opUnit != nil {
			var result []model.Unit
			for _, u := range allUnits {
				// 同分局 (Level1+Level2相同)，排除市局和自身
				if u.Level1 == opUnit.Level1 && u.Level2 == opUnit.Level2 && u.Level3 != "" {
					result = append(result, u)
				}
			}
			return result, nil
		}
		return nil, nil
	}
}

func pickName(u model.Unit) string {
	if u.Level3 != "" { return u.Level3 }
	if u.Level2 != "" { return u.Level2 }
	return u.Level1
}
