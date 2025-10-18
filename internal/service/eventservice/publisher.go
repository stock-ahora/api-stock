package eventservice

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/wagslane/go-rabbitmq"
)

type EventPublisher interface {
	PublishRequest(e RequestProcessEvent) error
	PublishMovements(e MovementsEvent) error
}

const MovementTopic = "movement.generated"
const RequestTopic = "Request.process.prod"

type MQPublisher struct {
	pub           *rabbitmq.Publisher
	urlConnection string
}

func NewMQPublisher(pub *rabbitmq.Publisher, urlConnection string) *MQPublisher {
	return &MQPublisher{pub: pub, urlConnection: urlConnection}
}

func (p *MQPublisher) PublishRequest(e RequestProcessEvent, ctx context.Context) error {
	if e.EventID == "" {
		e.EventID = newUUID()
	}
	if e.EventType == "" {
		e.EventType = "document"
	}
	if e.Version == "" {
		e.Version = "1"
	}
	if e.OccurredAt.IsZero() {
		e.OccurredAt = time.Now().UTC()
	}

	log.Println("Publishing request event:", e)
	return p.publishJSON(RequestTopic, e, map[string]any{
		"type":          "document",
		"version":       e.Version,
		"correlationId": e.CorrelationID,
	}, ctx)
}

func (p *MQPublisher) PublishMovements(e MovementsEvent) error {
	if e.EventID == "" {
		e.EventID = newUUID()
	}
	if e.EventType == "" {
		e.EventType = "document"
	}
	if e.Version == "" {
		e.Version = "1"
	}
	if e.OccurredAt.IsZero() {
		e.OccurredAt = time.Now().UTC()
	}

	log.Println("Publishing movement event:", e)
	return p.publishJSON(MovementTopic, e, map[string]any{
		"type":          "document",
		"version":       e.Version,
		"correlationId": e.CorrelationID,
	}, context.Background())
}

func newUUID() string { return uuid.New().String() }
