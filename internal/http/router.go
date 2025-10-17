package httpserver

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/stock-ahora/api-stock/internal/config"
	"github.com/stock-ahora/api-stock/internal/service/consumer"
	"github.com/stock-ahora/api-stock/internal/service/eventservice"
	"github.com/stock-ahora/api-stock/internal/service/request"
	"github.com/stock-ahora/api-stock/internal/service/s3"
	"github.com/stock-ahora/api-stock/internal/service/textract"
	"github.com/streadway/amqp"
	"gorm.io/gorm"

	"github.com/stock-ahora/api-stock/internal/http/handlers"
)

const APIBasePath = "/api/v1/stock"
const S3BasePath = APIBasePath + "/s3"
const RequestBasePath = APIBasePath + "/request"
const HealthPath = "/api/v1" + "/health"

func NewRouter(s3Config config.UploadService, db *gorm.DB, connMQ *amqp.Connection, channel *amqp.Channel, region string, urlConnectionMQ string, mqConfig config.MQConfig) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.Logger, middleware.Recoverer)
	h := handlers.NewStatusHandler()
	s3Svc := s3.NewS3Svs(s3.S3config{UploadService: s3Config})
	eventService := eventservice.NewMQPublisher(channel, connMQ, urlConnectionMQ)
	textractService := textract.NewTextractService(region)
	requestService := request.NewRequestService(db, s3Svc, eventService, textractService)
	handleRequest := &handlers.RequestHandler{Service: requestService}

	configListener(connMQ, channel, requestService, urlConnectionMQ, mqConfig)
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

func configListener(_ *amqp.Connection, _ *amqp.Channel, requestService request.RequestService, urlConnectionMQ string, mqConfig config.MQConfig) {
	go func() {
		subConn, subCh, _ := config.NewRabbitMq(mqConfig)

		// (idempotente) declara topología en el canal del consumidor
		if err := eventservice.EnsureTopology(subCh); err != nil {
			log.Fatalf("❌ Error declarando topología (consumer): %v", err)
		}

		listener := consumer.NewListener(subConn, subCh, "service.queue", requestService, 5, urlConnectionMQ)

		if err := listener.SetupListener([]string{eventservice.RequestTopic}); err != nil {
			log.Fatalf("❌ Error en setup listener: %v", err)
		}

		// 🔁 Activa el loop de reconexión
		listener.ReconnectLoop()

		if err := listener.StartListening(); err != nil {
			log.Fatalf("❌ Error en listener: %v", err)
		}
	}()
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
