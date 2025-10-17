package eventservice

import (
	"crypto/tls"
	"crypto/x509"
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
		rootCAs, _ := x509.SystemCertPool()
		tlsCfg := &tls.Config{RootCAs: rootCAs}

		newConn, err := amqp.DialTLS(p.urlConnection, tlsCfg)
		if err != nil {
			return err
		}
		defer func() {
			if newConn != nil {
				newConn.Close()
			}
		}()

		newCh, err := newConn.Channel()
		if err != nil {
			return err
		}

		// asignar y evitar que el defer cierre la conexión
		p.connection = newConn
		p.Channel = newCh
		newConn = nil
	}

	return p.Channel.Publish(
		ExchangeName,
		routingKey,
		false, // mandatory
		false, // immediate (deprecado)
		pub,
	)
}
