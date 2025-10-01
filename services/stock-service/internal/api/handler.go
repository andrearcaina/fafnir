package api

import (
	"fafnir/shared/pkg/errors"
	"fafnir/shared/pkg/utils"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	stockService *Service
}

func NewStockHandler(stockService *Service) *Handler {
	return &Handler{
		stockService: stockService,
	}
}

func (h *Handler) ServeStockRoutes() http.Handler {
	r := chi.NewRouter()

	r.Get("/metadata/{symbol}", h.getStockMetadata)
	r.Get("/quote/{symbol}", h.getStockQuote)

	return r
}

func (h *Handler) getStockMetadata(w http.ResponseWriter, r *http.Request) {
	symbol := chi.URLParam(r, "symbol")

	if symbol == "" {
		utils.HandleError(w, errors.BadRequestError("Invalid symbol").WithDetails("The provided symbol is empty"))
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"symbol": symbol,
	})
}

func (h *Handler) getStockQuote(w http.ResponseWriter, r *http.Request) {
	symbol := chi.URLParam(r, "symbol")

	if symbol == "" {
		utils.HandleError(w, errors.BadRequestError("Invalid symbol").WithDetails("The provided symbol is empty"))
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"symbol": symbol,
	})
}
