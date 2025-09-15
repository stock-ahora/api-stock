package eventservice

import "time"

type BaseEvent struct {
	EventID       string            `json:"event_id"`
	EventType     string            `json:"event_type"`
	OccurredAt    time.Time         `json:"occurred_at"`
	Version       string            `json:"version"`
	CorrelationID string            `json:"correlation_id,omitempty"`
	Source        string            `json:"source,omitempty"`
	Headers       map[string]string `json:"headers,omitempty"`
}

// ---- Movimiento ----
type MovementEvent struct {
	BaseEvent
	MovementID string            `json:"movement_id"`
	UserID     string            `json:"user_id"`
	Action     string            `json:"action"` // ej: create, cancel
	Amount     float64           `json:"amount"`
	Currency   string            `json:"currency"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// ---- Procesar Documento ----
type DocumentProcessEvent struct {
	BaseEvent
	DocumentID  string            `json:"document_id"`
	Operation   string            `json:"operation"` // ej: process, validate
	Bucket      string            `json:"bucket,omitempty"`
	ObjectKey   string            `json:"object_key,omitempty"`
	ContentType string            `json:"content_type,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}
