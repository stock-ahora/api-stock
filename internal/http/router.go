package httpserver

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/stock-ahora/api-stock/internal/config"
	"github.com/stock-ahora/api-stock/internal/service/eventservice"
	"github.com/stock-ahora/api-stock/internal/service/request"
	"github.com/stock-ahora/api-stock/internal/service/s3"
	"github.com/stock-ahora/api-stock/internal/service/textract"
	"github.com/wagslane/go-rabbitmq"
	"gorm.io/gorm"

	"github.com/stock-ahora/api-stock/internal/http/handlers"
)

const APIBasePath = "/api/v1/stock"
const S3BasePath = APIBasePath + "/s3"
const RequestBasePath = APIBasePath + "/request"
const HealthPath = "/api/v1" + "/health"

func NewRouter(s3Config config.UploadService, db *gorm.DB, _ any, _ any, region string, _ string, mqConfig config.MQConfig) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.Recoverer)
	h := handlers.NewStatusHandler()
	s3Svc := s3.NewS3Svs(s3.S3config{UploadService: s3Config})

	pub, urlConnectionMQ, err := config.RabbitPublisher(mqConfig)
	if err != nil {
		log.Fatalf("❌ Publisher MQ: %v", err)
	}
	eventService := eventservice.NewMQPublisher(pub, urlConnectionMQ)

	textractService := textract.NewTextractService(region)
	requestService := request.NewRequestService(db, s3Svc, eventService, textractService)
	handleRequest := &handlers.RequestHandler{Service: requestService}

	configListener(requestService, mqConfig)
	initHealthRoutes(r, h)

	initRequestRoutes(r, handleRequest)

	initTestGateway(r, *h)

	return r
}

func initHealthRoutes(r *chi.Mux, h *handlers.StatusHandler) {
	r.Get(HealthPath, func(w http.ResponseWriter, r *http.Request) {
		h.Health(w)
	})
}

func initTestGateway(r *chi.Mux, handler handlers.StatusHandler) {
	r.Get("/test-gateway", func(w http.ResponseWriter, r *http.Request) {
		if CheckNATGateway() {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("NAT Gateway is working"))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("NAT Gateway is not working"))
		}
	})
}

// /test-gateway
func initRequestRoutes(r *chi.Mux, requestService *handlers.RequestHandler) {
	r.Route(RequestBasePath, func(r chi.Router) {
		r.Get("/", requestService.List)
		r.Post("/", requestService.Create)
		r.Get("/{id}", requestService.Get)
	})
}

func CheckNATGateway() bool {
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get("https://www.google.com")
	if err != nil {
		log.Printf("❌ Error al conectar con Google (NAT Gateway puede estar fallando): %v", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		log.Printf("✅ NAT Gateway funcionando correctamente, conexión a Google exitosa")
		return true
	}

	log.Printf("❌ Error al conectar con Google, código de estado HTTP: %d", resp.StatusCode)
	return false
}

func configListener(requestService request.RequestService, mqCfg config.MQConfig) {
	go func() {
		// Conexión administrada para el consumer (con reconexión)
		conn, _, err := config.RabbitConn(mqCfg)
		if err != nil {
			log.Fatalf("❌ Consumer Conn MQ: %v", err)
		}

		const queueName = "service.queue"
		const routingKey = eventservice.RequestTopic
		const exchange = eventservice.ExchangeName

		consumerClient, err := rabbitmq.NewConsumer(
			conn,
			queueName,
			rabbitmq.WithConsumerOptionsLogging,
			rabbitmq.WithConsumerOptionsExchangeName(exchange),
			rabbitmq.WithConsumerOptionsExchangeKind("topic"),
			// 👇 sin QueueDeclare (no tocar la cola existente)
			rabbitmq.WithConsumerOptionsQueueDurable,
			rabbitmq.WithConsumerOptionsRoutingKey(routingKey),
			rabbitmq.WithConsumerOptionsQOSPrefetch(5),
			rabbitmq.WithConsumerOptionsConcurrency(5),
			rabbitmq.WithConsumerOptionsConsumerName("api-stock-consumer"),
		)
		if err != nil {
			log.Fatalf("❌ NewConsumer: %v", err)
		}

		// Run(handler) SIN opciones
		err = consumerClient.Run(func(d rabbitmq.Delivery) rabbitmq.Action {
			switch d.RoutingKey {
			case routingKey:
				var evt eventservice.RequestProcessEvent
				if err := json.Unmarshal(d.Body, &evt); err != nil {
					log.Printf("❌ parseando RequestProcessEvent: %v", err)
					return rabbitmq.NackRequeue
				}
				if err := requestService.Process(evt.RequestID, evt.ClientAccountId, evt.TypeIngress); err != nil {
					log.Printf("❌ error procesando request: %v", err)
					return rabbitmq.NackRequeue
				}
				return rabbitmq.Ack
			default:
				log.Printf("⚠️ No hay handler para routing key: %s", d.RoutingKey)
				return rabbitmq.Ack
			}
		})
		if err != nil {
			log.Fatalf("❌ Error en consumer.Run: %v", err)
		}
	}()
}
