package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fafnir/trade-engine/internal/api"
	"fafnir/trade-engine/internal/config"
	"fafnir/trade-engine/internal/engine"
)

func main() {
	log.Println("Starting Trade Engine Service...")

	cfg := config.New()

	eng, err := engine.NewEngine(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize engine: %v", err)
	}

	srv := api.NewServer(cfg)

	go func() {
		eng.Start()
	}()

	go func() {
		srv.Run()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Close(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	eng.Stop()

	log.Println("Trade Engine exited.")
}
