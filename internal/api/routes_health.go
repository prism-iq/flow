package api

import (
	"net/http"
	"runtime"

	"flow/internal/websocket"

	"github.com/gin-gonic/gin"
)

func RegisterHealthRoutes(rg *gin.RouterGroup, hub *websocket.Hub) {
	handler := &healthHandler{hub: hub}

	rg.GET("/health", handler.health)
	rg.GET("/status", handler.status)
}

type healthHandler struct {
	hub *websocket.Hub
}

func (h *healthHandler) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

func (h *healthHandler) status(c *gin.Context) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	c.JSON(http.StatusOK, gin.H{
		"status":            "ok",
		"connected_clients": h.hub.ClientCount(),
		"goroutines":        runtime.NumGoroutine(),
		"memory_alloc_mb":   mem.Alloc / 1024 / 1024,
		"memory_sys_mb":     mem.Sys / 1024 / 1024,
	})
}
