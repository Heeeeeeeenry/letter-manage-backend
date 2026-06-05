package controller

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"time"

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

// ToolTranscribe 音频转文字 (Go 直连 Gradio SenseVoice API)
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

	// 调用 Go service 直连 Gradio，收集所有 chunk 拼接
	ch, errCh := service.TranscribeStream(absPath)
	var texts []string
	for chunk := range ch {
		if chunk.Status != "" {
			continue
		}
		if chunk.Done {
			break
		}
		texts = append(texts, chunk.Text)
	}
	select {
	case e := <-errCh:
		if e != nil {
			c.JSON(http.StatusInternalServerError, model.ErrorResp("转写失败: "+e.Error()))
			return
		}
	default:
	}
	c.JSON(http.StatusOK, model.SuccessResp(strings.Join(texts, "\n\n")))
}

// ToolTranscribeStream 流式音频转文字 (SSE) — Go 直连 Gradio SenseVoice API
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

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, model.ErrorResp("不支持 SSE"))
		return
	}

	// Send immediate status so client knows we're processing
	emitSSE(c.Writer, flusher, "status", "正在上传音频并启动识别...")

	ch, errCh := service.TranscribeStream(absPath)
	var fullText string
	seen := make(map[string]bool) // dedup sentences across Gradio's repeated outputs

	for {
		select {
		case chunk, ok := <-ch:
			if !ok {
				// Channel closed — check for error
				select {
				case e := <-errCh:
					if e != nil {
						emitSSE(c.Writer, flusher, "error", e.Error())
						return
					}
				default:
				}
				emitSSE(c.Writer, flusher, "done", fullText)
				return
			}
			if chunk.Status != "" {
				emitSSE(c.Writer, flusher, "status", chunk.Status)
				continue
			}
			if chunk.Done {
				emitSSE(c.Writer, flusher, "done", fullText)
				return
			}
			fullText = chunk.Text // Gradio sends cumulative text each time
			// Simulate line-by-line streaming: Gradio outputs all text at once
			// (model.generate() is batch), so split into sentences and stream progressively
			for _, rawLine := range strings.Split(chunk.Text, "\n") {
				rawLine = strings.TrimSpace(rawLine)
				if rawLine == "" {
					continue
				}
				// Skip Gradio formatting lines (header/footer/dividers)
				if strings.Contains(rawLine, "实时转译中") ||
					strings.Contains(rawLine, "总时长") ||
					strings.Contains(rawLine, "转译完成") ||
					strings.ReplaceAll(rawLine, "─", "") == "" {
					continue
				}
				// Clean segment markers like "[1] " and leading emoji/noise
				cleanLine := regexp.MustCompile(`^\[\d+\]\s*`).ReplaceAllString(rawLine, "")
				cleanLine = leadingNoise.ReplaceAllString(cleanLine, "")
				cleanLine = strings.TrimSpace(cleanLine)
				if cleanLine == "" {
					continue
				}
				// Split into sentences for progressive streaming effect
				sentences := sentenceSplit.ReplaceAllString(cleanLine, "$0\n")
				for _, s := range strings.Split(sentences, "\n") {
					s = strings.TrimSpace(s)
					if s != "" && !seen[s] {
						seen[s] = true
						emitSSE(c.Writer, flusher, "chunk", s)
						time.Sleep(60 * time.Millisecond)
					}
				}
			}
		case e := <-errCh:
			if e != nil {
				emitSSE(c.Writer, flusher, "error", e.Error())
			}
			return
		}
	}
}

func emitSSE(w io.Writer, flusher http.Flusher, event, data string) {
	fmt.Fprintf(w, "event: %s\n", event)
	for _, line := range strings.Split(data, "\n") {
		fmt.Fprintf(w, "data: %s\n", line)
	}
	fmt.Fprintf(w, "\n")
	flusher.Flush()
}

var (
	leadingNoise = regexp.MustCompile(`^[^\p{Han}]+`)
	sentenceSplit = regexp.MustCompile(`([。！？])`)
)
