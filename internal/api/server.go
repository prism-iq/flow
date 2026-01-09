package api

import (
	"flow/internal/config"
	"flow/internal/middleware"
	"flow/internal/websocket"
	"flow/pkg/logger"

	"github.com/gin-gonic/gin"
)

type Server struct {
	cfg    *config.Config
	log    *logger.Logger
	hub    *websocket.Hub
	router *gin.Engine
}

func NewServer(cfg *config.Config, log *logger.Logger, hub *websocket.Hub) *Server {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	s := &Server{
		cfg:    cfg,
		log:    log.WithComponent("api-server"),
		hub:    hub,
		router: router,
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

func (s *Server) setupMiddleware() {
	s.router.Use(middleware.Logger(s.log))
	s.router.Use(middleware.Recovery(s.log))
	s.router.Use(middleware.CORS(s.cfg.AllowedOrigins))
}

func (s *Server) setupRoutes() {
	api := s.router.Group("/api/v1")
	{
		RegisterChatRoutes(api, s.hub, s.log, s.cfg)
		RegisterHealthRoutes(api, s.hub)
		RegisterConversationRoutes(api, s.log)
		RegisterPipelineRoutes(api, s.cfg, s.log)
		RegisterFastExtractRoutes(api, s.log)
		RegisterGraphRoutes(api, s.log)
	}

	s.router.GET("/ws", s.handleWebSocket)
}

func (s *Server) handleWebSocket(c *gin.Context) {
	HandleWebSocket(s.hub, s.log, s.cfg, c.Writer, c.Request)
}

func (s *Server) Start() error {
	addr := ":" + s.cfg.Port
	s.log.Info().Str("addr", addr).Msg("Server listening")
	return s.router.Run(addr)
}
