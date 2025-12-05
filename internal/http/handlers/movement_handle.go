package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/stock-ahora/api-stock/internal/models"
	"github.com/stock-ahora/api-stock/internal/service/movement"
	"gorm.io/gorm"
)

type MovementHandler struct {
	Service movement.MovementService
	Db      *gorm.DB
}

func (h *MovementHandler) Notification(w http.ResponseWriter, r *http.Request) {

	var notification []models.Notification
	h.Db.Find(&notification)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notification)
	h.Db.Exec("update public.notification set is_read = true")
}

func (h *MovementHandler) List(w http.ResponseWriter, r *http.Request) {

	segments := strings.Split(r.URL.Path, "/")
	idStr := segments[len(segments)-1] // última parte del path
	id, err := uuid.Parse(idStr)
	page, size := parsePagination(r)

	movements, err := h.Service.List(id, page, size)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movements)

}
