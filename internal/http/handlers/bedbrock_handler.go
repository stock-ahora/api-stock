package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/stock-ahora/api-stock/internal/models"
	"github.com/stock-ahora/api-stock/internal/service/bedrock"
	"gorm.io/gorm"
)

type BedbrockHandler struct {
	Db *gorm.DB
}

func (h *BedbrockHandler) ConsultaProductos(w http.ResponseWriter, r *http.Request) {

	productIdStr := r.URL.Query().Get("productId")
	if productIdStr == "" {
		http.Error(w, "productId es requerido", http.StatusBadRequest)
		return
	}

	productID, err := uuid.Parse(productIdStr)
	if err != nil {
		http.Error(w, "UUID inválido", http.StatusBadRequest)
		return
	}

	requestClient := r.URL.Query().Get("queryClient")

	svc := bedrock.NewBedrockService(context.Background(), bedrock.NOVA_PRO_AWS, "us-east-1")

	movements, product, _ := h.listLast2MonthsMovements(r.Context(), productID)

	movementsStr, _ := ToJSON(movements)
	productStr, _ := ToJSON(product)

	resultChatbot, _ := svc.ChatBot(context.Background(), bedrock.ChatBot, movementsStr+"\n"+productStr+"\nPregunta: "+requestClient)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resultChatbot)

}

func (h *BedbrockHandler) listLast2MonthsMovements(ctx context.Context, productId uuid.UUID) ([]models.Movement, models.Product, error) {
	twoMonthsAgo := time.Now().UTC().AddDate(0, -2, 0)

	var product models.Product

	err := h.Db.WithContext(ctx).First(&product, "id = ?", productId).Error
	if err != nil {
		return nil, models.Product{}, err
	}

	var ms []models.Movement
	err = h.Db.WithContext(ctx).
		// Si prefieres filtrar por date_limit, cambia "created_at" por "date_limit"
		Where("created_at >= ?", twoMonthsAgo).
		Where("product_id = ", productId).
		// Si quieres traer también los productos:
		Preload("Products").
		Order("created_at DESC").
		Find(&ms).Error

	return ms, product, err
}

func ToJSON(ms interface{}) (string, error) {
	b, err := json.Marshal(ms) // usa MarshalIndent si quieres bonito
	if err != nil {
		return "", err
	}
	return string(b), nil
}
