package api

import (
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

	metadata, err := h.stockService.GetStockMetadata(r.Context(), symbol)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, metadata)
}

func (h *Handler) getStockQuote(w http.ResponseWriter, r *http.Request) {
	symbol := chi.URLParam(r, "symbol")

	quote, err := h.stockService.SearchStockQuote(r.Context(), symbol)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, quote)
}
