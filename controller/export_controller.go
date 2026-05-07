package controller

import (
	"fmt"
	"net/http"
	"time"

	"letter-manage-backend/middleware"
	"letter-manage-backend/model"
	"letter-manage-backend/service"

	"github.com/gin-gonic/gin"
)

// ExportReportController 处理新的导出请求
// 路由: POST /api/letter/export_report
func ExportReportController(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResp("未登录"))
		return
	}

	var req struct {
		Period string `json:"period"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Period == "" {
		req.Period = "month"
	}

	filePath, _, err := service.GenerateFullExport(
		string(user.PermissionLevel),
		user.UnitID,
		req.Period,
	)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorResp("导出失败: "+err.Error()))
		return
	}

	periodLabel := getPeriodLabel(req.Period)
	c.FileAttachment(filePath, service.FormatZipFilename(periodLabel))
}

// handleExportNew 新导出处理函数（供旧API的export_monthly_report case调用）
func handleExportNew(c *gin.Context, args map[string]interface{}, user *model.PoliceUser) {
	period, _ := args["period"].(string)

	filePath, gaps, err := service.GenerateFullExport(
		string(user.PermissionLevel),
		user.UnitID,
		period,
	)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorResp("导出失败: "+err.Error()))
		return
	}

	// 如果有数据缺失，记录日志
	if len(gaps) > 0 {
		for _, g := range gaps {
			if g.Status == "empty" {
				c.Header("X-Export-Gap", g.Field+": "+g.Advice)
			}
		}
	}

	periodLabel := getPeriodLabel(period)
	c.FileAttachment(filePath, service.FormatZipFilename(periodLabel))
}

func getPeriodLabel(period string) string {
	now := time.Now()
	switch period {
	case "day":
		return "今日"
	case "week":
		return "本周"
	case "month":
		return fmt.Sprintf("%d月", now.Month())
	case "year":
		return fmt.Sprintf("%d年", now.Year())
	default:
		return fmt.Sprintf("%d月", now.Month())
	}
}
