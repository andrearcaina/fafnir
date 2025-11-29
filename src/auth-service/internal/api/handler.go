package api

import (
	"fafnir/shared/pkg/utils"
	"fafnir/shared/pkg/validator"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	authService *Service
	validator   *validator.Validator
}

func NewAuthHandler(authService *Service, validator *validator.Validator) *Handler {
	return &Handler{
		authService: authService,
		validator:   validator,
	}
}

func (h *Handler) ServeAuthRoutes() chi.Router {
	r := chi.NewRouter()

	r.Post("/register", h.register)
	r.Post("/login", h.login)

	// create a middleware for certain endpoints
	authMiddleware := CheckAuth(h.authService)

	r.With(authMiddleware).Delete("/logout", h.logout)
	r.With(authMiddleware).Delete("/delete", h.deleteAccount)
	r.With(authMiddleware).Get("/me", h.getUserInfo)
	return r
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var registerRequest RegisterRequest

	if err := utils.DecodeJSON(r, &registerRequest); err != nil {
		utils.HandleError(w, err)
		return
	}

	log.Printf("Received Registration: %+v\n", registerRequest)

	if err := h.validator.ValidateRequest(registerRequest); err != nil {
		utils.HandleError(w, err)
		return
	}

	log.Printf("Sup")

	resp, err := h.authService.RegisterUser(r.Context(), registerRequest)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, resp)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var loginRequest LoginRequest

	if err := utils.DecodeJSON(r, &loginRequest); err != nil {
		utils.HandleError(w, err)
		return
	}

	if err := h.validator.ValidateRequest(loginRequest); err != nil {
		utils.HandleError(w, err)
		return
	}

	resp, err := h.authService.Login(r.Context(), loginRequest)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	utils.SetCookie(w, "auth_token", resp.JwtToken, 24*3600, true, false, http.SameSiteLaxMode)
	utils.SetCookie(w, "csrf_token", resp.CsrfToken, 24*3600, false, false, http.SameSiteLaxMode)

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"message": resp.Message, // don't send the JWT token in the response body because it's already set in the cookie
	})
}

func (h *Handler) logout(w http.ResponseWriter, _ *http.Request) {
	utils.SetCookie(w, "auth_token", "", -1, true, false, http.SameSiteLaxMode)
	utils.SetCookie(w, "csrf_token", "", -1, false, false, http.SameSiteLaxMode)

	utils.WriteJSON(w, http.StatusNoContent, nil)
}

func (h *Handler) deleteAccount(w http.ResponseWriter, r *http.Request) {
	userId, err := GetUserIdFromContext(r.Context())
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	// delete the account first
	if err := h.authService.DeleteAccount(r.Context(), userId); err != nil {
		utils.HandleError(w, err)
		return
	}

	// then logout the user and send 204 response
	h.logout(w, r)
}

func (h *Handler) getUserInfo(w http.ResponseWriter, r *http.Request) {
	userId, err := GetUserIdFromContext(r.Context())
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	resp, err := h.authService.GetUserInfo(r.Context(), userId)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}
