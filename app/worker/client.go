package worker

import (
	"flash-sale-api/config"

	"github.com/hibiken/asynq"
)

func NewClient() *asynq.Client {
	redisOpt := config.NewAsynqRedisOpt()

	return asynq.NewClient(redisOpt)
}
