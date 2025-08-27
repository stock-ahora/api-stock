package httpserver

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/stock-ahora/api-stock/internal/http/handlers"
	"github.com/stock-ahora/api-stock/internal/repository"
	"github.com/stock-ahora/api-stock/internal/service"
)

func NewRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.Logger, middleware.Recoverer)

	// Dependencias (por ahora, repo en memoria)
	repo := repository.NewMemoryStockRepo()
	svc := service.NewStockService(repo)
	h := handlers.NewStockHandler(svc)

	// Rutas
	r.Get("/api/v1/health", h.Health)

	r.Route("/api/v1/stock", func(r chi.Router) {
		r.Get("/", h.List)
		r.Post("/", h.Create)
	})

	return r
}
