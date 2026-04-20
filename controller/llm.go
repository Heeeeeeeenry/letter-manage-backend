package controller

import (
	"encoding/json"
	"net/http"

	"letter-manage-backend/middleware"
	"letter-manage-backend/model"
	"letter-manage-backend/service"

	"github.com/gin-gonic/gin"
)

// LLMController handles /api/llm/
func LLMController(c *gin.Context) {
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
	case "chat":
		handleChat(c, req.Args)
	case "chat_stream":
		handleChatStream(c, req.Args)
	case "get_prompt":
		handleGetPrompt(c, req.Args)
	default:
		c.JSON(http.StatusBadRequest, model.ErrorResp("unknown order: "+req.Order))
	}
}

func handleChat(c *gin.Context, args map[string]interface{}) {
	messagesRaw, ok := args["messages"]
	if !ok {
		c.JSON(http.StatusOK, model.ErrorResp("messages required"))
		return
	}
	b, err := json.Marshal(messagesRaw)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorResp("invalid messages"))
		return
	}
	var messages []service.LLMMessage
	if err := json.Unmarshal(b, &messages); err != nil {
		c.JSON(http.StatusOK, model.ErrorResp("invalid messages format"))
		return
	}
	result, err := service.Chat(messages)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(map[string]interface{}{
		"content": result,
	}))
}

func handleChatStream(c *gin.Context, args map[string]interface{}) {
	messagesRaw, ok := args["messages"]
	if !ok {
		c.JSON(http.StatusOK, model.ErrorResp("messages required"))
		return
	}
	b, err := json.Marshal(messagesRaw)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorResp("invalid messages"))
		return
	}
	var messages []service.LLMMessage
	if err := json.Unmarshal(b, &messages); err != nil {
		c.JSON(http.StatusOK, model.ErrorResp("invalid messages format"))
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	flusher, hasFlusher := c.Writer.(http.Flusher)
	if !hasFlusher {
		c.JSON(http.StatusInternalServerError, model.ErrorResp("streaming not supported"))
		return
	}

	if err := service.ChatStream(messages, c.Writer, flusher); err != nil {
		// Stream already started, write error as SSE event
		c.Writer.Write([]byte("data: [ERROR] " + err.Error() + "\n\n"))
		flusher.Flush()
	}
}

func handleGetPrompt(c *gin.Context, args map[string]interface{}) {
	promptType, ok := args["prompt_type"].(string)
	if !ok || promptType == "" {
		c.JSON(http.StatusOK, model.ErrorResp("prompt_type required"))
		return
	}
	content, err := service.GetPrompt(promptType)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorResp(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(map[string]interface{}{
		"content":     content,
		"prompt_type": promptType,
	}))
}
