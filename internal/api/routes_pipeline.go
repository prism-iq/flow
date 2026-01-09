package api

import (
	"net/http"

	"flow/internal/config"
	"flow/internal/pipeline"
	"flow/pkg/logger"

	"github.com/gin-gonic/gin"
)

type PipelineRequest struct {
	Text        string   `json:"text" binding:"required"`
	WorkerTypes []string `json:"worker_types,omitempty"`
	ExtractAll  bool     `json:"extract_all,omitempty"`
}

func RegisterPipelineRoutes(rg *gin.RouterGroup, cfg *config.Config, log *logger.Logger) {
	orchestrator := pipeline.NewOrchestrator(cfg, log)

	handler := &pipelineHandler{
		orchestrator: orchestrator,
		log:          log.WithComponent("pipeline-handler"),
	}

	p := rg.Group("/pipeline")
	{
		p.POST("/extract", handler.extract)
		p.POST("/extract/all", handler.extractAll)
		p.POST("/classify", handler.classify)
	}
}

type pipelineHandler struct {
	orchestrator *pipeline.Orchestrator
	log          *logger.Logger
}

func (h *pipelineHandler) extract(c *gin.Context) {
	var req PipelineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	workerTypes := make([]pipeline.WorkerType, 0)
	for _, wt := range req.WorkerTypes {
		workerTypes = append(workerTypes, pipeline.WorkerType(wt))
	}

	result, err := h.orchestrator.Extract(c.Request.Context(), pipeline.ExtractionRequest{
		Text:        req.Text,
		WorkerTypes: workerTypes,
	})

	if err != nil {
		h.log.Error().Err(err).Msg("Extraction failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *pipelineHandler) extractAll(c *gin.Context) {
	var req PipelineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.orchestrator.ExtractAll(c.Request.Context(), req.Text)
	if err != nil {
		h.log.Error().Err(err).Msg("Full extraction failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *pipelineHandler) classify(c *gin.Context) {
	var req struct {
		Text string `json:"text" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"text":           req.Text,
		"classification": "multi",
		"scores": gin.H{
			"date":   0.3,
			"person": 0.4,
			"org":    0.2,
			"amount": 0.1,
		},
	})
}
