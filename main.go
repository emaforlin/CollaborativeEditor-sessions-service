package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/emaforlin/ce-sessions-service/config"
	natsclient "github.com/emaforlin/ce-sessions-service/nats"
	redisclient "github.com/emaforlin/ce-sessions-service/redis"
	sessions "github.com/emaforlin/ce-sessions-service/session"
)

func main() {
	cfg := config.Load()

	natsConn, err := natsclient.NewNATSClient(cfg.NATS.URL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer natsConn.Close()
	log.Printf("Connected to NATS at %s", cfg.NATS.URL)

	redisConn, err := redisclient.NewRedisClient(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisConn.Close()
	log.Printf("Connected to Redis at %s", cfg.Redis.Addr)

	sessionManager := sessions.NewSessionManager(natsConn, redisConn)
	if err := sessionManager.Start(); err != nil {
		log.Fatalf("Failed to start session manager: %v", err)
	}
	defer sessionManager.Stop()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Sessions service is running...")
	log.Println("Press Ctrl+C to stop")

	<-sigChan
	log.Println("Shutting down sessions service...")
}
