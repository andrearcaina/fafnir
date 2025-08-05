package api

import (
	"fafnir/shared/pkg/utils"
	"github.com/go-chi/chi/v5"
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

	// create a middleware for certain endpoints
	authMiddleware := CheckAuth(h.authService)

	r.With(authMiddleware).Delete("/logout", h.logout)
	r.With(authMiddleware).Get("/me", h.getUserInfo)
	return r
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var registerRequest RegisterRequest

	if err := utils.DecodeJSON(r, &registerRequest); err != nil {
		utils.WriteError(w, http.StatusBadRequest, RegisterResponse{
			Message: "Invalid request format",
		}, err)
		return
	}

	resp, code, err := h.authService.RegisterUser(r.Context(), registerRequest)
	if err != nil {
		utils.WriteError(w, int(code), resp, err)
		return
	}

	utils.WriteJSON(w, int(code), resp)
}

func (h *Handler) login(w http.ResponseWriter, request *http.Request) {
	var loginRequest LoginRequest

	err := utils.DecodeJSON(request, &loginRequest)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, RegisterResponse{
			Message: "Invalid request format",
		}, err)
		return
	}

	resp, code, err := h.authService.Login(request.Context(), loginRequest)
	if err != nil {
		utils.WriteError(w, int(code), resp, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    resp.JwtToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   24 * 3600,
	})

	utils.WriteJSON(w, int(code), resp)
}

func (h *Handler) logout(w http.ResponseWriter, _ *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	utils.WriteJSON(w, http.StatusNoContent, nil)
}

func (h *Handler) getUserInfo(w http.ResponseWriter, r *http.Request) {
	userId, err := GetUserIdFromContext(r.Context())
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, UserInfoResponse{}, err)
		return
	}

	resp, code, err := h.authService.GetUserInfo(r.Context(), userId)
	if err != nil {
		utils.WriteError(w, int(code), resp, err)
		return
	}

	utils.WriteJSON(w, int(code), resp)
}
