package main

import (
	"github.com/andrearcaina/den/shared/pkg/utils"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)

	// custom test for now
	r.Get("/user/test", func(w http.ResponseWriter, r *http.Request) {
		utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "Hello World"})
	})

	if err := http.ListenAndServe(":8081", r); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
