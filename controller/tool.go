package controller

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"

	"letter-manage-backend/config"
	"letter-manage-backend/model"
	"letter-manage-backend/service"

	"github.com/gin-gonic/gin"
	ws "github.com/gorilla/websocket"
)

func ToolController(c *gin.Context) {
	var req model.APIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResp("invalid request"))
		return
	}
	dispatchTool(c, req.Order, req.Args)
}

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

func parseToolArgs(c *gin.Context) map[string]interface{} {
	var req model.APIRequest
	if err := c.ShouldBindJSON(&req); err == nil && req.Args != nil {
		return req.Args
	}
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
	ch, errCh := service.TranscribeStream(absPath)
	var texts []string
	for chunk := range ch {
		if chunk.Status != "" || chunk.Done {
			continue
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
	c.JSON(http.StatusOK, model.SuccessResp(strings.Join(texts, "\n")))
}

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

	emitSSE(c.Writer, flusher, "status", "正在上传音频并启动识别...")

	ch, errCh := service.TranscribeStream(absPath)
	var emittedText string
	emitted := make(map[string]bool)

	for {
		select {
		case chunk, ok := <-ch:
			if !ok {
				select {
				case e := <-errCh:
					if e != nil {
						emitSSE(c.Writer, flusher, "error", e.Error())
						return
					}
				default:
				}
				emitSSE(c.Writer, flusher, "done", emittedText)
				return
			}
			if chunk.Status != "" {
				emitSSE(c.Writer, flusher, "status", chunk.Status)
				continue
			}
			if chunk.Done {
				emitSSE(c.Writer, flusher, "done", emittedText)
				return
			}
			for _, line := range strings.Split(chunk.Text, "\n") {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				if strings.Contains(line, "实时转译中") ||
					strings.Contains(line, "总时长") ||
					strings.Contains(line, "转译完成") ||
					strings.Contains(line, "AI 纠错") ||
					strings.ReplaceAll(line, "─", "") == "" {
					continue
				}
				cleanLine := regexp.MustCompile(`^\[\d+\]\s*`).ReplaceAllString(line, "")
				cleanLine = leadingNoise.ReplaceAllString(cleanLine, "")
				cleanLine = strings.TrimSpace(cleanLine)
				if cleanLine == "" || emitted[cleanLine] {
					continue
				}
				emitted[cleanLine] = true
				emittedText += cleanLine
				emitSSE(c.Writer, flusher, "chunk", cleanLine)
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
)

// decodeWAVToFloat32 解析 WAV 文件头，提取 PCM 数据转为 float32
// 返回小端序 float32 字节（Python 端可直接 np.frombuffer）
func decodeWAVToFloat32(data []byte) ([]byte, error) {
	if len(data) < 44 || string(data[0:4]) != "RIFF" {
		return nil, fmt.Errorf("not a WAV file")
	}
	// 读取 fmt chunk
	audioFormat := binary.LittleEndian.Uint16(data[20:22])
	numChannels := binary.LittleEndian.Uint16(data[22:24])
	sampleRate := binary.LittleEndian.Uint32(data[24:28])
	bitsPerSample := binary.LittleEndian.Uint16(data[34:36])

	if audioFormat != 1 { // PCM only
		return nil, fmt.Errorf("unsupported audio format: %d", audioFormat)
	}

	// 查找 data chunk (跳过可能的额外 chunk)
	offset := 36
	for offset < len(data)-8 {
		chunkID := string(data[offset : offset+4])
		chunkSize := binary.LittleEndian.Uint32(data[offset+4 : offset+8])
		if chunkID == "data" {
			pcmStart := offset + 8
			pcmEnd := pcmStart + int(chunkSize)
			if pcmEnd > len(data) {
				pcmEnd = len(data)
			}
			pcm := data[pcmStart:pcmEnd]

			// 转为 float32
			numSamples := len(pcm) / int(bitsPerSample/8)
			floatData := make([]byte, numSamples*4)
			for i := 0; i < numSamples; i++ {
				var sample float32
				switch bitsPerSample {
				case 16:
					s := int16(binary.LittleEndian.Uint16(pcm[i*2 : i*2+2]))
					sample = float32(s) / 32768.0
				case 32:
					s := int32(binary.LittleEndian.Uint32(pcm[i*4 : i*4+4]))
					sample = float32(s) / 2147483648.0
				default:
					return nil, fmt.Errorf("unsupported bits per sample: %d", bitsPerSample)
				}
				binary.LittleEndian.PutUint32(floatData[i*4:i*4+4], math.Float32bits(sample))
			}

			// 转为单声道（取平均）如果需要
			if numChannels > 1 {
				mono := make([]byte, numSamples/int(numChannels)*4)
				for i := 0; i < len(mono)/4; i++ {
					var sum float32
					for ch := uint16(0); ch < numChannels; ch++ {
						idx := i*int(numChannels) + int(ch)
						sum += math.Float32frombits(binary.LittleEndian.Uint32(floatData[idx*4 : idx*4+4]))
					}
					binary.LittleEndian.PutUint32(mono[i*4:i*4+4], math.Float32bits(sum/float32(numChannels)))
				}
				floatData = mono
			}

			// 重采样到 16kHz（简单降采样，丢帧）
			if sampleRate != 16000 {
				ratio := float64(sampleRate) / 16000
				srcSamples := len(floatData) / 4
				dstSamples := int(float64(srcSamples) / ratio)
				resampled := make([]byte, dstSamples*4)
				for i := 0; i < dstSamples; i++ {
					srcIdx := int(float64(i) * ratio)
					if srcIdx*4+4 <= len(floatData) {
						copy(resampled[i*4:i*4+4], floatData[srcIdx*4:srcIdx*4+4])
					}
				}
				floatData = resampled
			}

			_ = numChannels // used above
			return floatData, nil
		}
		offset += 8 + int(chunkSize)
	}
	return nil, fmt.Errorf("data chunk not found")
}

// ToolTranscribeWS — WebSocket 转译代理 (复用现有 Gradio)
// 
// 每个 Gradio SSE 事件包含完整的当前输出状态（所有已识别行）。
// 因此每次新事件到达时直接替换当前文本，而非累积追加。
// SenseVoice 侧的纠错完成后，HTML 中的行会从 "pending" 
// 变为 "corrected"（带 diff 标记），提取逻辑在 gradio.go 中处理。
func ToolTranscribeWS(c *gin.Context) {
	upgrader := ws.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	clientConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer clientConn.Close()

	_, msg, err := clientConn.ReadMessage()
	if err != nil {
		return
	}
	var req struct{ AudioURL string `json:"audio_url"` }
	json.Unmarshal(msg, &req)
	if req.AudioURL == "" {
		clientConn.WriteMessage(ws.TextMessage, []byte(`{"type":"error","message":"missing audio_url"}`))
		return
	}

	absPath, err := resolveAudioPath(req.AudioURL)
	if err != nil {
		clientConn.WriteMessage(ws.TextMessage, []byte(`{"type":"error","message":"`+err.Error()+`"}`))
		return
	}

	ch, errCh := service.TranscribeStream(absPath)
	sendJSON(clientConn, map[string]string{"type": "status", "message": "🎙️ 开始转译..."})

	// currentText 存放最新快照的文本（每次 SSE 事件替换，不累积）
	currentText := ""
	lastSent := ""
	bannerRe := regexp.MustCompile(`^[─═\-]{3,}$`)
	statusRe := regexp.MustCompile(`(实时转译|总时长|AI 纠错|转译完成|正在分析|检测到)`)

	for chunk := range ch {
		if chunk.Status != "" {
			sendJSON(clientConn, map[string]string{"type": "status", "message": chunk.Status})
			continue
		}
		if chunk.Done {
			break
		}

		isFinal := strings.Contains(chunk.Text, "转译完成")

		// 提取干净文本行（过滤状态/横幅/空行）
		var cleanLines []string
		for _, line := range strings.Split(chunk.Text, "\n") {
			line = strings.TrimSpace(line)
			if line == "" || statusRe.MatchString(line) || bannerRe.MatchString(line) {
				continue
			}
			// 去掉 VAD 段号标记 "[1] " 等
			line = regexp.MustCompile(`^\[\d+\]\s*`).ReplaceAllString(line, "")
			line = leadingNoise.ReplaceAllString(line, "")
			line = strings.TrimSpace(line)
			if line != "" {
				cleanLines = append(cleanLines, line)
			}
		}

		if len(cleanLines) == 0 {
			continue
		}

		// 每个 SSE 事件是完整快照 → 替换而非累积
		// 这样确保：同一个逻辑行的 pending → corrected 过渡时，
		// 只有最新的版本保留，不会出现 raw 和 corrected 同时存在
		snapshot := strings.Join(cleanLines, "")
		currentText = snapshot

		if isFinal {
			continue
		}

		if snapshot != lastSent {
			lastSent = snapshot
			sendJSON(clientConn, map[string]interface{}{
				"type": "partial", "text": snapshot,
			})
		}
	}

	// 检查错误
	select {
	case e := <-errCh:
		if e != nil {
			sendJSON(clientConn, map[string]string{"type": "error", "message": e.Error()})
			return
		}
	default:
	}

	sendJSON(clientConn, map[string]interface{}{
		"type": "final", "text": currentText,
	})
	sendJSON(clientConn, map[string]string{"type": "done"})
}

func sendJSON(conn *ws.Conn, v interface{}) {
	data, _ := json.Marshal(v)
	conn.WriteMessage(ws.TextMessage, data)
}
