package config

import "sync"

var (
	config *AppConfig
	once   sync.Once
)

type NATSConfig struct {
	URL string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type AppConfig struct {
	NATS  NATSConfig
	Redis RedisConfig
}

func Load() *AppConfig {
	once.Do(func() {
		config = &AppConfig{
			NATS: NATSConfig{
				URL: getEnv("NATS_URL", "nats://localhost:4222"),
			},
			Redis: RedisConfig{
				Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
				Password: getEnv("REDIS_PASSWORD", ""),
				DB:       getEnvInt("REDIS_DB", 0),
			},
		}
	})
	return config
}
