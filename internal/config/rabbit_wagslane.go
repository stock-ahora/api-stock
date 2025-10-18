package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/wagslane/go-rabbitmq"
)

func mqURL(mq MQConfig) string {
	return fmt.Sprintf("amqps://%s:%s@%s:%d/%s", mq.User, mq.Password, mq.Host, mq.Port, mq.VHost)
}

func tlsConfig() *tls.Config {
	rootCAs, _ := x509.SystemCertPool()
	return &tls.Config{
		RootCAs:    rootCAs,
		MinVersion: tls.VersionTLS12,
	}
}

// Conexión administrada (reconexión automática)
func RabbitConn(mq MQConfig) (*rabbitmq.Conn, string, error) {
	url := mqURL(mq)
	conn, err := rabbitmq.NewConn(
		url,
		rabbitmq.WithConnectionOptionsConfig(rabbitmq.Config{
			TLSClientConfig: tlsConfig(),
			Heartbeat:       2 * time.Second,
			Locale:          "en_US",
			Dial:            amqp.DefaultDial(30 * time.Second),
		}),
		rabbitmq.WithConnectionOptionsLogging,
		rabbitmq.WithConnectionOptionsReconnectInterval(5*time.Second),
	)
	return conn, url, err
}

// Publisher sobre esa conexión
func RabbitPublisher(mq MQConfig) (*rabbitmq.Publisher, string, error) {
	conn, url, err := RabbitConn(mq)
	if err != nil {
		return nil, url, err
	}
	pub, err := rabbitmq.NewPublisher(
		conn,
		rabbitmq.WithPublisherOptionsLogging,
		rabbitmq.WithPublisherOptionsConfirm, // publisher confirms
	)
	return pub, url, err
}
