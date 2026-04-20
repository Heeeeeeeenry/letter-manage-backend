package controller

import (
	"net/http"

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
		data, err := service.GetUnitList(req.Args)
		if err != nil {
			c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
			return
		}
		c.JSON(http.StatusOK, model.SuccessResp(data))

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
		data, err := service.GetUserList(req.Args)
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
		if err := service.UpdateUser(req.Args); err != nil {
			c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
			return
		}
		c.JSON(http.StatusOK, model.SuccessResp(nil))

	case "delete_user":
		if user.PermissionLevel != model.PermissionCity {
			c.JSON(http.StatusOK, model.ErrorResp("无权限"))
			return
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
