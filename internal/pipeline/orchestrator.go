package pipeline

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"flow/internal/config"
	"flow/pkg/logger"
)

type WorkerType string

const (
	WorkerDate   WorkerType = "date"
	WorkerPerson WorkerType = "person"
	WorkerOrg    WorkerType = "org"
	WorkerAmount WorkerType = "amount"
)

type ExtractionRequest struct {
	Text        string       `json:"text"`
	WorkerTypes []WorkerType `json:"worker_types,omitempty"`
	Context     string       `json:"context,omitempty"`
}

type Entity struct {
	Type       string                 `json:"type"`
	Value      interface{}            `json:"value"`
	Confidence float64                `json:"confidence"`
	Source     WorkerType             `json:"source"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

type ExtractionResult struct {
	Entities       []Entity               `json:"entities"`
	WorkersUsed    []WorkerType           `json:"workers_used"`
	TotalTime      time.Duration          `json:"total_time"`
	Classification map[string]float64     `json:"classification"`
	GraphOps       []map[string]interface{} `json:"graph_ops,omitempty"`
}

type WorkerResult struct {
	WorkerType WorkerType
	Entities   []Entity
	Error      error
	Duration   time.Duration
}

type Orchestrator struct {
	cfg        *config.Config
	log        *logger.Logger
	llmBaseURL string
	client     *http.Client
}

func NewOrchestrator(cfg *config.Config, log *logger.Logger) *Orchestrator {
	return &Orchestrator{
		cfg:        cfg,
		log:        log.WithComponent("pipeline-orchestrator"),
		llmBaseURL: cfg.LLMServiceURL,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

func (o *Orchestrator) Extract(ctx context.Context, req ExtractionRequest) (*ExtractionResult, error) {
	start := time.Now()

	workerTypes := req.WorkerTypes
	if len(workerTypes) == 0 {
		workerTypes = o.classifyQuery(req.Text)
	}

	results := o.runWorkersConcurrently(ctx, req.Text, workerTypes)

	allEntities := make([]Entity, 0)
	for _, r := range results {
		if r.Error != nil {
			o.log.Warn().
				Str("worker", string(r.WorkerType)).
				Err(r.Error).
				Msg("Worker failed")
			continue
		}
		allEntities = append(allEntities, r.Entities...)
	}

	merged := o.mergeEntities(allEntities)

	validated, graphOps := o.validateWithHaiku(ctx, merged, req.Text)

	return &ExtractionResult{
		Entities:    validated,
		WorkersUsed: workerTypes,
		TotalTime:   time.Since(start),
		GraphOps:    graphOps,
	}, nil
}

func (o *Orchestrator) classifyQuery(text string) []WorkerType {
	scores := make(map[WorkerType]float64)

	dateKeywords := []string{"when", "date", "time", "day", "month", "year", "deadline"}
	personKeywords := []string{"who", "person", "name", "employee", "sent", "from", "to"}
	orgKeywords := []string{"company", "organization", "corp", "inc", "firm", "bank"}
	amountKeywords := []string{"amount", "price", "cost", "dollar", "million", "total"}

	textLower := bytes.ToLower([]byte(text))

	for _, kw := range dateKeywords {
		if bytes.Contains(textLower, []byte(kw)) {
			scores[WorkerDate] += 0.2
		}
	}
	for _, kw := range personKeywords {
		if bytes.Contains(textLower, []byte(kw)) {
			scores[WorkerPerson] += 0.2
		}
	}
	for _, kw := range orgKeywords {
		if bytes.Contains(textLower, []byte(kw)) {
			scores[WorkerOrg] += 0.2
		}
	}
	for _, kw := range amountKeywords {
		if bytes.Contains(textLower, []byte(kw)) {
			scores[WorkerAmount] += 0.2
		}
	}

	var types []WorkerType
	for wt, score := range scores {
		if score >= 0.2 {
			types = append(types, wt)
		}
	}

	if len(types) == 0 {
		types = []WorkerType{WorkerDate, WorkerPerson, WorkerOrg, WorkerAmount}
	}

	return types
}

func (o *Orchestrator) runWorkersConcurrently(ctx context.Context, text string, workerTypes []WorkerType) []WorkerResult {
	var wg sync.WaitGroup
	results := make(chan WorkerResult, len(workerTypes))

	for _, wt := range workerTypes {
		wg.Add(1)
		go func(workerType WorkerType) {
			defer wg.Done()
			start := time.Now()

			entities, err := o.callWorker(ctx, workerType, text)

			results <- WorkerResult{
				WorkerType: workerType,
				Entities:   entities,
				Error:      err,
				Duration:   time.Since(start),
			}
		}(wt)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var allResults []WorkerResult
	for r := range results {
		allResults = append(allResults, r)
	}

	return allResults
}

func (o *Orchestrator) callWorker(ctx context.Context, workerType WorkerType, text string) ([]Entity, error) {
	url := fmt.Sprintf("%s/extract/%s", o.llmBaseURL, workerType)

	body, err := json.Marshal(map[string]string{"text": text})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("worker returned %d", resp.StatusCode)
	}

	var result struct {
		Entities []Entity `json:"entities"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	for i := range result.Entities {
		result.Entities[i].Source = workerType
	}

	return result.Entities, nil
}

func (o *Orchestrator) mergeEntities(entities []Entity) []Entity {
	if len(entities) <= 1 {
		return entities
	}

	merged := make([]Entity, 0, len(entities))
	seen := make(map[string]int)

	for _, e := range entities {
		key := fmt.Sprintf("%s:%v", e.Type, e.Value)

		if idx, exists := seen[key]; exists {
			if e.Confidence > merged[idx].Confidence {
				merged[idx] = e
			}
		} else {
			seen[key] = len(merged)
			merged = append(merged, e)
		}
	}

	return merged
}

func (o *Orchestrator) validateWithHaiku(ctx context.Context, entities []Entity, text string) ([]Entity, []map[string]interface{}) {
	graphOps := make([]map[string]interface{}, 0)

	labelMap := map[WorkerType]string{
		WorkerDate:   "Date",
		WorkerPerson: "Person",
		WorkerOrg:    "Organization",
		WorkerAmount: "Amount",
	}

	for _, e := range entities {
		label := labelMap[e.Source]
		if label == "" {
			label = "Entity"
		}

		graphOps = append(graphOps, map[string]interface{}{
			"operation":  "CREATE_NODE",
			"label":      label,
			"properties": e.Value,
		})
	}

	return entities, graphOps
}

func (o *Orchestrator) ExtractAll(ctx context.Context, text string) (*ExtractionResult, error) {
	return o.Extract(ctx, ExtractionRequest{
		Text: text,
		WorkerTypes: []WorkerType{
			WorkerDate,
			WorkerPerson,
			WorkerOrg,
			WorkerAmount,
		},
	})
}
