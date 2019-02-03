package api

import (
	chiprometheus "github.com/766b/chi-prometheus"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/nburunova/taxi-backend-sample/src/infrastructure/log"
)

// Router represents HTTP route multiplexer
type Router struct {
	*chi.Mux
}

// NewTaxiRouter возвращает роутер со всеми middleware
func NewTaxiRouter(logger *log.StructuredLogger) *Router {
	router := &Router{chi.NewRouter()}
	router.Use(chiprometheus.NewMiddleware("taxa", 500, 1000, 2000, 4000))
	router.Use(middleware.RequestID)
	router.Use(middleware.RequestLogger(logger))
	router.Use(middleware.Recoverer)

	return router
}

// NewCommonRouter возвращает роутер только с одним middleware (recoverer)
func NewCommonRouter(logger *log.StructuredLogger) *Router {
	router := &Router{chi.NewRouter()}
	router.Use(middleware.Recoverer)
	router.Mount("/debug", middleware.Profiler())
	return router

}
