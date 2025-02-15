package middleware

import (
	"net/http"
	"slices"
	"strings"

	"github.com/OleG2e/collector/pkg/logging"

	"github.com/OleG2e/collector/internal/network"
	"github.com/OleG2e/collector/internal/response"
)

func AllowedMetricsOnly(l *logging.ZapLogger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if hasAllowedMetricByURLPath(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			form, decodeErr := network.NewFormByRequest(r)
			hasAllowedMetric := form.IsGaugeType() || form.IsCounterType()
			if decodeErr != nil || !hasAllowedMetric {
				response.New(r.Context(), l).BadRequestError(w, http.StatusText(http.StatusBadRequest))
				return
			}
			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}

func hasAllowedMetricByURLPath(path string) bool {
	allowedMetricTypes := []string{"gauge", "counter"}

	return slices.ContainsFunc(allowedMetricTypes, func(m string) bool {
		return strings.Contains(path, m)
	})
}
