package service

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"letter-manage-backend/config"
	"letter-manage-backend/dao"
)

type LLMMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type LLMRequest struct {
	Model    string       `json:"model"`
	Messages []LLMMessage `json:"messages"`
	Stream   bool         `json:"stream"`
}

type LLMChoice struct {
	Message      *LLMMessage `json:"message,omitempty"`
	Delta        *LLMMessage `json:"delta,omitempty"`
	FinishReason string      `json:"finish_reason"`
}

type LLMResponse struct {
	ID      string      `json:"id"`
	Choices []LLMChoice `json:"choices"`
}

func buildHTTPClient(timeoutSec int) *http.Client {
	return &http.Client{Timeout: time.Duration(timeoutSec) * time.Second}
}

func Chat(messages []LLMMessage) (string, error) {
	cfg := config.Get().LLM
	reqBody := LLMRequest{
		Model:    cfg.Model,
		Messages: messages,
		Stream:   false,
	}
	b, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", cfg.APIURL, bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)

	client := buildHTTPClient(cfg.Timeout)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("LLM API error %d: %s", resp.StatusCode, string(body))
	}

	var llmResp LLMResponse
	if err := json.Unmarshal(body, &llmResp); err != nil {
		return "", err
	}
	if len(llmResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}
	if llmResp.Choices[0].Message == nil {
		return "", fmt.Errorf("empty message in response")
	}
	return llmResp.Choices[0].Message.Content, nil
}

// ChatStream calls LLM in streaming mode and writes SSE to the writer
func ChatStream(messages []LLMMessage, w io.Writer, flusher http.Flusher) error {
	cfg := config.Get().LLM
	reqBody := LLMRequest{
		Model:    cfg.Model,
		Messages: messages,
		Stream:   true,
	}
	b, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", cfg.APIURL, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)

	client := &http.Client{Timeout: time.Duration(cfg.Timeout) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("LLM API error %d: %s", resp.StatusCode, string(body))
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		fmt.Fprintf(w, "%s\n\n", line)
		if flusher != nil {
			flusher.Flush()
		}
		if line == "data: [DONE]" {
			break
		}
	}
	return scanner.Err()
}

func GetPrompt(promptType string) (string, error) {
	p, err := dao.GetPromptByType(promptType)
	if err != nil {
		return "", err
	}
	return p.Content, nil
}
