package main

import (
	"context"
	"fafnir/api-gateway/internal/api"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	server := api.NewServer()

	// this starts the server in a goroutine so it can run concurrently (so that we can listen for OS signals)
	// if we didn't do this, the server would block the main thread, and we wouldn't be able to listen for OS signals
	go func() {
		log.Fatal(server.Run())
	}()

	// this sets up a channel to listen for OS signals when a user wants to stop the service (like Ctrl+C)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Fatal(server.Close(ctx))
}
