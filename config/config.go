package config

import (
	"sync"
	"time"
)

var (
	config *Config
	once   sync.Once
)

type ServerConfig struct {
	Port         string
	Host         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type NATSConfig struct {
	URL string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type Config struct {
	Server ServerConfig
	NATS   NATSConfig
	Redis  RedisConfig
}

func Load() {
	once.Do(func() {
		config = &Config{
			Server: ServerConfig{
				Port:         getEnv("SERVER_PORT", "9002"),
				Host:         getEnv("SERVER_HOST", "localhost"),
				ReadTimeout:  getEnvDuration("SERVER_READ_TIMEOUT", 5*time.Second),
				WriteTimeout: getEnvDuration("SERVER_READ_TIMEOUT", 2*time.Second),
			},
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
}

func GetConfig() Config {
	Load()
	return *config
}

func (c *Config) GetServerAddress() string {
	return c.Server.Host + ":" + c.Server.Port
}
