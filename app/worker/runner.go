package worker

import (
	"context"
	"log"
)

func RunWorker(ctx context.Context) error {
	server := NewServer()
	mux := RegisterHandlers(server)

	shutdown := make(chan struct{})

	go func() {
		select {
		case <-ctx.Done():
			log.Println("Shutting down Asynq worker...")
			server.Shutdown()

		case <-shutdown:
			return
		}
	}()

	log.Println("Asynq worker is running...")

	err := server.Run(mux)

	close(shutdown)

	return err
}
