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

// ---- Procesar Documento ----
type RequestProcessEvent struct {
	BaseEvent
	RequestID       uuid.UUID `json:"request_id"`
	ClientAccountId uuid.UUID `json:"client_account_id"`
	TypeIngress     int       `json:"type_ingress"`
}
