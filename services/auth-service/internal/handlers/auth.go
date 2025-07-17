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
		var req struct {
			User string `json:"user"`
			Pass string `json:"pass"`
		}

		if err := utils.ParseJSON(request, &req); err != nil {
			response := map[string]string{"error": "Invalid request body"}
			if err := utils.WriteJSON(w, http.StatusBadRequest, response); err != nil {
				http.Error(w, "Failed to write response", http.StatusInternalServerError)
				return
			}
			return
		}

		if req.User == "" || req.Pass == "" {
			response := map[string]string{"error": "Username and password are required"}
			if err := utils.WriteJSON(w, http.StatusBadRequest, response); err != nil {
				http.Error(w, "Failed to write response", http.StatusInternalServerError)
				return
			}
			return
		}

		// just to test (this should be replaced with proper validation and proper error handling)
		// if the user does not send a password, it should return a status code of 400 Bad Request, but this doesn't handle that (yet)
		if h.authService.Login(req.User, req.Pass) {
			response := map[string]string{"message": "Login successful"}
			if err := utils.WriteJSON(w, http.StatusOK, response); err != nil {
				http.Error(w, "Failed to write response", http.StatusInternalServerError)
				return
			}
		} else {
			response := map[string]string{"error": "Invalid credentials"}
			if err := utils.WriteJSON(w, http.StatusUnauthorized, response); err != nil {
				http.Error(w, "Failed to write response", http.StatusInternalServerError)
				return
			}
		}
	})

	return r
}
