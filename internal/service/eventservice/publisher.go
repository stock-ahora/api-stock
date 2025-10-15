package eventservice

import (
	"log"
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
	Channel       *amqp.Channel
	connection    *amqp.Connection
	urlConnection string
}

func NewMQPublisher(Channel *amqp.Channel, connection *amqp.Connection, urlConnection string) *MQPublisher {
	return &MQPublisher{Channel: Channel, connection: connection, urlConnection: urlConnection}
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

	log.Println("Publishing request event:", e)

	return p.publishJSON(rk, e, headers)
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

	log.Println("Publishing movement event:", e)

	return p.publishJSON(rk, e, headers)
}

func newUUID() string {

	return uuid.New().String()
}
