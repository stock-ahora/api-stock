package eventservice

import (
	"time"

	"github.com/google/uuid"
	"github.com/streadway/amqp"
)

type EventPublisher interface {
	PublishDocument(e RequestProcessEvent) error
}

const MovementTopic = "movement.generated"
const RequestTopic = "Request.process"

type MQPublisher struct {
	Channel    *amqp.Channel
	connection *amqp.Connection
}

func NewMQPublisher(Channel *amqp.Channel, connection *amqp.Connection) *MQPublisher {
	return &MQPublisher{Channel: Channel, connection: connection}
}

func (p *MQPublisher) PublishRequest(e RequestProcessEvent) error {
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

	rk := RequestTopic

	headers := amqp.Table{
		"type":          "document",
		"version":       e.Version,
		"correlationId": e.CorrelationID,
	}

	return publishJSON(p.Channel, rk, e, headers)
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

	rk := MovementTopic

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
