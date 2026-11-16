package config

import (
	"os"

	"github.com/hibiken/asynq"
)

func NewAsynqRedisOpt() asynq.RedisClientOpt {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	return asynq.RedisClientOpt{
		Addr:     addr,
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	}
}
