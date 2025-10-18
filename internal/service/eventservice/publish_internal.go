package eventservice

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/wagslane/go-rabbitmq"
)

func (p *MQPublisher) publishJSON(routingKey string, msg any, headers map[string]any, ctx context.Context) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	t0 := time.Now()

	err = p.pub.PublishWithContext(
		ctx, body, []string{routingKey},
		rabbitmq.WithPublishOptionsExchange(ExchangeName),
		rabbitmq.WithPublishOptionsPersistentDelivery,
		rabbitmq.WithPublishOptionsContentType("application/json"),
		rabbitmq.WithPublishOptionsHeaders(headers),
		rabbitmq.WithPublishOptionsMandatory, // útil para detectar no-route
		rabbitmq.WithPublishOptionsHeaders(map[string]any{
			"x-delay-sla": 2000, // opcional para logging
		}),
		// 👇 TTL por mensaje (en ms)
		rabbitmq.WithPublishOptionsExpiration("2000"),
	)
	if err != nil {
		log.Printf("❌ publish err rk=%s err=%v elapsed=%s", routingKey, err, time.Since(t0))
		return nil
	}
	return nil
}
