package handlers

import (
	"github.com/andrearcaina/den/services/auth-service/internal/service"
	"github.com/andrearcaina/den/shared/pkg/utils"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type Handler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *Handler {
	return &Handler{
		authService: authService,
	}
}

func (h *Handler) ServeHTTP() chi.Router {
	r := chi.NewRouter()

	r.Post("/login", func(w http.ResponseWriter, request *http.Request) {
		user := request.FormValue("username")
		pass := request.FormValue("password")

		// just to test (this should be replaced with proper validation and proper error handling)
		// if the user does not send a password, it should return a status code of 400 Bad Request, but this doesn't handle that (yet)
		if h.authService.Login(user, pass) {
			response := map[string]string{"message": "Login successful"}
			utils.WriteJSON(w, http.StatusOK, response)
		} else {
			response := map[string]string{"error": "Invalid credentials"}
			utils.WriteJSON(w, http.StatusUnauthorized, response)
		}
	})

	return r
}
