package rest

import (
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"collector/internal/core/domain"
	"collector/pkg/network"
)

func AllowedMetricsOnly(
	resp *network.Response,
	logger *slog.Logger,
) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(writer http.ResponseWriter, req *http.Request) {
			if hasAllowedMetricByURLPath(req.URL.Path) {
				next.ServeHTTP(writer, req)

				return
			}

			form, decodeErr := domain.NewFormByRequest(req)
			hasAllowedMetric := form.IsGaugeType() || form.IsCounterType()

			if decodeErr != nil {
				logger.WarnContext(
					req.Context(),
					"decode error",
					slog.Any("error", decodeErr),
					slog.Any("requestInfo", network.NewRequestInfo(req)),
				)
				resp.BadRequestError(writer, http.StatusText(http.StatusBadRequest))

				return
			}

			if !hasAllowedMetric {
				logger.WarnContext(
					req.Context(),
					"not allowed metric",
					slog.Any("requestInfo", network.NewRequestInfo(req)),
				)
				resp.BadRequestError(writer, http.StatusText(http.StatusBadRequest))

				return
			}

			next.ServeHTTP(writer, req)
		}

		return http.HandlerFunc(fn)
	}
}

func hasAllowedMetricByURLPath(path string) bool {
	allowedMetricTypes := []string{string(domain.MetricTypeGauge), string(domain.MetricTypeCounter)}

	return slices.ContainsFunc(allowedMetricTypes, func(metric string) bool {
		return strings.Contains(path, metric)
	})
}
