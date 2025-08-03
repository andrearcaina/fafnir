package api

import (
	"fafnir/shared/pkg/utils"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

type Handler struct {
	authService *Service
}

func NewAuthHandler(authService *Service) *Handler {
	return &Handler{
		authService: authService,
	}
}

func (h *Handler) ServeAuthRoutes() chi.Router {
	r := chi.NewRouter()

	r.Post("/register", h.register)
	r.Post("/login", h.login)

	return r
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var registerRequest RegisterRequest

	if err := utils.ParseJSON(r, &registerRequest); err != nil {
		response := RegisterResponse{
			Message: "Invalid request body",
		}
		utils.WriteJSON(w, http.StatusBadRequest, response)
		return
	}

	resp, code, err := h.authService.RegisterUser(r.Context(), registerRequest)
	if err != nil {
		log.Printf("Failed to register user: %v", err)
	}

	utils.WriteJSON(w, code, resp)
}

func (h *Handler) login(w http.ResponseWriter, request *http.Request) {
	var loginRequest LoginRequest

	if err := utils.ParseJSON(request, &loginRequest); err != nil {
		response := LoginResponse{
			Message: "Invalid request body",
		}
		utils.WriteJSON(w, http.StatusBadRequest, response)
		return
	}

	if loginRequest.Email == "" || loginRequest.Password == "" {
		response := LoginResponse{
			Message: "Email and password required",
		}
		utils.WriteJSON(w, http.StatusBadRequest, response)
		return
	}

	if h.authService.Login(loginRequest.Email, loginRequest.Password) {
		response := LoginResponse{
			Message: "Login successful",
		}
		utils.WriteJSON(w, http.StatusOK, response)
	} else {
		response := LoginResponse{
			Message: "Invalid credentials",
		}

		log.Println(response)

		utils.WriteJSON(w, http.StatusUnauthorized, response)
	}
}
