package httpserver

import (
	"log"
	"net/http"

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

func NewRouter(s3Config config.UploadService, db *gorm.DB, connMQ *amqp.Connection, channel *amqp.Channel, region string) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.Logger, middleware.Recoverer)
	h := handlers.NewStatusHandler()
	s3Svc := s3.NewS3Svs(s3.S3config{UploadService: s3Config})
	eventService := eventservice.NewMQPublisher(channel, connMQ)
	textractService := textract.NewTextractService(region)
	requestService := request.NewRequestService(db, s3Svc, eventService, textractService)
	handleRequest := &handlers.RequestHandler{Service: requestService}

	configListener(connMQ, channel, requestService)
	initHealthRoutes(r, h)

	initRequestRoutes(r, handleRequest)

	return r
}

func initHealthRoutes(r *chi.Mux, h *handlers.StatusHandler) {
	r.Get(HealthPath, func(w http.ResponseWriter, r *http.Request) {
		h.Health(w)
	})
}

func initRequestRoutes(r *chi.Mux, requestService *handlers.RequestHandler) {
	r.Route(RequestBasePath, func(r chi.Router) {
		r.Get("/", requestService.List)
		r.Post("/", requestService.Create)
		r.Get("/{id}", requestService.Get)
	})
}

func configListener(connMQ *amqp.Connection, ch *amqp.Channel, requestService request.RequestService) {
	listener := consumer.NewListener(connMQ, ch, "service.queue", requestService)

	if err := listener.SetupListener([]string{eventservice.REQUEST_TOPIC, eventservice.MOVEMENT_TOPIC}); err != nil {
		log.Fatalf("❌ Error en setup listener: %v", err)
	}

	if err := listener.StartListening(); err != nil {
		log.Fatalf("❌ Error en listener: %v", err)
	}
}
