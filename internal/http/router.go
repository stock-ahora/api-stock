package httpserver

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/stock-ahora/api-stock/internal/config"
	"github.com/stock-ahora/api-stock/internal/service/Etl_service"
	"github.com/stock-ahora/api-stock/internal/service/eventservice"
	"github.com/stock-ahora/api-stock/internal/service/movement"
	"github.com/stock-ahora/api-stock/internal/service/request"
	"github.com/stock-ahora/api-stock/internal/service/s3"
	"github.com/stock-ahora/api-stock/internal/service/stock"
	"github.com/stock-ahora/api-stock/internal/service/textract"
	"github.com/wagslane/go-rabbitmq"
	"gorm.io/gorm"

	"github.com/stock-ahora/api-stock/internal/http/handlers"
)

const APIBasePath = "/prod/api/v1/stock"
const S3BasePath = APIBasePath + "/s3"
const RequestBasePath = APIBasePath + "/request"
const HealthPath = "/api/v1" + "/health"
const MovementPath = APIBasePath + "/movement"
const ChatBot = APIBasePath + "/chatbot"
const DashboardPath = "/prod/api/v1" + "/dashboard"

func NewRouter(s3Config config.UploadService, db *gorm.DB, dbStarts *gorm.DB, _ any, _ any, region string, _ string, mqConfig config.MQConfig) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // ← acepta cualquier origen
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"}, // ← acepta cualquier header
		ExposedHeaders:   []string{"*"},
		AllowCredentials: false,
		MaxAge:           300, // cache del preflight
	}))

	r.Use(middleware.RequestID, middleware.Recoverer)
	h := handlers.NewStatusHandler()
	s3Svc := s3.NewS3Svs(s3.S3config{UploadService: s3Config})
	stockSvc := stock.NewStockService(db)
	movementSvc := movement.NewMovementService(db)

	pub, urlConnectionMQ, err := config.RabbitPublisher(mqConfig)
	if err != nil {
		log.Printf("skip publish: MQ unavailable")
		pub = nil // degradamos, NO panic
	}
	eventService := eventservice.NewMQPublisher(pub, urlConnectionMQ)

	textractService := textract.NewTextractService(region)
	requestService := request.NewRequestService(db, s3Svc, eventService, textractService)
	handleRequest := &handlers.RequestHandler{Service: requestService}
	handleStock := &handlers.StockHandler{Service: stockSvc}
	handleChatBot := &handlers.BedbrockHandler{Db: db}
	habdleDashboard := &handlers.DashboardHandler{Db: dbStarts}
	movementHandler := &handlers.MovementHandler{Service: movementSvc}
	etlService := Etl_service.EtlService{Db: dbStarts}

	configListener(etlService, requestService, mqConfig)
	initHealthRoutes(r, h)

	initRequestRoutes(r, handleRequest)
	initStockRoutes(r, handleStock)
	initMovementRoutes(r, movementHandler)
	initChatRoutes(r, handleChatBot)
	initDashboardRoutes(r, habdleDashboard)

	initTestGateway(r, *h)

	return r
}

func initDashboardRoutes(r *chi.Mux, dashboard *handlers.DashboardHandler) {

	r.Route(DashboardPath, func(r chi.Router) {
		r.Get("/", dashboard.Get)
	})

}

func initChatRoutes(r *chi.Mux, bot *handlers.BedbrockHandler) {
	r.Route(ChatBot, func(r chi.Router) {
		r.Get("/", bot.ConsultaProductos)
	})
}

func initMovementRoutes(r *chi.Mux, handler *handlers.MovementHandler) {
	r.Route(MovementPath, func(r chi.Router) {
		r.Get("/{id}", handler.List)
	})
}

func initStockRoutes(r *chi.Mux, requestService *handlers.StockHandler) {

	r.Route(APIBasePath, func(r chi.Router) {
		r.Get("/", requestService.List)
		r.Get("/{id}", requestService.Get)
	})

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

func configListener(etlService Etl_service.EtlService, requestService request.RequestService, mqCfg config.MQConfig) {
	go listenRequestQueue(requestService, mqCfg)
	go listenEtlQueue(etlService, mqCfg)
}

func listenRequestQueue(requestService request.RequestService, mqCfg config.MQConfig) {
	conn, _, err := config.RabbitConn(mqCfg)
	if err != nil {
		log.Printf("MQ publisher error: %v", err)
		return
	}

	const exchange = eventservice.ExchangeName
	const routingKey = eventservice.RequestTopic
	const queueName = "service.queue"

	args := rabbitmq.WithConsumerOptionsQueueArgs(map[string]interface{}{
		"x-dead-letter-exchange":    "events.failover",
		"x-dead-letter-routing-key": "Request.process.failover",
		"x-message-ttl":             2000,
	})

	consumer, err := rabbitmq.NewConsumer(
		conn, queueName,
		rabbitmq.WithConsumerOptionsQueueDurable,
		rabbitmq.WithConsumerOptionsExchangeName(exchange),
		rabbitmq.WithConsumerOptionsExchangeKind("topic"),
		rabbitmq.WithConsumerOptionsRoutingKey(routingKey),
		rabbitmq.WithConsumerOptionsQOSPrefetch(5),
		rabbitmq.WithConsumerOptionsConcurrency(5),
		args,
	)
	if err != nil {
		log.Fatalf("❌ NewConsumer: %v", err)
	}

	err = consumer.Run(func(d rabbitmq.Delivery) rabbitmq.Action {
		var evt eventservice.RequestProcessEvent
		if err := json.Unmarshal(d.Body, &evt); err != nil {
			return rabbitmq.NackDiscard
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := requestService.ProcessCtx(ctx, evt.RequestID, evt.ClientAccountId, evt.TypeIngress); err != nil {
			return rabbitmq.NackRequeue
		}
		return rabbitmq.Ack
	})
	if err != nil {
		log.Fatalf("❌ Error en consumer.Run: %v", err)
	}
}

func listenEtlQueue(etlService Etl_service.EtlService, mqCfg config.MQConfig) {
	conn, _, err := config.RabbitConn(mqCfg)
	if err != nil {
		log.Printf("MQ publisher error: %v", err)
		return
	}

	const exchange = eventservice.ExchangeName
	const routingKey = eventservice.EtlProduct // p. ej. "etl.product.created"
	const queueName = routingKey

	consumer, err := rabbitmq.NewConsumer(
		conn, queueName,
		rabbitmq.WithConsumerOptionsExchangeName(exchange),
		rabbitmq.WithConsumerOptionsExchangeKind("topic"),
		rabbitmq.WithConsumerOptionsRoutingKey(routingKey),
		rabbitmq.WithConsumerOptionsQOSPrefetch(1),
		rabbitmq.WithConsumerOptionsConcurrency(1),
		rabbitmq.WithConsumerOptionsConsumerName("etl-product-consumer"),
	)
	if err != nil {
		log.Fatalf("❌ NewConsumer: %v", err)
	}

	err = consumer.Run(func(d rabbitmq.Delivery) rabbitmq.Action {
		var evt eventservice.ProductEvent
		if err := json.Unmarshal(d.Body, &evt); err != nil {
			log.Printf("❌ ETL payload inválido: %v", err)
			return rabbitmq.NackDiscard
		}

		etlService.StartETLConsumer(evt)

		log.Printf("✅ ETL product procesado correctamente: %s", evt.ProductoID)
		return rabbitmq.Ack
	})
	if err != nil {
		log.Fatalf("❌ Error en consumer.Run (ETL): %v", err)
	}
}
