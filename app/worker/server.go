package worker

import (
	"log"

	"flash-sale-api/config"

	"github.com/hibiken/asynq"
)

func NewServer() *asynq.Server {
	redisOpt := config.NewAsynqRedisOpt()

	return asynq.NewServer(
		redisOpt,
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
		},
	)
}

func RegisterHandlers(server *asynq.Server) *asynq.ServeMux {
	mux := asynq.NewServeMux()

	mux.HandleFunc(TypeOrderProcess, HandleOrderProcessTask)

	log.Println("Asynq handlers registered")

	return mux
}
