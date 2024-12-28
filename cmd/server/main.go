package main

import (
	"net/http"

	"github.com/OleG2e/collector/internal/controller"
	"github.com/OleG2e/collector/internal/response"
	"github.com/OleG2e/collector/internal/storage"

	"github.com/OleG2e/collector/internal/middleware"

	"github.com/OleG2e/collector/internal/container"
	"github.com/go-chi/chi/v5"
)

func main() {
	container.InitServerContainer()

	storage.InitStorage()
	defer func(storage *storage.MemStorage) {
		err := storage.FlushStorage()
		if err != nil {
			container.GetLogger().Error(err)
		}
	}(storage.GetStorage())

	router := chi.NewRouter()
	router.Use(middleware.GzipMiddleware)
	router.Use(middleware.Logger)

	router.Group(func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, req *http.Request) {
			w.Header().Add("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
		})
	})

	router.Route("/", func(r chi.Router) {
		r.Use(middleware.AllowedMetricsOnly)
		r.Post("/update/", controller.UpdateMetric())
		r.Post("/value/", controller.GetMetric())

		r.Get("/value/counter/{metric}", controller.GetCounter())
		r.Get("/value/gauge/{metric}", controller.GetGauge())

		r.Post("/update/counter/{metric}/{value}", controller.UpdateCounter())
		r.Post("/update/gauge/{metric}/{value}", controller.UpdateGauge())
		r.Post("/update/counter/", router.NotFoundHandler())
		r.Post("/update/gauge/", router.NotFoundHandler())

		r.Post("/", func(w http.ResponseWriter, req *http.Request) {
			response.Success(w)
		})
	})

	address := container.GetServerConfig().GetAddress()

	if err := http.ListenAndServe(address, router); err != nil {
		container.GetLogger().Panic("server panic error", err)
	}
}
