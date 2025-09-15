package eventservice

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/streadway/amqp"
)

type EventPublisher interface {
	PublishMovement(e MovementEvent) error
	PublishDocument(e DocumentProcessEvent) error
}

type MQPublisher struct {
	Channel    *amqp.Channel
	connection *amqp.Connection
}

func NewMQPublisher(Channel *amqp.Channel, connection *amqp.Connection) *MQPublisher {
	return &MQPublisher{Channel: Channel, connection: connection}
}

func (p *MQPublisher) PublishMovement(e MovementEvent) error {
	if e.EventID == "" {
		e.EventID = newUUID()
	}
	if e.EventType == "" {
		e.EventType = "movement"
	}
	if e.Version == "" {
		e.Version = "1"
	}
	if e.OccurredAt.IsZero() {
		e.OccurredAt = time.Now().UTC()
	}

	rk := "movement." + strings.ToLower(e.Action)

	headers := amqp.Table{
		"type":          "movement",
		"version":       e.Version,
		"correlationId": e.CorrelationID,
	}

	return publishJSON(p.Channel, rk, e, headers)
}

func (p *MQPublisher) PublishDocument(ch *amqp.Channel, e DocumentProcessEvent) error {
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

	rk := "document." + strings.ToLower(e.Operation)

	headers := amqp.Table{
		"type":          "document",
		"version":       e.Version,
		"correlationId": e.CorrelationID,
	}

	return publishJSON(p.Channel, rk, e, headers)
}

func newUUID() string {

	return uuid.New().String()
}
