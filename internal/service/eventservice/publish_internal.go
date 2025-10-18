package eventservice

import (
	"context"
	"encoding/json"
	"time"

	"github.com/wagslane/go-rabbitmq"
)

func (p *MQPublisher) publishJSON(routingKey string, msg any, headers map[string]any) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return p.pub.PublishWithContext(
		ctx,
		body,
		[]string{routingKey},
		rabbitmq.WithPublishOptionsExchange(ExchangeName),
		rabbitmq.WithPublishOptionsPersistentDelivery,
		rabbitmq.WithPublishOptionsContentType("application/json"),
		rabbitmq.WithPublishOptionsHeaders(headers),
	)
}
