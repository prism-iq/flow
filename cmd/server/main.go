package main

import (
	"flow/internal/api"
	"flow/internal/config"
	"flow/internal/websocket"
	"flow/pkg/logger"
)

func main() {
	cfg := config.Load()
	log := logger.New(cfg.LogLevel)

	log.Info().Str("port", cfg.Port).Msg("Starting Flow server")

	hub := websocket.NewHub(log)
	go hub.Run()

	server := api.NewServer(cfg, log, hub)
	if err := server.Start(); err != nil {
		log.Fatal().Err(err).Msg("Server failed")
	}
}
