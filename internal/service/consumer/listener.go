package consumer

import (
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"

	"github.com/stock-ahora/api-stock/internal/service/eventservice"
	"github.com/stock-ahora/api-stock/internal/service/request"
	"github.com/streadway/amqp"
)

type Listener struct {
	channel        *amqp.Channel
	connection     *amqp.Connection
	queueName      string
	requestService request.RequestService
	workerCount    int
	urlConnection  string
}

func NewListener(conn *amqp.Connection, ch *amqp.Channel, queue string, requestService request.RequestService, workers int, urlConnection string) *Listener {
	if workers <= 0 {
		workers = runtime.NumCPU() // por defecto usa el número de CPUs
	}
	return &Listener{
		channel:        ch,
		connection:     conn,
		queueName:      queue,
		requestService: requestService,
		workerCount:    workers,
		urlConnection:  urlConnection,
	}
}

func (l *Listener) SetupListener(routingKeys []string) error {
	q, err := l.channel.QueueDeclare(
		l.queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("error creando queue: %w", err)
	}

	for _, rk := range routingKeys {
		if err := l.channel.QueueBind(
			q.Name,
			rk,
			eventservice.ExchangeName,
			false,
			nil,
		); err != nil {
			return fmt.Errorf("error en binding de routing key %s: %w", rk, err)
		}
	}

	// QoS → 1 mensaje por worker
	if err := l.channel.Qos(l.workerCount, 0, false); err != nil {
		return fmt.Errorf("error configurando QoS: %w", err)
	}

	return nil
}

func (l *Listener) StartListening() error {
	msgs, err := l.channel.Consume(
		l.queueName,
		"",
		false, // autoAck = false → manejamos ACK manual
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("error creando consumidor: %w", err)
	}

	// Worker pool
	wg := sync.WaitGroup{}
	for i := 0; i < l.workerCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for d := range msgs {
				log.Printf("📥 [Worker %d] Recibido mensaje con routing key: %s", id, d.RoutingKey)
				if err := l.handleMessage(d); err != nil {
					log.Printf("❌ error procesando mensaje: %v", err)
					_ = d.Nack(false, false) // requeue = true
					continue
				}
				_ = d.Ack(false)
			}
		}(i)
	}

	log.Printf("👂 Listener iniciado con %d workers, esperando mensajes...", l.workerCount)
	wg.Wait()
	return nil
}

func (l *Listener) handleMessage(d amqp.Delivery) error {
	switch d.RoutingKey {
	case "Request.process":
		var event eventservice.RequestProcessEvent
		if err := json.Unmarshal(d.Body, &event); err != nil {
			return fmt.Errorf("parseando RequestProcessEvent: %w", err)
		}
		handleRequestProcess(event, l.requestService)

	default:
		log.Printf("⚠️ No hay handler para routing key: %s", d.RoutingKey)
	}
	return nil
}

func (l *Listener) reconnectLoop() {
	connErrCh := l.connection.NotifyClose(make(chan *amqp.Error))
	chanErrCh := l.channel.NotifyClose(make(chan *amqp.Error))

	go func() {
		for {
			select {
			case err, ok := <-connErrCh:
				if !ok {
					log.Println("⚠️ conexión cerrada sin error explícito")
					return
				}
				log.Printf("❌ conexión cerrada: %v", err)
				l.reconnect()

			case err, ok := <-chanErrCh:
				if !ok {
					log.Println("⚠️ canal cerrado sin error explícito")
					return
				}
				log.Printf("❌ canal cerrado: %v", err)
				l.reconnect()
			}
		}
	}()
}

func (l *Listener) reconnect() {
	for {
		time.Sleep(5 * time.Second) // backoff
		if err := l.SetupListener([]string{eventservice.RequestTopic}); err == nil {
			log.Println("🔄 Reconexión exitosa, reiniciando consumo...")
			if err := l.StartListening(); err != nil {
				log.Printf("❌ error al reiniciar consumo: %v", err)
				continue
			}
			break
		}
		log.Println("⏳ Reintentando conexión...")
	}
}
