package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/stock-ahora/api-stock/internal/models"
	"gorm.io/gorm"
)

type DashboardHandler struct {
	Db *gorm.DB
}

func (d DashboardHandler) Get(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	typeRequest := r.URL.Query().Get("typeRequet")
	clientAccountID, _, _ := getClientAccountIdHeader(w, r)

	clienteID, _ := d.GetDimClienteID(clientAccountID)

	switch typeRequest {
	case "movementOverTime":
		result, _ := d.GetMovementOverTime(clienteID)
		json.NewEncoder(w).Encode(result)
	case "topProducts":

		result, _ := d.GetTopProducts(clienteID, 7, r)
		json.NewEncoder(w).Encode(result)

	case "stockTrend":

		result, _ := d.GetStockTrend(clienteID)
		json.NewEncoder(w).Encode(result)

	case "summaryForClient":

		result, _ := d.GetSummaryForClient(clienteID)
		json.NewEncoder(w).Encode(result)

	case "movementsByType":

		result, _ := d.GetMovementsByTypeForClient(clienteID)
		json.NewEncoder(w).Encode(result)

	case "movementsByUser":

		result, _ := d.GetMovementsByUserForClient(clienteID)
		json.NewEncoder(w).Encode(result)

	default:
		http.Error(w, "Tipo de solicitud no válido", http.StatusBadRequest)
		return
	}

}

type MovementOverTime struct {
	Fecha    time.Time `json:"fecha"`
	Ingresos int64     `json:"ingresos"`
	Egresos  int64     `json:"egresos"`
}

func (d DashboardHandler) GetMovementOverTime(clientID int) ([]MovementOverTime, error) {
	var results []MovementOverTime

	err := d.Db.Table("fact_product_movement f").
		Select(`"df".fecha as fecha, `+
			`SUM(CASE WHEN f.signo = 1 THEN f.cantidad ELSE 0 END) AS ingresos, `+
			`SUM(CASE WHEN f.signo = -1 THEN f.cantidad ELSE 0 END) AS egresos`).
		Joins("JOIN dim_fecha df ON f.fecha_key = df.fecha_key").
		Where("f.cliente_id = ?", clientID).
		Group("df.fecha").
		Order("df.fecha").
		Scan(&results).Error

	return results, err
}

type TopProduct struct {
	NombreProducto string `json:"nombre_producto"`
	Egresos        int64  `json:"egresos"`
	Ingresos       int64  `json:"ingresos"`
	Total          int64  `json:"total"`
}

func (d DashboardHandler) GetTopProducts(clientID int, limit int, r *http.Request) ([]TopProduct, error) {
	var results []TopProduct

	startDate, endDate, _ := parseDateParams(r)

	query := d.Db.Table("fact_product_movement f").
		Select(`dp.id, dp.nombre AS nombre_producto,
		SUM(CASE WHEN f.tipo_movimiento_id = 8 THEN f.cantidad ELSE 0 END) AS egresos,
		SUM(CASE WHEN f.tipo_movimiento_id = 7 THEN f.cantidad ELSE 0 END) AS ingresos,
		SUM(f.cantidad) AS total`).
		Joins("JOIN dim_producto dp ON f.producto_id = dp.id").
		Joins("JOIN dim_fecha df ON df.fecha_key = f.fecha_key").
		Where("f.cliente_id = ?", clientID)

	if !startDate.IsZero() && !endDate.IsZero() {
		query = query.Where("df.fecha BETWEEN ? AND ?", startDate, endDate)
	}

	err := query.
		Group("dp.nombre, dp.id").
		Order("total DESC").
		Limit(limit).
		Scan(&results).Error

	return results, err
}

func parseDateParams(r *http.Request) (time.Time, time.Time, error) {
	layout := "2006-01-02"
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	var startDate, endDate time.Time
	var err error

	if startStr != "" {
		startDate, err = time.Parse(layout, startStr)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}
	if endStr != "" {
		endDate, err = time.Parse(layout, endStr)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}
	return startDate, endDate, nil
}

type StockTrend struct {
	Fecha          time.Time `json:"fecha"`
	StockAcumulado int64     `json:"stock_acumulado"`
}

func (d DashboardHandler) GetStockTrend(clientID int) ([]StockTrend, error) {
	var results []StockTrend
	query := `
        SELECT df.fecha AS fecha,
               SUM(f.cantidad * f.signo) OVER (ORDER BY df.fecha) AS stock_acumulado
        FROM fact_product_movement f
        JOIN dim_fecha df ON f.fecha_key = df.fecha_key
        WHERE f.cliente_id = ?
        ORDER BY df.fecha
    `

	err := d.Db.Raw(query, clientID).Scan(&results).Error
	return results, err
}

type SummaryForClient struct {
	Ingresos int64 `json:"ingresos"`
	Egresos  int64 `json:"egresos"`
}

func (d DashboardHandler) GetSummaryForClient(clientID int) (SummaryForClient, error) {
	var result SummaryForClient
	err := d.Db.Table("fact_product_movement f").
		Select(`
            SUM(CASE WHEN f.signo = 1 THEN f.cantidad ELSE 0 END) AS ingresos,
            SUM(CASE WHEN f.signo = -1 THEN f.cantidad ELSE 0 END) AS egresos
        `).
		Where("f.cliente_id = ?", clientID).
		Scan(&result).Error
	return result, err
}

type MovementsByType struct {
	Tipo  string `json:"tipo"`
	Total int64  `json:"total"`
}

func (d DashboardHandler) GetMovementsByTypeForClient(clientID int) ([]MovementsByType, error) {
	var results []MovementsByType

	err := d.Db.Table("fact_product_movement f").
		Select("dtm.nombre AS tipo, SUM(f.cantidad) AS total").
		Joins("JOIN dim_tipo_movimiento dtm ON f.tipo_movimiento_id = dtm.id").
		Where("f.cliente_id = ?", clientID).
		Group("dtm.nombre").
		Scan(&results).Error

	return results, err
}

type MovementsByUser struct {
	Usuario     string `json:"usuario"`
	Movimientos int64  `json:"movimientos"`
}

func (d DashboardHandler) GetMovementsByUserForClient(clientID int) ([]MovementsByUser, error) {
	var results []MovementsByUser

	err := d.Db.Table("fact_product_movement f").
		Select("du.nombre AS usuario, COUNT(f.id) AS movimientos").
		Joins("JOIN dim_usuario du ON f.usuario_id = du.id").
		Where("f.cliente_id = ?", clientID).
		Group("du.nombre").
		Order("movimientos DESC").
		Scan(&results).Error

	return results, err
}

func (d DashboardHandler) GetDimClienteID(uuid uuid.UUID) (int, error) {
	var cliente models.DimCliente
	// Busca el registro cuyo cliente_uuid coincida y trae el primero
	err := d.Db.Where("cliente_uuid = ?", uuid).First(&cliente).Error
	if err != nil {
		return 0, err
	}
	return cliente.ID, nil
}
