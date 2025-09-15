package eventservice

import "github.com/streadway/amqp"

const (
	ExchangeName      = "events.topic"
	ExchangeKindTopic = "topic"
)

func EnsureTopology(ch *amqp.Channel) error {
	return ch.ExchangeDeclare(
		ExchangeName,
		ExchangeKindTopic,
		true,  // durable
		false, // auto-delete
		false, // internal
		false, // no-wait
		nil,   // args
	)
}
