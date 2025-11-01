package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/emaforlin/ce-sessions-service/config"
	"github.com/emaforlin/ce-sessions-service/events"
	"github.com/emaforlin/ce-sessions-service/redis"
	"github.com/emaforlin/ce-sessions-service/repository"
	"github.com/nats-io/nats.go"
)

func main() {
	preStartupSetup()
	conf := config.GetConfig()

	redisClient, err := redis.NewRedisClient(conf.Redis.Addr, conf.Redis.Password, conf.Redis.DB)
	if err != nil {
		log.Fatalf("failed to establish connection with redis: %v", err)
	}

	natsClient, err := nats.Connect(conf.NATS.URL,
		nats.RetryOnFailedConnect(true),
		nats.ReconnectWait(2*time.Second),
	)
	if err != nil {
		log.Fatalf("failed to connect to NATS: %v", err)
	}
	defer natsClient.Close()

	sessionRepository, err := repository.NewRedisSessionsRepo(redisClient)
	if err != nil {
		log.Fatalf("failed to initialize the sessions repository: %v", err)
	}

	eventConsumer := events.NewEventConsumer(natsClient, sessionRepository)

	if err := eventConsumer.Start(); err != nil {
		log.Fatalf("failed to start event consumer: %v", err)
	}
	defer func() {
		if err := eventConsumer.Stop(); err != nil {
			log.Printf("error stopping event consumer: %v", err)
		}
	}()

	log.Println("Sessions service started. Waiting for events...")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Shutting down sessions service...")
}

func preStartupSetup() {
	config.Load()
}
