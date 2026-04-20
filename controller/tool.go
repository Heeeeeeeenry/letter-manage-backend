package controller

import (
	"net/http"

	"letter-manage-backend/model"
	"letter-manage-backend/service"

	"github.com/gin-gonic/gin"
)

// ToolController handles /api/tool/ (unified dispatch) and sub-paths
func ToolController(c *gin.Context) {
	var req model.APIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResp("invalid request"))
		return
	}
	dispatchTool(c, req.Order, req.Args)
}

// Sub-path handlers

func ToolTimeDiff(c *gin.Context) {
	args := parseToolArgs(c)
	data, err := service.TimeDiff(args)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(data))
}

func ToolTimeAdd(c *gin.Context) {
	args := parseToolArgs(c)
	data, err := service.TimeAdd(args)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(data))
}

func ToolHolidayCheck(c *gin.Context) {
	args := parseToolArgs(c)
	data, err := service.HolidayCheck(args)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(data))
}

func ToolWorkdaysCalculate(c *gin.Context) {
	args := parseToolArgs(c)
	data, err := service.WorkdaysCalculate(args)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(data))
}

func ToolWorkdaysAdd(c *gin.Context) {
	args := parseToolArgs(c)
	data, err := service.WorkdaysAdd(args)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(data))
}

func ToolMonthCalendar(c *gin.Context) {
	args := parseToolArgs(c)
	data, err := service.MonthCalendar(args)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(data))
}

// parseToolArgs parses the request body or form data into args map
func parseToolArgs(c *gin.Context) map[string]interface{} {
	var req model.APIRequest
	if err := c.ShouldBindJSON(&req); err == nil && req.Args != nil {
		return req.Args
	}
	// try reading as plain JSON map
	var args map[string]interface{}
	c.ShouldBindJSON(&args)
	return args
}

func dispatchTool(c *gin.Context, order string, args map[string]interface{}) {
	var (
		data interface{}
		err  error
	)
	switch order {
	case "time_diff":
		data, err = service.TimeDiff(args)
	case "time_add":
		data, err = service.TimeAdd(args)
	case "holiday_check":
		data, err = service.HolidayCheck(args)
	case "workdays_calculate":
		data, err = service.WorkdaysCalculate(args)
	case "workdays_add":
		data, err = service.WorkdaysAdd(args)
	case "month_calendar":
		data, err = service.MonthCalendar(args)
	default:
		c.JSON(http.StatusBadRequest, model.ErrorResp("unknown tool: "+order))
		return
	}
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(data))
}
