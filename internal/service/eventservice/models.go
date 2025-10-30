package eventservice

import (
	"time"

	"github.com/google/uuid"
)

type BaseEvent struct {
	EventID       string            `json:"event_id"`
	EventType     string            `json:"event_type"`
	OccurredAt    time.Time         `json:"occurred_at"`
	Version       string            `json:"version"`
	CorrelationID string            `json:"correlation_id,omitempty"`
	Source        string            `json:"source,omitempty"`
	Headers       map[string]string `json:"headers,omitempty"`
}

type MovementsEvent struct {
	Id uuid.UUID `json:"id"`
	BaseEvent
	ProductPerMovement []ProductPerMovement `json:"products"`
	RequestId          uuid.UUID            `json:"request_id"`
}

type ProductPerMovement struct {
	Id             string    `json:"id"`
	ProductID      uuid.UUID `json:"product_id"`
	Count          int       `json:"count"`
	MovementId     uuid.UUID `json:"movement_id"`
	DateLimit      time.Time `json:"date_limit"`
	MovementTypeId int       `json:"movement_type"`
	CreatedAt      time.Time `json:"created_at"`
}

// ---- Procesar Documento ----
type RequestProcessEvent struct {
	BaseEvent
	RequestID       uuid.UUID `json:"request_id"`
	ClientAccountId uuid.UUID `json:"client_account_id"`
	TypeIngress     int       `json:"type_ingress"`
}

type ProductEvent struct {
	BaseEvent
	ProductoID      string    `json:"producto_id"`
	NombreProducto  string    `json:"nombre_producto"`
	ClienteID       string    `json:"cliente_id"`
	Cantidad        int       `json:"cantidad"`
	Signo           int       `json:"signo"` // 1 ingreso, -1 egreso
	Fecha           time.Time `json:"fecha"`
	SolicitudId     string    `json:"solicitud_id"`
	StatusSolicitud string    `json:"status_solicitud"`
	TipoMovimiento  string    `json:"tipo_movimiento"`
}
