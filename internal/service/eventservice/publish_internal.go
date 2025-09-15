package eventservice

import (
	"encoding/json"
	"time"

	"github.com/streadway/amqp"
)

func publishJSON(ch *amqp.Channel, routingKey string, msg interface{}, headers amqp.Table) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	pub := amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent, // persistente (2)
		Timestamp:    time.Now(),
		Headers:      headers,
		Body:         body,
	}

	return ch.Publish(
		ExchangeName,
		routingKey,
		false, // mandatory
		false, // immediate (deprecado)
		pub,
	)
}
