package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"letter-manage-backend/config"
)

func sttURL() string {
	if url := config.Get().Gradio.BaseURL; url != "" {
		return url
	}
	return "http://stt:7860"
}

type TranscribeChunk struct {
	Text   string
	Done   bool
	Status string
}

func TranscribeStream(audioPath string) (<-chan TranscribeChunk, <-chan error) {
	ch := make(chan TranscribeChunk, 10)
	errCh := make(chan error, 1)

	go func() {
		defer close(ch)
		defer close(errCh)

		ch <- TranscribeChunk{Status: "uploading"}
		text, err := transcribe(audioPath)
		if err != nil {
			errCh <- err
			return
		}

		ch <- TranscribeChunk{Status: "transcribing"}
		for _, line := range strings.Split(text, "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				ch <- TranscribeChunk{Text: line}
				time.Sleep(50 * time.Millisecond)
			}
		}
		ch <- TranscribeChunk{Done: true}
	}()

	return ch, errCh
}

func transcribe(audioPath string) (string, error) {
	// Upload file to Gradio
	cachedPath, err := uploadToGradio(audioPath)
	if err != nil {
		return "", fmt.Errorf("upload: %w", err)
	}

	// Call /gradio_api/run/stream_inference (sync, non-generator function)
	body := map[string]interface{}{
		"data": []interface{}{
			map[string]interface{}{
				"path": cachedPath,
				"meta": map[string]string{"_type": "gradio.FileData"},
			},
			"auto",
		},
	}
	b, _ := json.Marshal(body)

	client := &http.Client{Timeout: 300 * time.Second}
	resp, err := client.Post(
		sttURL()+"/gradio_api/run/stream_inference",
		"application/json",
		bytes.NewReader(b),
	)
	if err != nil {
		return "", fmt.Errorf("run: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody)[:300])
	}

	var result struct {
		Data []string `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("decode: %w", err)
	}
	if len(result.Data) == 0 {
		return "", fmt.Errorf("empty result")
	}
	return result.Data[0], nil
}

func uploadToGradio(audioPath string) (string, error) {
	file, err := os.Open(audioPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("files", filepath.Base(audioPath))
	if err != nil {
		return "", err
	}
	io.Copy(part, file)
	writer.Close()

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post(
		sttURL()+"/gradio_api/upload",
		writer.FormDataContentType(),
		body,
	)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("upload HTTP %d: %s", resp.StatusCode, string(respBody)[:200])
	}
	var paths []string
	if err := json.Unmarshal(respBody, &paths); err != nil {
		return "", err
	}
	if len(paths) == 0 {
		return "", fmt.Errorf("no cached path")
	}
	return paths[0], nil
}
