package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	r := chi.NewRouter()

	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello world"))
	})

	logger.Info(fmt.Sprintf("Starting auth on port 8080"))
	if err := http.ListenAndServe(":8080", r); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
