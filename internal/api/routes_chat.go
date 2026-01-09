package api

import (
	"net/http"

	"flow/internal/config"
	"flow/internal/models"
	"flow/internal/services"
	"flow/internal/websocket"
	"flow/pkg/logger"

	"github.com/gin-gonic/gin"
)

type ChatRequest struct {
	Message        string `json:"message" binding:"required"`
	ConversationID string `json:"conversation_id,omitempty"`
	UseRAG         bool   `json:"use_rag,omitempty"`
}

type ChatResponse struct {
	ConversationID string          `json:"conversation_id"`
	Message        *models.Message `json:"message"`
}

func RegisterChatRoutes(rg *gin.RouterGroup, hub *websocket.Hub, log *logger.Logger, cfg *config.Config) {
	handler := &chatHandler{
		hub:        hub,
		log:        log.WithComponent("chat-handler"),
		llmService: services.NewLLMService(cfg.LLMServiceURL, log),
		ragService: services.NewRAGService(cfg.RAGServiceURL, log),
	}

	chat := rg.Group("/chat")
	{
		chat.POST("/send", handler.sendMessage)
		chat.POST("/stream", handler.streamMessage)
	}
}

type chatHandler struct {
	hub        *websocket.Hub
	log        *logger.Logger
	llmService *services.LLMService
	ragService *services.RAGService
}

func (h *chatHandler) sendMessage(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userMsg := models.NewMessage(models.RoleUser, req.Message)

	var context string
	if req.UseRAG {
		docs, err := h.ragService.Query(c.Request.Context(), req.Message, 5)
		if err != nil {
			h.log.Warn().Err(err).Msg("RAG query failed")
		} else {
			context = docs
		}
	}

	response, err := h.llmService.Generate(c.Request.Context(), req.Message, context)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	assistantMsg := models.NewMessage(models.RoleAssistant, response)

	c.JSON(http.StatusOK, ChatResponse{
		ConversationID: req.ConversationID,
		Message:        assistantMsg,
	})
}

func (h *chatHandler) streamMessage(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	var context string
	if req.UseRAG {
		docs, err := h.ragService.Query(c.Request.Context(), req.Message, 5)
		if err == nil {
			context = docs
		}
	}

	tokenChan := make(chan string, 100)
	errChan := make(chan error, 1)

	go func() {
		err := h.llmService.GenerateStream(c.Request.Context(), req.Message, context, tokenChan)
		if err != nil {
			errChan <- err
		}
		close(tokenChan)
	}()

	c.Stream(func(w gin.ResponseWriter) bool {
		select {
		case token, ok := <-tokenChan:
			if !ok {
				c.SSEvent("done", gin.H{"done": true})
				return false
			}
			c.SSEvent("token", gin.H{"content": token})
			return true
		case err := <-errChan:
			c.SSEvent("error", gin.H{"error": err.Error()})
			return false
		}
	})
}
