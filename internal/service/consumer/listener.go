package consumer

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/stock-ahora/api-stock/internal/service/eventservice"
	"github.com/streadway/amqp"
)

type Listener struct {
	channel    *amqp.Channel
	connection *amqp.Connection
	queueName  string
}

func NewListener(conn *amqp.Connection, ch *amqp.Channel, queue string) *Listener {
	return &Listener{
		channel:    ch,
		connection: conn,
		queueName:  queue,
	}
}

func (l *Listener) SetupListener(routingKeys []string) error {
	// Declarar cola (durable)
	q, err := l.channel.QueueDeclare(
		l.queueName,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("error creando queue: %w", err)
	}

	// Hacer binding de cada routing key
	for _, rk := range routingKeys {
		if err := l.channel.QueueBind(
			q.Name,
			rk,
			eventservice.ExchangeName, // mismo exchange que el publisher
			false,
			nil,
		); err != nil {
			return fmt.Errorf("error en binding de routing key %s: %w", rk, err)
		}
	}

	return nil
}

func (l *Listener) StartListening() error {
	msgs, err := l.channel.Consume(
		l.queueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("error creando consumidor: %w", err)
	}

	go func() {
		for d := range msgs {
			log.Printf("📥 Recibido mensaje con routing key: %s", d.RoutingKey)
			switch d.RoutingKey {
			case "movement.generated":
				var event eventservice.MovementEvent
				if err := json.Unmarshal(d.Body, &event); err != nil {
					log.Printf("❌ error parseando MovementEvent: %v", err)
					continue
				}
				handleMovement(event)

			case "Request.process":
				var event eventservice.RequestProcessEvent
				if err := json.Unmarshal(d.Body, &event); err != nil {
					log.Printf("❌ error parseando RequestProcessEvent: %v", err)
					continue
				}
				handleRequestProcess(event)

			default:
				log.Printf("⚠️ No hay handler para routing key: %s", d.RoutingKey)
			}
		}
	}()

	log.Println("👂 Listener iniciado, esperando mensajes...")
	return nil
}
