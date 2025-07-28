package main

import (
	"github.com/andrearcaina/den/services/user-service/internal/config"
	"github.com/andrearcaina/den/shared/pkg/utils"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	router := chi.NewRouter()

	router.Use(middleware.Logger)
	cfg := config.NewConfig()

	// custom test for now
	router.Get("/user/test", func(w http.ResponseWriter, r *http.Request) {
		err := utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "Hello World"})
		if err != nil {
			return
		}
	})

	server := &http.Server{
		Addr:    cfg.PORT,
		Handler: router,
	}

	log.Printf("Starting api service on port %v\n", server.Addr)
	log.Fatal(server.ListenAndServe())
}
