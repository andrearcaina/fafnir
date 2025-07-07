package main

import (
	"github.com/andrearcaina/den/pkg/utils"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	r := chi.NewRouter()

	// custom test for now
	r.Get("/portfolio/test", func(w http.ResponseWriter, r *http.Request) {
		utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "Hello World"})
	})

	if err := http.ListenAndServe(":8081", r); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
