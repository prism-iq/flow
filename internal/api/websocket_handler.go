package api

import (
	"encoding/json"
	"net/http"

	"flow/internal/config"
	"flow/internal/models"
	"flow/internal/services"
	"flow/internal/websocket"
	"flow/pkg/logger"
)

type WSMessage struct {
	Type           string `json:"type"`
	Content        string `json:"content,omitempty"`
	ConversationID string `json:"conversation_id,omitempty"`
	UseRAG         bool   `json:"use_rag,omitempty"`
}

func HandleWebSocket(hub *websocket.Hub, log *logger.Logger, cfg *config.Config, w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("WebSocket upgrade failed")
		return
	}

	llmService := services.NewLLMService(cfg.LLMServiceURL, log)
	ragService := services.NewRAGService(cfg.RAGServiceURL, log)

	handler := func(client *websocket.Client, message []byte) {
		handleWSMessage(client, message, llmService, ragService, log)
	}

	client := websocket.NewClient(hub, conn, log, handler)
	hub.Register(client)

	go client.WritePump()
	go client.ReadPump()
}

func handleWSMessage(client *websocket.Client, message []byte, llm *services.LLMService, rag *services.RAGService, log *logger.Logger) {
	var msg WSMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		client.SendEvent(models.NewErrorEvent("invalid message format"))
		return
	}

	switch msg.Type {
	case "chat":
		handleChatMessage(client, msg, llm, rag, log)
	case "ping":
		client.SendJSON(map[string]string{"type": "pong"})
	default:
		client.SendEvent(models.NewErrorEvent("unknown message type"))
	}
}

func handleChatMessage(client *websocket.Client, msg WSMessage, llm *services.LLMService, rag *services.RAGService, log *logger.Logger) {
	client.SendEvent(models.NewStartEvent())

	var context string
	if msg.UseRAG {
		docs, err := rag.Query(nil, msg.Content, 5)
		if err == nil {
			context = docs
		}
	}

	tokenChan := make(chan string, 100)
	go func() {
		llm.GenerateStream(nil, msg.Content, context, tokenChan)
		close(tokenChan)
	}()

	for token := range tokenChan {
		client.SendEvent(models.NewTokenEvent(token))
	}

	client.SendEvent(models.NewDoneEvent())
}
