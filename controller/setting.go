package controller

import (
	"net/http"

	"letter-manage-backend/dao"
	"letter-manage-backend/middleware"
	"letter-manage-backend/model"
	"letter-manage-backend/service"

	"github.com/gin-gonic/gin"
)

// SettingController handles /api/setting/
func SettingController(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResp("未登录"))
		return
	}

	var req model.APIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResp("invalid request"))
		return
	}

	switch req.Order {
	// Category
	case "category_list":
		cats, err := service.GetCategoryList()
		if err != nil {
			c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
			return
		}
		c.JSON(http.StatusOK, model.SuccessResp(cats))

	case "category_create":
		if err := service.CreateCategory(req.Args); err != nil {
			c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
			return
		}
		c.JSON(http.StatusOK, model.SuccessResp(nil))

	case "category_update":
		if err := service.UpdateCategory(req.Args); err != nil {
			c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
			return
		}
		c.JSON(http.StatusOK, model.SuccessResp(nil))

	case "category_delete":
		if err := service.DeleteCategory(req.Args); err != nil {
			c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
			return
		}
		c.JSON(http.StatusOK, model.SuccessResp(nil))

	// Units
	case "get_units":
		result, err := service.GetUnitsWithFilter(req.Args)
		if err != nil {
			c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
			return
		}
		c.JSON(http.StatusOK, model.SuccessResp(result))

	case "get_dispatch_units":
		units, err := service.GetDispatchUnits(user)
		if err != nil {
			c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
			return
		}
		c.JSON(http.StatusOK, model.SuccessResp(units))

	case "create_unit":
		if user.PermissionLevel != model.PermissionCity {
			c.JSON(http.StatusOK, model.ErrorResp("无权限"))
			return
		}
		if err := service.CreateUnit(req.Args); err != nil {
			c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
			return
		}
		c.JSON(http.StatusOK, model.SuccessResp(nil))

	case "update_unit":
		if user.PermissionLevel != model.PermissionCity {
			c.JSON(http.StatusOK, model.ErrorResp("无权限"))
			return
		}
		if err := service.UpdateUnit(req.Args); err != nil {
			c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
			return
		}
		c.JSON(http.StatusOK, model.SuccessResp(nil))

	case "delete_unit":
		if user.PermissionLevel != model.PermissionCity {
			c.JSON(http.StatusOK, model.ErrorResp("无权限"))
			return
		}
		if err := service.DeleteUnit(req.Args); err != nil {
			c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
			return
		}
		c.JSON(http.StatusOK, model.SuccessResp(nil))

	// Users
	case "get_user_list":
		if user.PermissionLevel == model.PermissionOfficer {
			c.JSON(http.StatusOK, model.ErrorResp("无权限"))
			return
		}
		data, err := service.GetUserList(req.Args, user.UnitName, string(user.PermissionLevel), user.IsAdmin, user.UnitID)
		if err != nil {
			c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
			return
		}
		c.JSON(http.StatusOK, model.SuccessResp(data))

	case "create_user":
		if user.PermissionLevel == model.PermissionOfficer {
			c.JSON(http.StatusOK, model.ErrorResp("无权限"))
			return
		}
		// 检查目标用户级别，不能创建同级用户除非有管理员权限
		targetLevel, _ := req.Args["permission_level"].(string)
		if targetLevel != "" && string(user.PermissionLevel) == targetLevel && !user.IsAdmin {
			c.JSON(http.StatusOK, model.ErrorResp("无权限创建同级用户"))
			return
		}
		// 检查目标单位是否在当前用户的管理范围内
		targetUnitIDf, _ := req.Args["unit_id"].(float64)
		if targetUnitIDf > 0 {
			targetUnitID := uint(targetUnitIDf)
			if !isUnitInScope(user, "", &targetUnitID) {
				c.JSON(http.StatusOK, model.ErrorResp("无权限管理该单位的用户"))
				return
			}
		}
		if err := service.CreateUser(req.Args); err != nil {
			c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
			return
		}
		c.JSON(http.StatusOK, model.SuccessResp(nil))

	case "update_user":
		if user.PermissionLevel == model.PermissionOfficer {
			c.JSON(http.StatusOK, model.ErrorResp("无权限"))
			return
		}
		// 权限校验规则：
		// 1. 非admin不能操作同级用户
		// 2. admin不能编辑同级admin
		// 3. 不能操作上级
		// 4. 必须在单位管理范围内
		// 5. admin编辑同级用户时，手机号只读（被后端剥离）
		idF, _ := req.Args["id"].(float64)
		if idF > 0 {
			targetUser, err := dao.GetUserByID(uint(idF))
			if err == nil && targetUser != nil {
				// 规则1：非admin不能操作同级
				if !user.IsAdmin && user.PermissionLevel == targetUser.PermissionLevel {
					c.JSON(http.StatusOK, model.ErrorResp("无权限修改同级用户"))
					return
				}
				// 规则2：admin不能编辑同级admin
				if user.IsAdmin && targetUser.IsAdmin && user.PermissionLevel == targetUser.PermissionLevel {
					c.JSON(http.StatusOK, model.ErrorResp("无权限修改同级管理员"))
					return
				}
				// 规则3：不能操作上级
				if service.LevelRank(string(user.PermissionLevel)) < service.LevelRank(string(targetUser.PermissionLevel)) {
					c.JSON(http.StatusOK, model.ErrorResp("无权限修改上级用户"))
					return
				}
				// 规则4：检查目标用户的单位是否在当前用户的管理范围内
				if !isUnitInScope(user, targetUser.UnitName, targetUser.UnitID) {
					c.JSON(http.StatusOK, model.ErrorResp("无权限管理该单位的用户"))
					return
				}
				// 规则5：管理员不能编辑同级用户的手机号（可见但只读）
				if user.IsAdmin && user.PermissionLevel == targetUser.PermissionLevel {
					delete(req.Args, "phone")
				}
			}
		}
		if err := service.UpdateUser(req.Args); err != nil {
			c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
			return
		}
		c.JSON(http.StatusOK, model.SuccessResp(nil))

	case "delete_user":
		if user.PermissionLevel != model.PermissionCity {
			// 区县局管理员可删除本单位及下属单位的用户
			idF, _ := req.Args["id"].(float64)
			if idF > 0 {
				targetUser, err := dao.GetUserByID(uint(idF))
				if err == nil && targetUser != nil {
					// 权限校验规则（同update_user）：
					// 1. 非admin不能操作同级用户
					if !user.IsAdmin && user.PermissionLevel == targetUser.PermissionLevel {
						c.JSON(http.StatusOK, model.ErrorResp("无权限删除同级用户"))
						return
					}
					// 2. admin不能删除同级admin
					if user.IsAdmin && targetUser.IsAdmin && user.PermissionLevel == targetUser.PermissionLevel {
						c.JSON(http.StatusOK, model.ErrorResp("无权限删除同级管理员"))
						return
					}
					// 3. 不能操作上级
					if service.LevelRank(string(user.PermissionLevel)) < service.LevelRank(string(targetUser.PermissionLevel)) {
						c.JSON(http.StatusOK, model.ErrorResp("无权限删除上级用户"))
						return
					}
					// 4. 必须在单位管理范围内
					if !isUnitInScope(user, targetUser.UnitName, targetUser.UnitID) {
						c.JSON(http.StatusOK, model.ErrorResp("无权限管理该单位的用户"))
						return
					}
				}
			} else {
				c.JSON(http.StatusOK, model.ErrorResp("无权限"))
				return
			}
		}
		if err := service.DeleteUser(req.Args); err != nil {
			c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
			return
		}
		c.JSON(http.StatusOK, model.SuccessResp(nil))

	case "reset_password":
		if user.PermissionLevel == model.PermissionOfficer {
			c.JSON(http.StatusOK, model.ErrorResp("无权限"))
			return
		}
		// 权限校验规则（同update_user）：
		idF, _ := req.Args["id"].(float64)
		if idF > 0 {
			targetUser, err := dao.GetUserByID(uint(idF))
			if err == nil && targetUser != nil {
				// 1. 非admin不能操作同级
				if !user.IsAdmin && user.PermissionLevel == targetUser.PermissionLevel {
					c.JSON(http.StatusOK, model.ErrorResp("无权限修改同级用户密码"))
					return
				}
				// 2. admin不能操作同级admin
				if user.IsAdmin && targetUser.IsAdmin && user.PermissionLevel == targetUser.PermissionLevel {
					c.JSON(http.StatusOK, model.ErrorResp("无权限修改同级管理员密码"))
					return
				}
				// 3. 不能操作上级
				if service.LevelRank(string(user.PermissionLevel)) < service.LevelRank(string(targetUser.PermissionLevel)) {
					c.JSON(http.StatusOK, model.ErrorResp("无权限修改上级用户密码"))
					return
				}
				// 4. 检查目标用户的单位是否在当前用户的管理范围内
				if !isUnitInScope(user, targetUser.UnitName, targetUser.UnitID) {
					c.JSON(http.StatusOK, model.ErrorResp("无权限管理该单位的用户"))
					return
				}
			}
		}
		if err := service.ResetPassword(req.Args); err != nil {
			c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
			return
		}
		c.JSON(http.StatusOK, model.SuccessResp(nil))

	// Dispatch Permissions
	case "get_dispatch_permissions":
		if user.PermissionLevel != model.PermissionCity {
			c.JSON(http.StatusOK, model.ErrorResp("无权限"))
			return
		}
		perms, err := service.GetDispatchPermissions()
		if err != nil {
			c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
			return
		}
		c.JSON(http.StatusOK, model.SuccessResp(perms))

	case "create_dispatch_permission":
		if user.PermissionLevel != model.PermissionCity {
			c.JSON(http.StatusOK, model.ErrorResp("无权限"))
			return
		}
		if err := service.CreateDispatchPermission(req.Args); err != nil {
			c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
			return
		}
		c.JSON(http.StatusOK, model.SuccessResp(nil))

	case "update_dispatch_permission":
		if user.PermissionLevel != model.PermissionCity {
			c.JSON(http.StatusOK, model.ErrorResp("无权限"))
			return
		}
		if err := service.UpdateDispatchPermission(req.Args); err != nil {
			c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
			return
		}
		c.JSON(http.StatusOK, model.SuccessResp(nil))

	case "delete_dispatch_permission":
		if user.PermissionLevel != model.PermissionCity {
			c.JSON(http.StatusOK, model.ErrorResp("无权限"))
			return
		}
		if err := service.DeleteDispatchPermission(req.Args); err != nil {
			c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
			return
		}
		c.JSON(http.StatusOK, model.SuccessResp(nil))

	case "check_dispatch_permission":
		ok, err := service.CheckDispatchPermissionAPI(req.Args, user)
		if err != nil {
			c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
			return
		}
		c.JSON(http.StatusOK, model.SuccessResp(map[string]interface{}{
			"can_dispatch": ok,
		}))

	// SpecialFocus
	case "get_special_focus_list":
		sfs, err := service.GetSpecialFocusList()
		if err != nil {
			c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
			return
		}
		c.JSON(http.StatusOK, model.SuccessResp(sfs))

	case "create_special_focus":
		if user.PermissionLevel != model.PermissionCity {
			c.JSON(http.StatusOK, model.ErrorResp("无权限"))
			return
		}
		if err := service.CreateSpecialFocus(req.Args); err != nil {
			c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
			return
		}
		c.JSON(http.StatusOK, model.SuccessResp(nil))

	case "update_special_focus":
		if user.PermissionLevel != model.PermissionCity {
			c.JSON(http.StatusOK, model.ErrorResp("无权限"))
			return
		}
		if err := service.UpdateSpecialFocus(req.Args); err != nil {
			c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
			return
		}
		c.JSON(http.StatusOK, model.SuccessResp(nil))

	case "delete_special_focus":
		if user.PermissionLevel != model.PermissionCity {
			c.JSON(http.StatusOK, model.ErrorResp("无权限"))
			return
		}
		if err := service.DeleteSpecialFocus(req.Args); err != nil {
			c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
			return
		}
		c.JSON(http.StatusOK, model.SuccessResp(nil))

	default:
		c.JSON(http.StatusBadRequest, model.ErrorResp("unknown order: "+req.Order))
	}
}

// isUnitInScope 检查目标单位是否在当前用户的管理范围内
func isUnitInScope(user *model.PoliceUser, targetUnit string, targetUnitID ...*uint) bool {
	if targetUnit == "" && len(targetUnitID) == 0 {
		return false
	}
	// 如果传了 targetUnitID，使用 ID 判断
	if len(targetUnitID) > 0 && targetUnitID[0] != nil {
		switch user.PermissionLevel {
		case model.PermissionCity:
			return true
		case model.PermissionDistrict:
			// 检查是否本单位
			if user.UnitID != nil && *user.UnitID == *targetUnitID[0] {
				return true
			}
			// 检查是否下属单位
			subIDs := dao.GetSubordinateUnitIDs(*user.UnitID)
			for _, sid := range subIDs {
				if sid == *targetUnitID[0] {
					return true
				}
			}
			return false
		default:
			// OFFICER：只能管理本单位
			if user.UnitID != nil {
				return *user.UnitID == *targetUnitID[0]
			}
			return false
		}
	}
	// 没有 targetUnitID，fallback 到字符串判断
	if targetUnit == "" {
		return false
	}
	switch user.PermissionLevel {
	case model.PermissionCity:
		// 市局：可管理所有单位
		return true
	case model.PermissionDistrict:
		// 区县局：可管理本单位及下属单位
		if user.UnitName == targetUnit {
			return true
		}
		subUnits := dao.GetSubordinateUnitNames(user.UnitName)
		for _, u := range subUnits {
			if u == targetUnit {
				return true
			}
		}
		return false
	default:
		// OFFICER：只能管理本单位
		return user.UnitName == targetUnit
	}
}
