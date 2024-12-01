package middleware

import (
	"github.com/OleG2e/collector/internal/response"
	"net/http"
	"slices"
	"strings"
)

func AllowedMetricsOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowedMetricTypes := []string{"gauge", "counter"}

		hasAllowedMetric := slices.ContainsFunc(allowedMetricTypes, func(m string) bool {
			return strings.Contains(r.URL.Path, m)
		})

		if !hasAllowedMetric {
			response.BadRequestError(w, http.StatusText(http.StatusBadRequest))
			return
		}
		next.ServeHTTP(w, r)
	})
}
