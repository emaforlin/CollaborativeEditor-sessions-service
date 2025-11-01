package main

import (
	"log"

	"github.com/emaforlin/ce-sessions-service/config"
	"github.com/emaforlin/ce-sessions-service/redis"
	"github.com/emaforlin/ce-sessions-service/repository"
)

func main() {
	preStartupSetup()
	conf := config.GetConfig()

	redisClient, err := redis.NewRedisClient(conf.Redis.Addr, conf.Redis.Password, conf.Redis.DB)
	if err != nil {
		log.Fatalf("failed to establish connection with redis: %v", err)
	}

	_, err = repository.NewRedisSessionsRepo(redisClient)
	if err != nil {
		log.Fatalf("failed to create sessions repository: %v", err)
	}

}

func preStartupSetup() {
	config.Load()
}
