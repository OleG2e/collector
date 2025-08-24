package rest

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"collector/internal/config"
	"collector/internal/core/domain"
	"collector/pkg/hashing"
	"collector/pkg/network"
	"github.com/google/uuid"
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
	requestIDKey string
)

const RequestIDKey = requestIDKey("request_id")

func (resp *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := resp.ResponseWriter.Write(b)
	if err != nil {
		return 0, fmt.Errorf("write response error: %w", err)
	}

	resp.responseData.size += size

	return size, err
}

func (resp *loggingResponseWriter) WriteHeader(statusCode int) {
	resp.ResponseWriter.WriteHeader(statusCode)
	resp.responseData.status = statusCode
}

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (writer gzipWriter) Write(b []byte) (int, error) {
	size, err := writer.Writer.Write(b)
	if err != nil {
		return 0, fmt.Errorf("gzip writer response error: %w", err)
	}

	return size, err
}

func LoggerMiddleware(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(writer http.ResponseWriter, req *http.Request) {
			start := time.Now()

			uri := req.RequestURI
			method := req.Method

			responseData := &responseData{
				status: 0,
				size:   0,
			}
			logRespWriter := loggingResponseWriter{
				ResponseWriter: writer,
				responseData:   responseData,
			}

			next.ServeHTTP(&logRespWriter, req)

			duration := time.Since(start)

			logger.InfoContext(
				req.Context(),
				"request info",
				slog.String("uri", uri),
				slog.String("method", method),
				slog.Duration("duration", duration),
				slog.Int("status", responseData.status),
				slog.Int("size", responseData.size),
			)
		}

		return http.HandlerFunc(fn)
	}
}

func GzipMiddleware(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(writer http.ResponseWriter, req *http.Request) {
			contentEncoding := req.Header.Get("Content-Encoding")

			sendsGzip := strings.Contains(contentEncoding, "gzip")
			if sendsGzip {
				gzReader, gzipReadErr := gzip.NewReader(req.Body)
				if gzipReadErr != nil {
					logger.ErrorContext(
						req.Context(),
						"compress read error",
						slog.Any("error", gzipReadErr),
					)

					return
				}

				req.Body = gzReader

				defer func(gzReader *gzip.Reader) {
					gzReaderErr := gzReader.Close()
					if gzReaderErr != nil {
						logger.ErrorContext(
							req.Context(),
							"compress close read error",
							slog.Any("error", gzReaderErr),
						)
					}
				}(gzReader)
			}

			contentType := req.Header.Get("Content-Type")
			enableEncoding := strings.Contains(req.Header.Get("Accept-Encoding"), "gzip")

			if !enableEncoding || !isSupportedContentType(contentType) {
				next.ServeHTTP(writer, req)

				return
			}

			gzWriter, err := gzip.NewWriterLevel(writer, gzip.BestSpeed)
			if err != nil {
				logger.ErrorContext(req.Context(), "compress error", slog.Any("error", err))

				return
			}

			defer func(gz *gzip.Writer) {
				closeGZipErr := gz.Close()
				if closeGZipErr != nil {
					logger.ErrorContext(
						req.Context(),
						"close compress error",
						slog.Any("error", closeGZipErr),
					)
				}
			}(gzWriter)

			writer.Header().Set("Content-Encoding", "gzip")

			next.ServeHTTP(gzipWriter{ResponseWriter: writer, Writer: gzWriter}, req)
		}

		return http.HandlerFunc(fn)
	}
}

func RecoverMiddleware(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(writer http.ResponseWriter, req *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.ErrorContext(
						req.Context(),
						"recovered from panic",
						slog.Any("panic", rec),
					)
					http.Error(
						writer,
						http.StatusText(http.StatusInternalServerError),
						http.StatusInternalServerError,
					)
				}
			}()

			next.ServeHTTP(writer, req)
		}

		return http.HandlerFunc(fn)
	}
}

func CheckSignMiddleware(
	config *config.ServerConfig,
	logger *slog.Logger,
) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(writer http.ResponseWriter, req *http.Request) {
			hashKey := config.GetHashKey()
			if hashKey != "" {
				headerHash := req.Header.Get(domain.HashHeader)
				logger.InfoContext(
					req.Context(),
					"check sign",
					slog.String("headerHash", headerHash),
				)

				if headerHash != "" {
					var bodyBuffer, bodyData bytes.Buffer
					req.Body = io.NopCloser(io.TeeReader(req.Body, &bodyBuffer))
					_, readErr := bodyData.ReadFrom(req.Body)

					if readErr != nil {
						logger.ErrorContext(req.Context(), "sign error", slog.Any("error", readErr))

						return
					}

					req.Body = io.NopCloser(&bodyBuffer)

					hashBody := hashing.HashByKey(bodyData.String(), hashKey)
					if headerHash != hashBody {
						network.NewResponse(
							logger,
							config,
						).BadRequestError(writer, http.StatusText(http.StatusBadRequest))

						return
					}
				}
			}

			next.ServeHTTP(writer, req)
		}

		return http.HandlerFunc(fn)
	}
}

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		ctx = context.WithValue(ctx, RequestIDKey, uuid.New().String())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func isSupportedContentType(contentType string) bool {
	if contentType == "" {
		contentType = "text/html"
	}

	return strings.Contains(contentType, "application/json") ||
		strings.Contains(contentType, "text/html")
}
