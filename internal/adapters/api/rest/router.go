package rest

import (
	"log/slog"
	"net/http"
	"strconv"

	"collector/internal/adapters/store"
	"collector/internal/config"
	"collector/internal/core/domain"
	"collector/pkg/network"
	"github.com/go-chi/chi/v5"
)

const (
	metricReqPathName = "metric"
	valueReqPathName  = "value"
)

func NewRouter(
	st store.Store,
	logger *slog.Logger,
	conf *config.ServerConfig,
	resp *network.Response,
) *chi.Mux {
	router := chi.NewRouter()

	registerMiddlewares(router, logger, conf)
	registerMultipleMetricRoutes(st, router, logger, resp)
	registerSingleMetricRoutes(st, router, logger, resp)

	return router
}

func registerMultipleMetricRoutes(
	st store.Store,
	router *chi.Mux,
	logger *slog.Logger,
	resp *network.Response,
) chi.Router {
	return router.Group(func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Add("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
		})
		r.Get("/ping", pingDB(st, resp))
		r.Post("/updates/", updateMetrics(st, logger, resp))
	})
}

func registerMiddlewares(router *chi.Mux, logger *slog.Logger, conf *config.ServerConfig) {
	router.Use(RequestIDMiddleware)
	router.Use(LoggerMiddleware(logger))
	router.Use(RecoverMiddleware(logger))
	router.Use(GzipMiddleware(logger))
	router.Use(CheckSignMiddleware(conf, logger))
}

func registerSingleMetricRoutes(
	st store.Store,
	router *chi.Mux,
	logger *slog.Logger,
	resp *network.Response,
) {
	router.Route("/", func(r chi.Router) {
		r.Use(AllowedMetricsOnly(resp, logger))
		r.Post("/update/", updateMetric(st, logger, resp))
		r.Post("/value/", getMetric(st, logger, resp))

		r.Get("/value/counter/{metric}", getCounter(st, resp))
		r.Get("/value/gauge/{metric}", getGauge(st, resp))

		r.Post("/update/counter/{metric}/{value}", updateCounter(st, resp))
		r.Post("/update/gauge/{metric}/{value}", updateGauge(st, resp))
		r.Post("/update/counter/", http.NotFound)
		r.Post("/update/gauge/", http.NotFound)

		r.Post("/", func(w http.ResponseWriter, _ *http.Request) {
			resp.Success(w)
		})
	})
}

func updateMetric(st store.Store, logger *slog.Logger, resp *network.Response) http.HandlerFunc {
	return func(writer http.ResponseWriter, req *http.Request) {
		form, decodeErr := domain.NewFormByRequest(req)

		if decodeErr != nil {
			logger.ErrorContext(req.Context(), "decodeErr", slog.Any("error", decodeErr))
			resp.BadRequestError(writer, decodeErr.Error())

			return
		}

		if form.IsGaugeType() {
			st.GetMetrics().SetGaugeValue(form.ID, *form.Value)

			resp.Send(req.Context(), writer, http.StatusOK, form)

			return
		}

		if form.IsCounterType() {
			st.GetMetrics().AddCounterValue(form.ID, *form.Delta)

			val, hasVal := st.GetMetrics().GetCounterValue(form.ID)
			if !hasVal {
				http.NotFound(writer, req)

				return
			}

			form.Delta = &val

			resp.Send(req.Context(), writer, http.StatusOK, form)

			return
		}

		resp.BadRequestError(writer, "unknown metric type")
	}
}

func updateMetrics(st store.Store, logger *slog.Logger, resp *network.Response) http.HandlerFunc {
	return func(writer http.ResponseWriter, req *http.Request) {
		forms, decodeErr := domain.NewFormArrayByRequest(req)

		if decodeErr != nil {
			logger.ErrorContext(req.Context(), "decodeErr", slog.Any("error", decodeErr))
			resp.BadRequestError(writer, decodeErr.Error())

			return
		}

		if len(forms) == 0 {
			resp.BadRequestError(writer, "no metrics found")

			return
		}

		for _, form := range forms {
			if form.IsGaugeType() {
				st.GetMetrics().SetGaugeValue(form.ID, *form.Value)
			}

			if form.IsCounterType() {
				st.GetMetrics().AddCounterValue(form.ID, *form.Delta)
			}
		}

		resp.Success(writer)
	}
}

func pingDB(st store.Store, resp *network.Response) http.HandlerFunc {
	return func(writer http.ResponseWriter, _ *http.Request) {
		storeType := st.GetStoreType()

		if storeType != store.DBStoreType {
			resp.ServerError(writer, "connect to db doesn't exist")

			return
		}

		resp.Success(writer)
	}
}

func getMetric(st store.Store, logger *slog.Logger, resp *network.Response) http.HandlerFunc {
	return func(writer http.ResponseWriter, req *http.Request) {
		form, decodeErr := domain.NewFormByRequest(req)

		if decodeErr != nil {
			logger.ErrorContext(req.Context(), "decodeErr", slog.Any("error", decodeErr))
			resp.BadRequestError(writer, decodeErr.Error())

			return
		}

		if form.IsGaugeType() {
			value, _ := st.GetMetrics().GetGaugeValue(form.ID)
			form.Value = &value
			resp.Send(req.Context(), writer, http.StatusOK, form)

			return
		}

		if form.IsCounterType() {
			value, _ := st.GetMetrics().GetCounterValue(form.ID)
			form.Delta = &value
			resp.Send(req.Context(), writer, http.StatusOK, form)

			return
		}

		resp.BadRequestError(writer, "unknown metric type")
	}
}

func updateCounter(st store.Store, resp *network.Response) http.HandlerFunc {
	return func(writer http.ResponseWriter, req *http.Request) {
		metric := req.PathValue(metricReqPathName)
		value, convErr := strconv.ParseInt(req.PathValue(valueReqPathName), 10, 64)

		if convErr != nil {
			resp.BadRequestError(writer, convErr.Error())
		}

		st.GetMetrics().AddCounterValue(metric, value)

		resp.Success(writer)
	}
}

func updateGauge(st store.Store, resp *network.Response) http.HandlerFunc {
	return func(writer http.ResponseWriter, req *http.Request) {
		metric := req.PathValue(metricReqPathName)
		value, convErr := strconv.ParseFloat(req.PathValue(valueReqPathName), 64)

		if convErr != nil {
			resp.BadRequestError(writer, convErr.Error())
		}

		st.GetMetrics().SetGaugeValue(metric, value)

		resp.Success(writer)
	}
}

func getCounter(st store.Store, resp *network.Response) http.HandlerFunc {
	return func(writer http.ResponseWriter, req *http.Request) {
		metric := req.PathValue(metricReqPathName)

		val, hasVal := st.GetMetrics().GetCounterValue(metric)

		if !hasVal {
			http.NotFound(writer, req)

			return
		}

		resp.Send(req.Context(), writer, http.StatusOK, val)
	}
}

func getGauge(st store.Store, resp *network.Response) http.HandlerFunc {
	return func(writer http.ResponseWriter, req *http.Request) {
		metric := req.PathValue(metricReqPathName)

		val, hasVal := st.GetMetrics().GetGaugeValue(metric)

		if !hasVal {
			http.NotFound(writer, req)

			return
		}

		resp.Send(req.Context(), writer, http.StatusOK, val)
	}
}
