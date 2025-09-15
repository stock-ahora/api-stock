package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

func NewRabbitMq(mq MQConfig) (*amqp.Connection, *amqp.Channel) {

	rootCAs, _ := x509.SystemCertPool()

	tlsConfig := &tls.Config{
		RootCAs: rootCAs,
	}

	url := fmt.Sprintf("amqps://%s:%s@%s:%s/%s", mq.User, mq.Password, mq.Host, mq.Port, mq.VHost)

	conn, err := amqp.DialTLS(url, tlsConfig)
	if err != nil {
		log.Fatalf("❌ Error conectando a RabbitMQ: %v", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("❌ Error creando canal: %v", err)
	}
	return conn, ch

}
