package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/stock-ahora/api-stock/internal/service/stock"
)

type StockHandler struct {
	Service stock.StockService
}

func (h *StockHandler) List(w http.ResponseWriter, r *http.Request) {

	page, size := parsePagination(r)

	clientAccountId, _, _ := getClientAccountIdHeader(w, r)

	list, err := h.Service.List(clientAccountId, page, size)
	if err != nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (h *StockHandler) Get(w http.ResponseWriter, r *http.Request) {

	segments := strings.Split(r.URL.Path, "/")
	idStr := segments[len(segments)-1] // última parte del path
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "UUID inválido", http.StatusBadRequest)
		return
	}
	result, err := h.Service.Get(id)
	if err != nil {
		http.Error(w, "Error al obtener el producto: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)

}
