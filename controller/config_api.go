package controller

import (
	"net/http"

	"letter-manage-backend/dao"
	"letter-manage-backend/model"

	"github.com/gin-gonic/gin"
)

// FrontMenuItem 与前端 WorkplaceLayout 约定的菜单项格式
type FrontMenuItem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Icon     string `json:"icon"`
	IsAction bool   `json:"is_action"`
}

// FrontMenuGroup 菜单分组
type FrontMenuGroup struct {
	Group string          `json:"group"`
	Items []FrontMenuItem `json:"items"`
}

// ConfigController handles /api/config/
func ConfigController(c *gin.Context) {
	var req model.APIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResp("invalid request"))
		return
	}

	switch req.Order {
	case "get_menu":
		handleGetMenu(c)
	case "get_system_config":
		handleGetSystemConfig(c)
	default:
		c.JSON(http.StatusBadRequest, model.ErrorResp("unknown order: "+req.Order))
	}
}

func handleGetMenu(c *gin.Context) {
	// 从 Cookie 读取 session，获取当前用户（config 路由无 AuthRequired 中间件）
	var user *model.PoliceUser
	if sessionKey, err := c.Cookie("session_key"); err == nil && sessionKey != "" {
		if session, err := dao.GetSessionByKey(sessionKey); err == nil && session != nil {
			user = &session.User
		}
	}

	// 第一组：工作台（无分组标题，对应老代码中分割线前的菜单项）
	workbenchItems := []FrontMenuItem{
		{ID: "home", Name: "首页", Icon: "fa-home"},
		{ID: "letters", Name: "所有信件", Icon: "fa-envelope"},
		{ID: "processing", Name: "处理工作台", Icon: "fa-edit"},
	}

	// 区县局及以上可见的工作台菜单
	if user != nil && (user.PermissionLevel == model.PermissionCity || user.PermissionLevel == model.PermissionDistrict) {
		workbenchItems = append(workbenchItems,
			FrontMenuItem{ID: "dispatch", Name: "下发工作台", Icon: "fa-paper-plane"},
			FrontMenuItem{ID: "audit", Name: "核查工作台", Icon: "fa-check-circle"},
		)
	}

	// 统计工作台（所有用户可见）
	workbenchItems = append(workbenchItems,
		FrontMenuItem{ID: "statistics", Name: "统计工作台", Icon: "fa-chart-bar"},
	)

	// 第二组：管理员功能（对应老代码中“管理员功能”分组）
	adminItems := []FrontMenuItem{}
	if user != nil && (user.PermissionLevel == model.PermissionCity || user.PermissionLevel == model.PermissionDistrict) {
		adminItems = append(adminItems,
			FrontMenuItem{ID: "users", Name: "用户管理", Icon: "fa-users"},
		)
	}
	// 市局可见的管理员功能
	if user != nil && user.PermissionLevel == model.PermissionCity {
		adminItems = append(adminItems,
			FrontMenuItem{ID: "organization", Name: "组织管理", Icon: "fa-sitemap"},
			FrontMenuItem{ID: "special-focus", Name: "专项关注", Icon: "fa-star"},
			FrontMenuItem{ID: "category", Name: "分类管理", Icon: "fa-tags"},
		)
	}

	// 第三组：系统相关（对应老代码中“系统相关”分组）
	systemItems := []FrontMenuItem{
		{ID: "settings", Name: "系统设置", Icon: "fa-sliders-h"},
		{ID: "logout", Name: "退出登录", Icon: "fa-sign-out-alt", IsAction: true},
	}

	groups := []FrontMenuGroup{
		{Group: "", Items: workbenchItems},               // 第一组无标题，前端渲染时不会显示标题
		{Group: "管理员功能", Items: adminItems},          // 第二组标题“管理员功能”
		{Group: "系统相关", Items: systemItems},           // 第三组标题“系统相关”
	}

	// 同时返回用户信息
	var userInfo interface{}
	if user != nil {
		userInfo = map[string]interface{}{
			"id":               user.ID,
			"name":             user.Name,
			"nickname":         user.Nickname,
			"police_number":    user.PoliceNumber,
			"unit_name":        user.UnitName,
			"permission_level": user.PermissionLevel,
		}
	}

	c.JSON(http.StatusOK, model.SuccessResp(map[string]interface{}{
		"menu": groups,
		"user": userInfo,
	}))
}

func handleGetSystemConfig(c *gin.Context) {
	config := map[string]interface{}{
		"system_name":    "信件管理系统",
		"system_version": "1.0.0",
		"company":        "公安局",
		"logo":           "/assets/logo.png",
		"statuses": []string{
			model.StatusPreProcess,
			model.StatusCityDispatched,
			model.StatusCityDirectDispatch,
			model.StatusDispatched,
			model.StatusProcessing,
			model.StatusFeedback,
			model.StatusAudit,
			model.StatusDistrictAudited,
			model.StatusDone,
			model.StatusInvalid,
			model.StatusReturned,
		},
		"channels": []string{
			"网络",
			"来访",
			"电话",
			"信件",
			"其他",
		},
	}
	c.JSON(http.StatusOK, model.SuccessResp(config))
}
