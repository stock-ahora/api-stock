package httpserver

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/stock-ahora/api-stock/internal/config"
	"gorm.io/gorm"

	"github.com/stock-ahora/api-stock/internal/http/handlers"
	"github.com/stock-ahora/api-stock/internal/service"
)

const APIBasePath = "/api/v1/stock"
const S3BasePath = APIBasePath + "/s3"
const RequestBasePath = APIBasePath + "/request"
const HealthPath = "/api/v1" + "/health"

func NewRouter(s3Config config.UploadService, db *gorm.DB) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.Logger, middleware.Recoverer)
	h := handlers.NewStatusHandler()
	s3Svc := service.NewS3Svs(service.S3config{UploadService: s3Config})
	requestService := service.NewRequestService(db, s3Svc)

	r.Get(HealthPath, func(w http.ResponseWriter, r *http.Request) {
		h.Health(w)
	})

	r.Route(S3BasePath, func(r chi.Router) {
		r.Post("/upload", s3Svc.HandleUpload)
	})

	r.Route(RequestBasePath, func(r chi.Router) {
		r.Post("/", requestService.Create)
	})

	return r
}
