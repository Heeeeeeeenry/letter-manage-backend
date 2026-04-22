package service

import (
	"encoding/json"
	"errors"

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

// Users

func GetUserList(args map[string]interface{}, currentUnitName string, permLevel string) (map[string]interface{}, error) {
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
	users, total, err := dao.GetUserList(page, pageSize, unitFilter, permLevel)
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
		// 权限级别中文映射
		permissionLevelChinese := "基层单位"
		switch user.PermissionLevel {
		case model.PermissionCity:
			permissionLevelChinese = "市级"
		case model.PermissionDistrict:
			permissionLevelChinese = "区级"
		case model.PermissionOfficer:
			permissionLevelChinese = "基层单位"
		}
		mappedUsers[i] = map[string]interface{}{
			"id":           user.ID,
			"姓名":          user.Name,
			"name":         user.Name,
			"警号":          user.PoliceNumber,
			"police_number": user.PoliceNumber,
			"所属单位":       user.UnitName,
			"org":          user.UnitName,
			"unit_name":    user.UnitName,
			"权限级别":       permissionLevelChinese,
			"role":         string(user.PermissionLevel),
			"permission_level": string(user.PermissionLevel),
			"状态":          statusText,
			"status":       statusText,
			"is_active":    user.IsActive,
			"nickname":     user.Nickname,
			"phone":        user.Phone,
			"created_at":   user.CreatedAt,
			"last_login":   user.LastLogin,
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
		user.UnitName = v
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
		user.UnitName = v
	}
	if v, ok := args["permission_level"].(string); ok {
		user.PermissionLevel = model.PermissionLevel(v)
	}
	if v, ok := args["is_active"].(bool); ok {
		user.IsActive = v
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
	if v, ok := args["unit_name"].(string); ok {
		perm.UnitName = v
	}
	if perm.UnitName == "" {
		return errors.New("unit_name required")
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
	if v, ok := args["unit_name"].(string); ok {
		perm.UnitName = v
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
