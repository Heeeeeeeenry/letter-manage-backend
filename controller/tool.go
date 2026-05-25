package controller

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"letter-manage-backend/config"
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

// resolveAudioPath 将 audio_url 转为安全的本地绝对路径
func resolveAudioPath(audioURL string) (string, error) {
	cfg := config.Get()
	mediaRoot := cfg.Media.Root
	relativePath := audioURL[len("/media/"):]
	localPath := filepath.Join(mediaRoot, relativePath)

	absPath, _ := filepath.Abs(localPath)
	absRoot, _ := filepath.Abs(mediaRoot)
	if !strings.HasPrefix(absPath, absRoot) {
		return "", fmt.Errorf("非法的音频路径")
	}
	return absPath, nil
}

// ToolTranscribe 音频转文字 (通过 Python stdlib 脚本调用 Gradio SenseVoice API)
func ToolTranscribe(c *gin.Context) {
	var req struct {
		AudioURL string `json:"audio_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.AudioURL == "" {
		c.JSON(http.StatusBadRequest, model.ErrorResp("请提供 audio_url"))
		return
	}

	absPath, err := resolveAudioPath(req.AudioURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResp(err.Error()))
		return
	}

	// 调用 Python stdlib 脚本（无需任何 pip 依赖）
	cmd := exec.Command("python3", "scripts/gradio_call.py", absPath)
	out, err := cmd.Output()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResp("转写失败: "+err.Error()))
		return
	}

	// 解析最后一行 JSON 结果
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	lastLine := lines[len(lines)-1]
	var result struct {
		Done     bool   `json:"done"`
		FullText string `json:"full_text"`
		Error    string `json:"error"`
	}
	if err := json.Unmarshal([]byte(lastLine), &result); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResp("解析失败: "+err.Error()))
		return
	}
	if result.Error != "" {
		c.JSON(http.StatusInternalServerError, model.ErrorResp(result.Error))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResp(result.FullText))
}

// ToolTranscribeStream 流式音频转文字 (SSE) — Python stdlib 脚本调用 Gradio SenseVoice API
func ToolTranscribeStream(c *gin.Context) {
	var req struct {
		AudioURL string `json:"audio_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.AudioURL == "" {
		c.JSON(http.StatusBadRequest, model.ErrorResp("请提供 audio_url"))
		return
	}

	absPath, err := resolveAudioPath(req.AudioURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResp(err.Error()))
		return
	}

	// 设置 SSE 头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, model.ErrorResp("不支持 SSE"))
		return
	}

	// 调用 Python stdlib 脚本（无需 pip 依赖），实时读取 stdout
	cmd := exec.Command("python3", "-u", "scripts/gradio_call.py", absPath)
	cmd.Env = append(os.Environ(), "PYTHONUNBUFFERED=1")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		emitSSE(c.Writer, flusher, "error", fmt.Sprintf("创建管道失败: %v", err))
		return
	}

	if err := cmd.Start(); err != nil {
		emitSSE(c.Writer, flusher, "error", fmt.Sprintf("启动转写失败: %v", err))
		return
	}

	// 逐行读取 JSON 输出并转发为 SSE
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var msg struct {
			Text     string `json:"text"`
			Done     bool   `json:"done"`
			FullText string `json:"full_text"`
			Error    string `json:"error"`
		}
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			fmt.Printf("[transcribe] skip: %s\n", line)
			continue
		}
		fmt.Printf("[transcribe] ok: text=%s done=%v\n", msg.Text, msg.Done)

		if msg.Error != "" {
			emitSSE(c.Writer, flusher, "error", msg.Error)
			cmd.Wait()
			return
		}

		if msg.Done {
			emitSSE(c.Writer, flusher, "done", msg.FullText)
			break
		}

		if msg.Text != "" {
			emitSSE(c.Writer, flusher, "chunk", msg.Text)
		}
	}

	if err := scanner.Err(); err != nil {
		emitSSE(c.Writer, flusher, "error", fmt.Sprintf("读取失败: %v", err))
	}
	cmd.Wait()
}

func emitSSE(w io.Writer, flusher http.Flusher, event, data string) {
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data)
	flusher.Flush()
}
