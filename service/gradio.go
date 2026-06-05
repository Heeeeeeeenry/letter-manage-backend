package service

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"letter-manage-backend/config"
)

// gradioURL returns the configured Gradio base URL
func gradioURL() string {
	if url := config.Get().Gradio.BaseURL; url != "" {
		return url
	}
	return "http://localhost:7860"
}

// TranscribeChunk is a piece of transcription text from Gradio
type TranscribeChunk struct {
	Text   string
	Done   bool
	Status string // non-empty = progress/status update (not transcription text)
}

// TranscribeStream calls Gradio SenseVoice API and returns a channel of text chunks
func TranscribeStream(audioPath string) (<-chan TranscribeChunk, <-chan error) {
	ch := make(chan TranscribeChunk, 10)
	errCh := make(chan error, 1)

	go func() {
		defer close(ch)
		defer close(errCh)

		session := fmt.Sprintf("go_%d", time.Now().UnixNano())

		// Step 1: Upload file
		ch <- TranscribeChunk{Status: "uploading"}
		cachedPath, err := gradioUploadFile(audioPath)
		if err != nil {
			errCh <- fmt.Errorf("gradio upload: %w", err)
			return
		}

		// Step 2: Submit job (queues the job, Gradio starts processing)
		ch <- TranscribeChunk{Status: "processing"}
		if _, err := gradioSubmitWithPath(cachedPath, session); err != nil {
			errCh <- fmt.Errorf("gradio submit: %w", err)
			return
		}

		// Step 3: Connect to SSE queue (Gradio buffers events, won't miss them)
		ch <- TranscribeChunk{Status: "transcribing"}
		if err := gradioStreamEvents(session, ch); err != nil {
			errCh <- err
		}
	}()

	return ch, errCh
}

// gradioSubmitWithPath submits the transcribe job with an already-uploaded cached file path
func gradioSubmitWithPath(cachedPath, session string) (string, error) {
	body := map[string]interface{}{
		"data": []interface{}{
			map[string]interface{}{
				"path": cachedPath,
				"meta": map[string]string{"_type": "gradio.FileData"},
			},
			"auto",
		},
		"event_data":   nil,
		"session_hash": session,
	}
	b, _ := json.Marshal(body)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(
		gradioURL()+"/gradio_api/call/stream_inference",
		"application/json",
		strings.NewReader(string(b)),
	)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		EventID string `json:"event_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode gradio response: %w", err)
	}
	if result.EventID == "" {
		return "", fmt.Errorf("no event_id from Gradio (status %d)", resp.StatusCode)
	}
	return result.EventID, nil
}

// gradioUploadFile uploads a local file to Gradio's cache and returns the cached path
func gradioUploadFile(localPath string) (string, error) {
	file, err := os.Open(localPath)
	if err != nil {
		return "", fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("files", filepath.Base(localPath))
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(part, file); err != nil {
		return "", err
	}
	writer.Close()

	uploadID := fmt.Sprintf("upload_%d", time.Now().UnixNano())
	url := fmt.Sprintf("%s/gradio_api/upload?upload_id=%s", gradioURL(), uploadID)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post(url, writer.FormDataContentType(), body)
	if err != nil {
		return "", fmt.Errorf("upload request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("upload HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	// Response is a JSON array of cached paths
	var paths []string
	if err := json.NewDecoder(resp.Body).Decode(&paths); err != nil {
		return "", fmt.Errorf("decode upload response: %w", err)
	}
	if len(paths) == 0 {
		return "", fmt.Errorf("upload returned no paths")
	}
	return paths[0], nil
}

// gradioStreamEvents opens SSE connection and parses events into the channel.
// This is a blocking call — call it in a goroutine before submitting the job.
func gradioStreamEvents(session string, ch chan<- TranscribeChunk) error {
	url := fmt.Sprintf("%s/gradio_api/queue/data?session_hash=%s", gradioURL(), session)
	client := &http.Client{
		Transport: &http.Transport{
			ResponseHeaderTimeout: 10 * time.Second,
		},
		Timeout: 0,
	}

	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("gradio queue connect: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("gradio queue HTTP %d: %s", resp.StatusCode, string(body)[:200])
	}

	return parseGradioSSE(resp.Body, ch)
}

// parseGradioSSE reads Gradio's SSE event stream and extracts transcription text
func parseGradioSSE(reader io.Reader, ch chan<- TranscribeChunk) error {
	scanner := bufio.NewScanner(reader)
	// Large buffer: Gradio HTML snapshots can be > 1MB for long transcriptions
	scanner.Buffer(make([]byte, 128*1024), 16*1024*1024)

	var buf strings.Builder
	for scanner.Scan() {
		line := scanner.Text()

		// Accumulate lines until we hit an empty line (SSE event boundary)
		if line != "" {
			buf.WriteString(line)
			buf.WriteByte('\n')
			continue
		}

		// Empty line = end of event, parse accumulated data
		if buf.Len() > 0 {
			msg := buf.String()
			buf.Reset()
			text, isError := processSSEEvent(msg)
			if isError {
				return fmt.Errorf("gradio transcribe failed: %s", text)
			}
			if text != "" {
				ch <- TranscribeChunk{Text: text}
			}
			if isComplete(msg) {
				ch <- TranscribeChunk{Done: true}
				return nil
			}
		}
	}
	return scanner.Err()
}

// processSSEEvent parses a Gradio SSE data event and extracts text.
// Returns (text, isError) — isError=true if the event indicates a Gradio error.
func processSSEEvent(msg string) (string, bool) {
	lines := strings.Split(strings.TrimSpace(msg), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		var event map[string]interface{}
		if err := json.Unmarshal([]byte(line[6:]), &event); err != nil {
			continue
		}

		msgType, _ := event["msg"].(string)
		if msgType == "process_completed" || msgType == "complete" {
			// Check for error
			output, _ := event["output"].(map[string]interface{})
			if output != nil {
				if errMsg, ok := output["error"].(string); ok && errMsg != "" {
					return errMsg, true
				}
			}
		}
		if msgType != "process_generating" && msgType != "process_completed" {
			continue
		}

		// Extract text from output.data array (HTML format)
		output, _ := event["output"].(map[string]interface{})
		if output == nil {
			continue
		}
		dataArr, _ := output["data"].([]interface{})
		return extractTextFromHTML(dataArr), false
	}
	return "", false
}

// isComplete checks if the SSE event indicates job completion
func isComplete(msg string) bool {
	return strings.Contains(msg, `"msg":"process_completed"`) ||
		strings.Contains(msg, `"msg":"complete"`)
}

var (
	tagRe        = regexp.MustCompile(`<[^>]+>`)
	cursorRe     = regexp.MustCompile(`<span class="cursor">[|▌]</span>`)
	hasCJK       = regexp.MustCompile(`[\p{Han}]`)
	stripEmoji   = regexp.MustCompile(`[🎼😊😅😂🤣❤️🔥💯✨]+`)
	leadingNoise = regexp.MustCompile(`^[^\p{Han}]+`)
	htmlEntityRe = regexp.MustCompile(`&[a-z]+;`)
	// Gradio 6 textarea: <textarea ...>text</textarea>
	textareaRe = regexp.MustCompile(`<textarea[^>]*>(.*?)</textarea>`)
)

// extractTextFromHTML extracts readable text from Gradio's HTML output.
// Tries multiple formats: line divs, textarea, then plain text fallback.
func extractTextFromHTML(dataArr []interface{}) string {
	var lines []string
	seen := make(map[string]bool)

	for _, item := range dataArr {
		html, ok := item.(string)
		if !ok {
			continue
		}

		// Remove cursor spans
		html = cursorRe.ReplaceAllString(html, "")

		// Try 1: Extract from <div class="line ...">text</div>
		lineRe := regexp.MustCompile(`<div class="line[^"]*">(.*?)</div>`)
		matches := lineRe.FindAllStringSubmatch(html, -1)
		for _, m := range matches {
			text := cleanText(m[1])
			if text != "" && !seen[text] && hasCJK.MatchString(text) {
				seen[text] = true
				lines = append(lines, text)
			}
		}
		if len(matches) > 0 {
			continue
		}

		// Try 2: Extract from <textarea>text</textarea> (Gradio 6 Textbox)
		taMatches := textareaRe.FindAllStringSubmatch(html, -1)
		for _, m := range taMatches {
			text := strings.TrimSpace(m[1])
			if text != "" {
				for _, t := range strings.Split(text, "\n") {
					t = cleanText(t)
					if t != "" && !seen[t] && hasCJK.MatchString(t) {
						seen[t] = true
						lines = append(lines, t)
					}
				}
			}
		}
		if len(taMatches) > 0 {
			continue
		}

		// Try 3: Plain text fallback — strip all tags
		text := tagRe.ReplaceAllString(html, "")
		text = htmlEntityRe.ReplaceAllString(text, "")
		text = strings.TrimSpace(text)
		if text != "" {
			for _, t := range strings.Split(text, "\n") {
				t = cleanText(t)
				if t != "" && !seen[t] && hasCJK.MatchString(t) {
					seen[t] = true
					lines = append(lines, t)
				}
			}
		}
	}

	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n")
}

func cleanText(s string) string {
	s = tagRe.ReplaceAllString(s, "")
	s = stripEmoji.ReplaceAllString(s, "")
	s = leadingNoise.ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}
