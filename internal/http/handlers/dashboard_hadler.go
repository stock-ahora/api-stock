package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
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
		result, _ := d.GetMovementOverTime(clienteID, r)
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
	Periodo        time.Time `json:"periodo"`
	Mes            string    `json:"mes"`
	Ingresos       int64     `json:"ingresos"`
	Egresos        int64     `json:"egresos"`
	StockAcumulado int64     `json:"stock_acumulado"`
}

func (d DashboardHandler) GetMovementOverTime(clientID int, r *http.Request) ([]MovementOverTime, error) {
	var results []MovementOverTime

	startDate, endDate, _ := parseDateParams(r)
	period := r.URL.Query().Get("period") // "week" o "month"
	if period != "week" {
		period = "month"
	}
	productoId := r.URL.Query().Get("productoId")

	// Construir WHERE dinámico según parámetros opcionales
	whereClause := " WHERE f.cliente_id = ?"
	args := []interface{}{period, clientID}

	if !startDate.IsZero() && !endDate.IsZero() {
		whereClause += " AND df.fecha BETWEEN ? AND ?"
		args = append(args, startDate, endDate)
	}
	if productoId != "" {
		if pid, err := strconv.Atoi(productoId); err == nil {
			whereClause += " AND f.producto_id = ?"
			args = append(args, pid)
		}
	}

	query := `
WITH base AS (
    SELECT
        CASE WHEN ? = 'week' THEN date_trunc('week', df.fecha) ELSE date_trunc('month', df.fecha) END AS periodo,
        f.tipo_movimiento_id,
        f.cantidad,
        (f.cantidad * f.signo) AS movimiento
    FROM fact_product_movement f
    JOIN dim_fecha df ON f.fecha_key = df.fecha_key` + whereClause + `
),
resumen AS (
    SELECT
        periodo,
        SUM(CASE WHEN tipo_movimiento_id = 7 THEN cantidad ELSE 0 END) AS ingresos,
        SUM(CASE WHEN tipo_movimiento_id = 8 THEN cantidad ELSE 0 END) AS egresos,
        SUM(movimiento) AS movimiento_del_periodo
    FROM base
    GROUP BY periodo
)
SELECT
    periodo,
    (ARRAY['enero','febrero','marzo','abril','mayo','junio','julio','agosto','septiembre','octubre','noviembre','diciembre'])[date_part('month', periodo)::int] AS mes,
    ingresos,
    egresos,
    SUM(movimiento_del_periodo) OVER (ORDER BY periodo) AS stock_acumulado
FROM resumen
ORDER BY periodo;
`

	err := d.Db.Raw(query, args...).Scan(&results).Error

	if period == "week" && err == nil && !startDate.IsZero() && !endDate.IsZero() {
		filtered := make([]MovementOverTime, 0, len(results))
		for _, it := range results {
			if !it.Periodo.Before(startDate) && !it.Periodo.After(endDate) {
				filtered = append(filtered, it)
			}
		}
		results = filtered
	}

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

func (d DashboardHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	var results []products

	_ = d.Db.Table("dim_producto dp").
		Select("dp.nombre as name, dp.id as id, dp.producto_uuid as uuid").
		Scan(&results).Error

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

type products struct {
	Name string    `json:"name"`
	Id   int       `json:"id"`
	Uuid uuid.UUID `json:"uuid"`
}
