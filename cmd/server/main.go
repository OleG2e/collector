package main

import (
	"net/http"

	"github.com/OleG2e/collector/internal/middleware"

	"github.com/OleG2e/collector/internal/container"
	"github.com/OleG2e/collector/internal/controller"
	"github.com/OleG2e/collector/internal/response"
	"github.com/go-chi/chi/v5"
)

func main() {
	container.InitContainer()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.AllowedMetricsOnly)

	r.Post("/update/", controller.UpdateMetric())
	r.Post("/value/", controller.GetMetric())

	r.Get("/value/counter/{metric}", controller.GetCounter())
	r.Get("/value/gauge/{metric}", controller.GetGauge())

	r.Post("/update/counter/{metric}/{value}", controller.UpdateCounter())
	r.Post("/update/gauge/{metric}/{value}", controller.UpdateGauge())
	r.Post("/update/counter/", r.NotFoundHandler())
	r.Post("/update/gauge/", r.NotFoundHandler())

	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		response.BadRequestError(w, http.StatusText(http.StatusBadRequest))
	})

	address := container.GetConfig().GetAddress()

	if err := http.ListenAndServe(address, r); err != nil {
		container.GetLogger().Sugar().Panic("server panic error", err)
	}
}
