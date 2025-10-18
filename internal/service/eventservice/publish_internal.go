package eventservice

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/wagslane/go-rabbitmq"
)

func (p *MQPublisher) publishJSON(routingKey string, msg any, headers map[string]any) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	t0 := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = p.pub.PublishWithContext(
		ctx, body, []string{routingKey},
		rabbitmq.WithPublishOptionsExchange(ExchangeName),
		rabbitmq.WithPublishOptionsPersistentDelivery,
		rabbitmq.WithPublishOptionsContentType("application/json"),
		rabbitmq.WithPublishOptionsHeaders(headers),
		rabbitmq.WithPublishOptionsMandatory, // útil para detectar no-route
	)
	if err != nil {
		log.Printf("❌ publish err rk=%s err=%v elapsed=%s", routingKey, err, time.Since(t0))
		return nil
	}
	return nil
}
