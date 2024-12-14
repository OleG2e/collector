package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"syscall"
	"time"

	"go.uber.org/zap"
)

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	fmt.Println(statusCode)
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger, _ := zap.NewDevelopment()

		defer func(logger *zap.Logger) {
			syncErr := logger.Sync()
			if syncErr != nil {
				if errors.Is(syncErr, syscall.EINVAL) {
					// Sync is not supported on os.Stderr / os.Stdout on all platforms.
					return
				}
				logger.Error("Failed to sync logger", zap.Error(syncErr))
			}
		}(logger)

		sugar := logger.Sugar()

		start := time.Now()

		uri := r.RequestURI
		method := r.Method

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		fmt.Println(lw.responseData)

		next.ServeHTTP(&lw, r)

		duration := time.Since(start)

		sugar.Infoln(
			"uri", uri,
			"method", method,
			"duration", duration,
			"status", responseData.status,
			"size", responseData.size,
		)
	})
}
