package api

import (
	"fafnir/shared/pkg/utils"
	"log"
	"net/http"
	"strings"

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
	r.Get("/historical/{symbol}/{period}", h.getStockHistoricalData)
	r.Get("/quote/batch", h.getStockQuoteBatch)

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

	quote, err := h.stockService.GetStockQuote(r.Context(), symbol)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, quote)
}

func (h *Handler) getStockQuoteBatch(w http.ResponseWriter, r *http.Request) {
	symbols := strings.Split(r.URL.Query().Get("symbols"), ",")

	quotes, err := h.stockService.GetStockQuoteBatch(r.Context(), symbols)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	log.Printf("quotes retrieved: %+v", quotes)

	utils.WriteJSON(w, http.StatusOK, quotes)
}

func (h *Handler) getStockHistoricalData(w http.ResponseWriter, r *http.Request) {
	symbol := chi.URLParam(r, "symbol")
	period := chi.URLParam(r, "period")

	historicalData, err := h.stockService.GetStockHistoricalData(r.Context(), symbol, period)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, historicalData)
}
