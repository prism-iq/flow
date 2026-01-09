package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"flow/pkg/logger"
)

type RAGService struct {
	baseURL string
	client  *http.Client
	log     *logger.Logger
}

type RAGQueryRequest struct {
	Query   string `json:"query"`
	TopK    int    `json:"top_k"`
	Filters string `json:"filters,omitempty"`
}

type RAGDocument struct {
	ID       string            `json:"id"`
	Content  string            `json:"content"`
	Score    float64           `json:"score"`
	Metadata map[string]string `json:"metadata"`
}

type RAGQueryResponse struct {
	Documents []RAGDocument `json:"documents"`
	QueryTime float64       `json:"query_time_ms"`
}

func NewRAGService(baseURL string, log *logger.Logger) *RAGService {
	return &RAGService{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		log: log.WithComponent("rag-service"),
	}
}

func (s *RAGService) Query(ctx context.Context, query string, topK int) (string, error) {
	req := RAGQueryRequest{
		Query: query,
		TopK:  topK,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/query", bytes.NewReader(body))
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
		return "", fmt.Errorf("rag service returned %d", resp.StatusCode)
	}

	var ragResp RAGQueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&ragResp); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	return s.formatDocuments(ragResp.Documents), nil
}

func (s *RAGService) formatDocuments(docs []RAGDocument) string {
	var sb strings.Builder
	for i, doc := range docs {
		sb.WriteString(fmt.Sprintf("[Document %d] (score: %.3f)\n%s\n\n", i+1, doc.Score, doc.Content))
	}
	return sb.String()
}
