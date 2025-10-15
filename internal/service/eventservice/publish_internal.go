package eventservice

import (
	"encoding/json"
	"log"
	"time"

	"github.com/streadway/amqp"
)

func (p *MQPublisher) publishJSON(routingKey string, msg interface{}, headers amqp.Table) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	log.Println("Publishing JSON:", routingKey, body)

	pub := amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent, // persistente (2)
		Timestamp:    time.Now(),
		Headers:      headers,
		Body:         body,
	}

	if p.connection.IsClosed() {
		newConn, err := amqp.Dial(p.urlConnection)
		if err != nil {
			return err
		}
		newCh, err := newConn.Channel()
		if err != nil {
			return err
		}
		p.connection = newConn
		p.Channel = newCh
	}

	return p.Channel.Publish(
		ExchangeName,
		routingKey,
		false, // mandatory
		false, // immediate (deprecado)
		pub,
	)
}
