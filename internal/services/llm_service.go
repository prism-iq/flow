package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"flow/pkg/logger"
)

type LLMService struct {
	baseURL string
	client  *http.Client
	log     *logger.Logger
}

type LLMRequest struct {
	Prompt    string `json:"prompt"`
	MaxTokens int    `json:"max_tokens"`
	Context   string `json:"context,omitempty"`
	Stream    bool   `json:"stream,omitempty"`
}

type LLMResponse struct {
	Text       string `json:"text"`
	TokenCount int    `json:"token_count"`
}

func NewLLMService(baseURL string, log *logger.Logger) *LLMService {
	return &LLMService{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
		log: log.WithComponent("llm-service"),
	}
}

func (s *LLMService) Generate(ctx context.Context, prompt, contextDocs string) (string, error) {
	req := LLMRequest{
		Prompt:    s.buildPrompt(prompt, contextDocs),
		MaxTokens: 2048,
		Stream:    false,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/generate", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("llm service returned %d", resp.StatusCode)
	}

	var llmResp LLMResponse
	if err := json.NewDecoder(resp.Body).Decode(&llmResp); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	return llmResp.Text, nil
}

func (s *LLMService) GenerateStream(ctx context.Context, prompt, contextDocs string, tokenChan chan<- string) error {
	req := LLMRequest{
		Prompt:    s.buildPrompt(prompt, contextDocs),
		MaxTokens: 2048,
		Stream:    true,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/generate/stream", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			tokenChan <- string(buf[:n])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read stream: %w", err)
		}
	}

	return nil
}

func (s *LLMService) buildPrompt(userPrompt, context string) string {
	if context == "" {
		return userPrompt
	}
	return fmt.Sprintf("Context:\n%s\n\nQuery: %s", context, userPrompt)
}
