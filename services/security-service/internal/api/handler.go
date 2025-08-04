package api

import (
	"fafnir/shared/pkg/utils"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type Handler struct {
	securityService *Service
}

func NewSecurityHandler(securityService *Service) *Handler {
	return &Handler{
		securityService: securityService,
	}
}

func (h *Handler) ServeSecurityRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/test", h.test)

	return r
}

func (h *Handler) test(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Security service is running",
	})
}
