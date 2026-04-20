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

	// 基础菜单（所有用户可见）
	mainItems := []FrontMenuItem{
		{ID: "home", Name: "首页", Icon: "fa-home"},
		{ID: "letters", Name: "信件列表", Icon: "fa-envelope"},
		{ID: "processing", Name: "处理工作台", Icon: "fa-tools"},
		{ID: "statistics", Name: "统计分析", Icon: "fa-chart-bar"},
	}

	// 区县局及以上可见
	if user != nil && (user.PermissionLevel == model.PermissionCity || user.PermissionLevel == model.PermissionDistrict) {
		mainItems = append(mainItems,
			FrontMenuItem{ID: "dispatch", Name: "下发工作台", Icon: "fa-paper-plane"},
			FrontMenuItem{ID: "audit", Name: "核查工作台", Icon: "fa-search"},
		)
	}

	// 市局可见
	if user != nil && user.PermissionLevel == model.PermissionCity {
		mainItems = append(mainItems,
			FrontMenuItem{ID: "special-focus", Name: "专项关注", Icon: "fa-star"},
			FrontMenuItem{ID: "category", Name: "分类管理", Icon: "fa-tags"},
			FrontMenuItem{ID: "organization", Name: "组织机构", Icon: "fa-sitemap"},
		)
	}

	// 管理菜单
	adminItems := []FrontMenuItem{
		{ID: "settings", Name: "系统设置", Icon: "fa-cog"},
	}
	if user != nil && (user.PermissionLevel == model.PermissionCity || user.PermissionLevel == model.PermissionDistrict) {
		adminItems = append([]FrontMenuItem{{ID: "users", Name: "用户管理", Icon: "fa-users"}}, adminItems...)
	}
	adminItems = append(adminItems,
		FrontMenuItem{ID: "logout", Name: "退出登录", Icon: "fa-sign-out-alt", IsAction: true},
	)

	groups := []FrontMenuGroup{
		{Group: "工作台", Items: mainItems},
		{Group: "管理", Items: adminItems},
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
			model.StatusDispatched,
			model.StatusProcessing,
			model.StatusFeedback,
			model.StatusAudit,
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
