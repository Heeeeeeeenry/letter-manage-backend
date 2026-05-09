package controller

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"letter-manage-backend/config"
	"letter-manage-backend/dao"
	"letter-manage-backend/middleware"
	"letter-manage-backend/model"
	"letter-manage-backend/service"

	"github.com/gin-gonic/gin"
)

// LetterController handles /api/letter/
func LetterController(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResp("未登录"))
		return
	}

	// Check if multipart upload
	contentType := c.ContentType()
	if contentType == "multipart/form-data" || (len(contentType) > 19 && contentType[:19] == "multipart/form-data") {
		handleFileUpload(c, user)
		return
	}

	var req model.APIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResp("invalid request"))
		return
	}

	switch req.Order {
	case "get_list":
		handleGetList(c, req.Args, user)
	case "get_dispatch_list":
		handleGetDispatchList(c, req.Args, user)
	case "get_processing_list":
		handleGetProcessingList(c, req.Args, user)
	case "get_audit_list":
		handleGetAuditList(c, req.Args, user)
	case "get_detail":
		handleGetDetail(c, req.Args, user)
	case "get_files":
		handleGetFiles(c, req.Args, user)
	case "get_by_phone":
		handleGetByPhone(c, req.Args, user)
	case "get_by_idcard":
		handleGetByIDCard(c, req.Args, user)
	case "create":
		handleCreateLetter(c, req.Args, user)
	case "update":
		handleUpdateLetter(c, req.Args, user)
	case "delete":
		handleDeleteLetter(c, req.Args, user)
	case "update_status":
		handleUpdateStatus(c, req.Args, user)
	case "get_statistics":
		handleGetStatistics(c, req.Args, user)
	case "get_attachments":
		handleGetAttachments(c, req.Args, user)
	case "update_attachments":
		handleUpdateAttachments(c, req.Args)
	case "get_categories":
		handleGetCategories(c)
	case "dispatch":
		handleDispatch(c, req.Args, user)
	case "analyze_letter":
		handleAnalyzeLetter(c, req.Args)
	case "auto_dispatch":
		handleAutoDispatch(c, req.Args, user)
	case "mark_invalid":
		handleMarkInvalid(c, req.Args, user)
	case "submit_processing":
		handleSubmitProcessing(c, req.Args, user)
	case "handle_by_self":
		handleHandleBySelf(c, req.Args, user)
	case "set_special_focus":
		handleSetSpecialFocus(c, req.Args)
	case "get_letter_special_focus":
		handleGetLetterSpecialFocus(c, req.Args)
	case "return_letter":
		handleReturnLetter(c, req.Args, user)
	case "audit_approve":
		handleAuditApprove(c, req.Args, user)
	case "audit_reject":
		handleAuditReject(c, req.Args, user)
	case "export":
		handleExport(c, req.Args, user)
	case "export_monthly_report":
		handleExportNew(c, req.Args, user)
	default:
		c.JSON(http.StatusBadRequest, model.ErrorResp("unknown order: "+req.Order))
	}
}

func handleGetList(c *gin.Context, args map[string]interface{}, user *model.PoliceUser) {
	data, err := service.GetLetterList(args, user)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(data))
}

func handleGetDispatchList(c *gin.Context, args map[string]interface{}, user *model.PoliceUser) {
	data, err := service.GetDispatchList(user.UnitID, string(user.PermissionLevel), args)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(data))
}

func handleGetProcessingList(c *gin.Context, args map[string]interface{}, user *model.PoliceUser) {
	data, err := service.GetProcessingList(user.UnitID, string(user.PermissionLevel), args, user.ID)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(data))
}

func handleGetAuditList(c *gin.Context, args map[string]interface{}, user *model.PoliceUser) {
	data, err := service.GetAuditList(user.UnitID, string(user.PermissionLevel), args)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(data))
}

func handleGetDetail(c *gin.Context, args map[string]interface{}, user *model.PoliceUser) {
	letterNo, ok := args["letter_no"].(string)
	if !ok || letterNo == "" {
		c.JSON(http.StatusOK, model.ErrorResp("letter_no required"))
		return
	}
	data, err := service.GetLetterDetail(letterNo, string(user.PermissionLevel), user.UnitID)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(data))
}

func handleGetFiles(c *gin.Context, args map[string]interface{}, user *model.PoliceUser) {
	letterNo, ok := args["letter_no"].(string)
	if !ok || letterNo == "" {
		c.JSON(http.StatusOK, model.ErrorResp("letter_no required"))
		return
	}
	att, err := service.GetAttachments(letterNo)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(att))
}

func handleGetByPhone(c *gin.Context, args map[string]interface{}, user *model.PoliceUser) {
	phone, ok := args["phone"].(string)
	if !ok || phone == "" {
		c.JSON(http.StatusOK, model.ErrorResp("phone required"))
		return
	}
	letters, err := service.GetLettersByPhone(phone, string(user.PermissionLevel), user.UnitID)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(letters))
}

func handleGetByIDCard(c *gin.Context, args map[string]interface{}, user *model.PoliceUser) {
	idCard, ok := args["id_card"].(string)
	if !ok || idCard == "" {
		c.JSON(http.StatusOK, model.ErrorResp("id_card required"))
		return
	}
	letters, err := service.GetLettersByIDCard(idCard, string(user.PermissionLevel), user.UnitID)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(letters))
}

func handleCreateLetter(c *gin.Context, args map[string]interface{}, user *model.PoliceUser) {
	letter, err := service.CreateLetter(args)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(letter))
	citizenName, _ := args["citizen_name"].(string)
	service.AddOperationLog(user.ID, user.Name, user.PoliceNumber, "create", "信函登记", letter.LetterNo, fmt.Sprintf("新增信件，编号:%s，群众:%s", letter.LetterNo, citizenName))
}

func handleUpdateLetter(c *gin.Context, args map[string]interface{}, user *model.PoliceUser) {
	letterNo, _ := args["letter_no"].(string)
	oldLetter, _ := dao.GetLetterByNo(letterNo)
	if err := service.UpdateLetter(args); err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(nil))
	var detail string
	if oldLetter != nil {
		if newName, ok := args["citizen_name"].(string); ok && newName != oldLetter.CitizenName {
			detail = fmt.Sprintf("群众名称从%s改为%s", oldLetter.CitizenName, newName)
		}
	}
	service.AddOperationLog(user.ID, user.Name, user.PoliceNumber, "update", "信函修改", letterNo, detail)
}

func handleDeleteLetter(c *gin.Context, args map[string]interface{}, user *model.PoliceUser) {
	if err := service.DeleteLetter(args); err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(nil))
	letterNo, _ := args["letter_no"].(string)
	service.AddOperationLog(user.ID, user.Name, user.PoliceNumber, "delete", "信件", letterNo, fmt.Sprintf("删除信件，编号:%s", letterNo))
}

func handleUpdateStatus(c *gin.Context, args map[string]interface{}, user *model.PoliceUser) {
	if err := service.UpdateLetterStatus(args, user); err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(nil))
}

func handleGetStatistics(c *gin.Context, args map[string]interface{}, user *model.PoliceUser) {
	period, _ := args["period"].(string)
	viewMode, _ := args["view_mode"].(string)
	// 优先使用前端传的 unit_id，否则用用户自身单位
	unitID := user.UnitID
	if v, ok := args["unit_id"].(float64); ok && v > 0 {
		uid := uint(v)
		unitID = &uid
	}
	data, err := service.GetStatistics(string(user.PermissionLevel), period, unitID, user.ID, viewMode)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(data))
}

func handleGetAttachments(c *gin.Context, args map[string]interface{}, user *model.PoliceUser) {
	letterNo, ok := args["letter_no"].(string)
	if !ok || letterNo == "" {
		c.JSON(http.StatusOK, model.ErrorResp("letter_no required"))
		return
	}
	att, err := service.GetAttachments(letterNo)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(att))
}

func handleUpdateAttachments(c *gin.Context, args map[string]interface{}) {
	if err := service.UpdateAttachments(args); err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(nil))
}

func handleGetCategories(c *gin.Context) {
	tree, err := service.GetCategories()
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(tree))
}

func handleDispatch(c *gin.Context, args map[string]interface{}, user *model.PoliceUser) {
	if err := service.DispatchLetter(args, user); err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(nil))
	letterNo, _ := args["letter_no"].(string)
	targetUnit, _ := args["target_unit"].(string)
	service.AddOperationLog(user.ID, user.Name, user.PoliceNumber, "dispatch", "下发工作台", letterNo, fmt.Sprintf("下发信件，编号:%s，下发至:%s", letterNo, targetUnit))
}

func handleAnalyzeLetter(c *gin.Context, args map[string]interface{}) {
	letterNo, ok := args["letter_no"].(string)
	if !ok || letterNo == "" {
		c.JSON(http.StatusOK, model.ErrorResp("letter_no required"))
		return
	}
	analysis, err := service.AnalyzeLetterForDispatch(letterNo)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(analysis))
}

func handleAutoDispatch(c *gin.Context, args map[string]interface{}, user *model.PoliceUser) {
	result, err := service.AutoDispatchLetter(args, user)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(result))
	letterNo, _ := args["letter_no"].(string)
	service.AddOperationLog(user.ID, user.Name, user.PoliceNumber, "auto_dispatch", "下发工作台", letterNo, fmt.Sprintf("AI自动下发，编号:%s", letterNo))
}

func handleMarkInvalid(c *gin.Context, args map[string]interface{}, user *model.PoliceUser) {
	if err := service.MarkInvalid(args, user); err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(nil))
	letterNo, _ := args["letter_no"].(string)
	service.AddOperationLog(user.ID, user.Name, user.PoliceNumber, "mark_invalid", "信函管理", letterNo, fmt.Sprintf("标记无效，编号:%s", letterNo))
}

func handleSubmitProcessing(c *gin.Context, args map[string]interface{}, user *model.PoliceUser) {
	if err := service.SubmitProcessing(args, user); err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(nil))
	letterNo, _ := args["letter_no"].(string)
	service.AddOperationLog(user.ID, user.Name, user.PoliceNumber, "submit_processing", "处理工作台", letterNo, fmt.Sprintf("提交处理，编号:%s", letterNo))
}

func handleHandleBySelf(c *gin.Context, args map[string]interface{}, user *model.PoliceUser) {
	if err := service.HandleBySelf(args, user); err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(nil))
	letterNo, _ := args["letter_no"].(string)
	service.AddOperationLog(user.ID, user.Name, user.PoliceNumber, "handle_by_self", "处理工作台", letterNo, fmt.Sprintf("自行处理，编号:%s", letterNo))
}

func handleReturnLetter(c *gin.Context, args map[string]interface{}, user *model.PoliceUser) {
	if err := service.ReturnLetter(args, user); err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(nil))
	letterNo, _ := args["letter_no"].(string)
	service.AddOperationLog(user.ID, user.Name, user.PoliceNumber, "return_letter", "核查工作台", letterNo, fmt.Sprintf("退回信件，编号:%s", letterNo))
}

func handleAuditApprove(c *gin.Context, args map[string]interface{}, user *model.PoliceUser) {
	if err := service.AuditApprove(args, user); err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(nil))
	letterNo, _ := args["letter_no"].(string)
	service.AddOperationLog(user.ID, user.Name, user.PoliceNumber, "审核通过", "核查工作台", letterNo, fmt.Sprintf("核查通过，编号:%s", letterNo))
}

func handleAuditReject(c *gin.Context, args map[string]interface{}, user *model.PoliceUser) {
	if err := service.AuditReject(args, user); err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(nil))
	letterNo, _ := args["letter_no"].(string)
	service.AddOperationLog(user.ID, user.Name, user.PoliceNumber, "审核不通过", "核查工作台", letterNo, fmt.Sprintf("核查驳回，编号:%s", letterNo))
}

func handleFileUpload(c *gin.Context, user *model.PoliceUser) {
	letterNo := c.PostForm("letter_no")
	if letterNo == "" {
		c.JSON(http.StatusOK, model.ErrorResp("letter_no required"))
		return
	}
	fileType := c.PostForm("file_type")
	if fileType == "" {
		fileType = "call_recordings"
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorResp("file required: "+err.Error()))
		return
	}
	defer file.Close()

	mediaRoot := config.Get().Media.Root
	dir := filepath.Join(mediaRoot, "letters", letterNo, "recordings")
	if err := os.MkdirAll(dir, 0755); err != nil {
		c.JSON(http.StatusOK, model.ErrorResp("create dir error: "+err.Error()))
		return
	}

	filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), header.Filename)
	savePath := filepath.Join(dir, filename)
	if err := c.SaveUploadedFile(header, savePath); err != nil {
		c.JSON(http.StatusOK, model.ErrorResp("save file error: "+err.Error()))
		return
	}

	urlPath := fmt.Sprintf("/media/letters/%s/recordings/%s", letterNo, filename)

	// 将附件信息写入 letter_attachments 表
	if err := service.AppendAttachment(letterNo, fileType, urlPath, header.Filename); err != nil {
		c.JSON(http.StatusOK, model.ErrorResp("save attachment record error: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.SuccessResp(map[string]interface{}{
		"file_url":  urlPath,
		"file_name": header.Filename,
		"file_type": fileType,
		"letter_no": letterNo,
	}))
}

func handleExport(c *gin.Context, args map[string]interface{}, user *model.PoliceUser) {
	filePath, err := service.ExportLetters(string(user.PermissionLevel), user.UnitID, user.ID, args)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.FileAttachment(filePath, filepath.Base(filePath))
}

func handleSetSpecialFocus(c *gin.Context, args map[string]interface{}) {
	if err := service.SetLetterSpecialFocus(args); err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(nil))
}

func handleGetLetterSpecialFocus(c *gin.Context, args map[string]interface{}) {
	letterNo, ok := args["letter_no"].(string)
	if !ok || letterNo == "" {
		c.JSON(http.StatusOK, model.ErrorResp("letter_no required"))
		return
	}
	focusID, focusName, err := service.GetLetterSpecialFocus(letterNo)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(map[string]interface{}{
		"focus_id":   focusID,
		"focus_name": focusName,
	}))
}

func handleExportMonthlyReport(c *gin.Context, args map[string]interface{}, user *model.PoliceUser) {
	period, _ := args["period"].(string)

	var filePath string
	var err error

	if period == "all" || period == "" {
		// 全部：直接导出信件数据，与 LettersView 导出一致
		filePath, err = service.ExportLetters(string(user.PermissionLevel), user.UnitID, user.ID, args)
	} else {
		filePath, err = service.ExportMonthlyReport(string(user.PermissionLevel), user.UnitID, period)
	}

	if err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.FileAttachment(filePath, filepath.Base(filePath))
	os.RemoveAll(filepath.Dir(filePath))
}
