package eventservice

import (
	"time"

	"github.com/google/uuid"
	"github.com/streadway/amqp"
)

type EventPublisher interface {
	PublishDocument(e RequestProcessEvent) error
}

const MOVEMENT_TOPIC = "movement.generated"
const REQUEST_TOPIC = "Request.process"

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

	rk := REQUEST_TOPIC

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
