package api

import (
	"net/http"
	"sync"

	"flow/internal/models"
	"flow/pkg/logger"

	"github.com/gin-gonic/gin"
)

var (
	conversations = make(map[string]*models.Conversation)
	convMu        sync.RWMutex
)

func RegisterConversationRoutes(rg *gin.RouterGroup, log *logger.Logger) {
	handler := &conversationHandler{
		log: log.WithComponent("conversation-handler"),
	}

	conv := rg.Group("/conversations")
	{
		conv.GET("", handler.list)
		conv.POST("", handler.create)
		conv.GET("/:id", handler.get)
		conv.DELETE("/:id", handler.delete)
		conv.GET("/:id/messages", handler.getMessages)
	}
}

type conversationHandler struct {
	log *logger.Logger
}

func (h *conversationHandler) list(c *gin.Context) {
	convMu.RLock()
	defer convMu.RUnlock()

	list := make([]*models.Conversation, 0, len(conversations))
	for _, conv := range conversations {
		list = append(list, conv)
	}

	c.JSON(http.StatusOK, gin.H{"conversations": list})
}

func (h *conversationHandler) create(c *gin.Context) {
	conv := models.NewConversation()

	convMu.Lock()
	conversations[conv.ID] = conv
	convMu.Unlock()

	c.JSON(http.StatusCreated, conv)
}

func (h *conversationHandler) get(c *gin.Context) {
	id := c.Param("id")

	convMu.RLock()
	conv, ok := conversations[id]
	convMu.RUnlock()

	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		return
	}

	c.JSON(http.StatusOK, conv)
}

func (h *conversationHandler) delete(c *gin.Context) {
	id := c.Param("id")

	convMu.Lock()
	delete(conversations, id)
	convMu.Unlock()

	c.Status(http.StatusNoContent)
}

func (h *conversationHandler) getMessages(c *gin.Context) {
	id := c.Param("id")

	convMu.RLock()
	conv, ok := conversations[id]
	convMu.RUnlock()

	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": conv.GetMessages()})
}
