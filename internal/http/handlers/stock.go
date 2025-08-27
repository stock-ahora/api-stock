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

func NewStockHandler(svc service.StockService) StockHandler {
    return &StockHandler{svc: svc}
}

func Health(w http.ResponseWriter, rhttp.Request) {
    w.Header().Set("Content-Type", "application/json")
     = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h StockHandler) List(w http.ResponseWriter, rhttp.Request) {
    w.Header().Set("Content-Type", "application/json")
    list, err := h.svc.List()
    if err != nil {
        http.Error(w, "error listing", http.StatusInternalServerError)
        return
    }
     = json.NewEncoder(w).Encode(list)
}

func (h StockHandler) Create(w http.ResponseWriter, rhttp.Request) {
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
    _ = json.NewEncoder(w).Encode(created)
}