package eventservice

import amqp "github.com/rabbitmq/amqp091-go"

const (
	ExchangeKindTopic = "topic"
	ExchangeName      = "events.topic"
)

func EnsureTopology(ch *amqp.Channel) error {
	return ch.ExchangeDeclare(
		ExchangeName,
		ExchangeKindTopic,
		true,
		false,
		false,
		false,
		nil,
	)
}
