package controller

import (
	"context"
	"net/http"

	"github.com/OleG2e/collector/internal/config"
	"github.com/OleG2e/collector/internal/middleware"
	"github.com/OleG2e/collector/internal/response"
	"github.com/OleG2e/collector/internal/storage"

	"github.com/OleG2e/collector/pkg/logging"
	"github.com/go-chi/chi/v5"
)

type Controller struct {
	l        *logging.ZapLogger
	ctx      context.Context
	router   chi.Router
	response *response.Response
	ms       *storage.MemStorage
	conf     *config.ServerConfig
}

func New(logger *logging.ZapLogger, ctx context.Context, ms *storage.MemStorage, conf *config.ServerConfig) *Controller {
	return &Controller{l: logger, ms: ms, conf: conf, router: chi.NewRouter(), response: response.New(ctx, logger)}
}

func (c *Controller) Routes() *Controller {
	c.router.Use(middleware.Recover(c.l))
	c.router.Use(middleware.GzipMiddleware(c.l))
	c.router.Use(middleware.Logger(c.l))

	c.router.Group(func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, req *http.Request) {
			w.Header().Add("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
		})
		r.Get("/ping", c.PingDB())
		r.Post("/updates/", c.UpdateMetrics())
	})

	c.router.Route("/", func(r chi.Router) {
		r.Use(middleware.AllowedMetricsOnly(c.l))
		r.Post("/update/", c.UpdateMetric())
		r.Post("/value/", c.GetMetric())

		r.Get("/value/counter/{metric}", c.GetCounter())
		r.Get("/value/gauge/{metric}", c.GetGauge())

		r.Post("/update/counter/{metric}/{value}", c.UpdateCounter())
		r.Post("/update/gauge/{metric}/{value}", c.UpdateGauge())
		r.Post("/update/counter/", http.NotFound)
		r.Post("/update/gauge/", http.NotFound)

		r.Post("/", func(w http.ResponseWriter, req *http.Request) {
			c.response.Success(w)
		})
	})

	return c
}

func (c *Controller) ServeHTTP(serverConfig *config.ServerConfig) error {
	s := http.Server{
		Addr:     serverConfig.GetAddress(),
		ErrorLog: c.l.Std(),
		Handler:  c.router,
	}

	return s.ListenAndServe()
}
