package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/stock-ahora/api-stock/internal/domain"
	"github.com/stock-ahora/api-stock/internal/service"
)

type StockHandler struct {
	svc service.StockService
}

// Return a *StockHandler if you construct with &StockHandler{...}
func NewStockHandler(svc service.StockService) *StockHandler {
	return &StockHandler{svc: svc}
}

// Handlers must use (w http.ResponseWriter, r *http.Request)
func (h *StockHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		http.Error(w, "failed to write response", http.StatusInternalServerError)
	}
}

func (h *StockHandler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	list, err := h.svc.List()
	if err != nil {
		http.Error(w, "error listing", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(list); err != nil {
		http.Error(w, "failed to write response", http.StatusInternalServerError)
	}
}

func (h *StockHandler) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req domain.Stock
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	created, err := h.svc.Create(req)
	if err != nil {
		http.Error(w, "error creating", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(created); err != nil {
		// At this point headers are sent, so best we can do is log in real app
	}
}
