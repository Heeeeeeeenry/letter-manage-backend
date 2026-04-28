package controller

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"letter-manage-backend/config"
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
		handleCreateLetter(c, req.Args)
	case "update":
		handleUpdateLetter(c, req.Args)
	case "delete":
		handleDeleteLetter(c, req.Args)
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
	case "return_letter":
		handleReturnLetter(c, req.Args, user)
	case "audit_approve":
		handleAuditApprove(c, req.Args, user)
	case "audit_reject":
		handleAuditReject(c, req.Args, user)
	default:
		c.JSON(http.StatusBadRequest, model.ErrorResp("unknown order: "+req.Order))
	}
}

func handleGetList(c *gin.Context, args map[string]interface{}, user *model.PoliceUser) {
	data, err := service.GetLetterList(args, user.UnitName, string(user.PermissionLevel))
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(data))
}

func handleGetDispatchList(c *gin.Context, args map[string]interface{}, user *model.PoliceUser) {
	data, err := service.GetDispatchList(user.UnitName, string(user.PermissionLevel), args)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(data))
}

func handleGetProcessingList(c *gin.Context, args map[string]interface{}, user *model.PoliceUser) {
	data, err := service.GetProcessingList(user.UnitName, string(user.PermissionLevel), args)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(data))
}

func handleGetAuditList(c *gin.Context, args map[string]interface{}, user *model.PoliceUser) {
	data, err := service.GetAuditList(user.UnitName, string(user.PermissionLevel), args)
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
	data, err := service.GetLetterDetail(letterNo, user.UnitName, string(user.PermissionLevel))
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
	letters, err := service.GetLettersByPhone(phone, user.UnitName, string(user.PermissionLevel))
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
	letters, err := service.GetLettersByIDCard(idCard, user.UnitName, string(user.PermissionLevel))
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(letters))
}

func handleCreateLetter(c *gin.Context, args map[string]interface{}) {
	letter, err := service.CreateLetter(args)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(letter))
}

func handleUpdateLetter(c *gin.Context, args map[string]interface{}) {
	if err := service.UpdateLetter(args); err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(nil))
}

func handleDeleteLetter(c *gin.Context, args map[string]interface{}) {
	if err := service.DeleteLetter(args); err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(nil))
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
	data, err := service.GetStatistics(user.UnitName, string(user.PermissionLevel), period, user.UnitID)
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
}

func handleMarkInvalid(c *gin.Context, args map[string]interface{}, user *model.PoliceUser) {
	if err := service.MarkInvalid(args, user); err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(nil))
}

func handleSubmitProcessing(c *gin.Context, args map[string]interface{}, user *model.PoliceUser) {
	if err := service.SubmitProcessing(args, user); err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(nil))
}

func handleHandleBySelf(c *gin.Context, args map[string]interface{}, user *model.PoliceUser) {
	if err := service.HandleBySelf(args, user); err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(nil))
}

func handleReturnLetter(c *gin.Context, args map[string]interface{}, user *model.PoliceUser) {
	if err := service.ReturnLetter(args, user); err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(nil))
}

func handleAuditApprove(c *gin.Context, args map[string]interface{}, user *model.PoliceUser) {
	if err := service.AuditApprove(args, user); err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(nil))
}

func handleAuditReject(c *gin.Context, args map[string]interface{}, user *model.PoliceUser) {
	if err := service.AuditReject(args, user); err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(nil))
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
	c.JSON(http.StatusOK, model.SuccessResp(map[string]interface{}{
		"file_url":  urlPath,
		"file_name": header.Filename,
		"file_type": fileType,
		"letter_no": letterNo,
	}))
}
