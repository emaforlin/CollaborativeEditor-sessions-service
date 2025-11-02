package main

import (
	"log"
	"time"

	"github.com/emaforlin/ce-sessions-service/api"
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

	server := api.NewAPIServer(conf)

	sessionsHandler := api.NewSessionsHandler(sessionRepository)

	server.RegisterHandler("/session/documents/{document_id}", sessionsHandler.ServeHTTP)

	log.Fatal(server.Start())
}

func preStartupSetup() {
	config.Load()
}
