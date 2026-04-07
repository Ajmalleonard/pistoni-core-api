package main

import (
	"log"

	"pistoni/core/internal/config"
	"pistoni/core/internal/server"
)

func main() {
	cfg := config.Load()
	srv, err := server.New(cfg)
	if err != nil {
		log.Fatalf("failed to init server: %v", err)
	}

	log.Printf("pistoni core starting http_port=%s socket_path=%s socket_port=%s", cfg.Port, cfg.SocketPath, cfg.SocketPort)
	if err := srv.Run(); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
