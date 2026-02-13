package main

import (
	"context"
	"fafnir/auth-service/internal/api"
	"fafnir/shared/pkg/logger"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// instantiate custom logger (that wraps slog/logrus) for structured logging
	l := logger.New(nil)

	server := api.NewServer(l)

	// this starts the server in a goroutine so it can run concurrently (so that we can listen for OS signals)
	// if we didn't do this, the server would block the main thread, and we wouldn't be able to listen for OS signals
	go func() {
		if err := server.Run(); err != nil {
			l.Error(context.Background(), "Server run failed", "error", err)
			os.Exit(1)
		}
	}()

	// this sets up a channel to listen for OS signals when a user wants to stop the service (like Ctrl+C)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Close(ctx); err != nil {
		l.Error(ctx, "Server close failed", "error", err)
		os.Exit(1)
	}
}
