package middleware

import (
	"bytes"
	"compress/gzip"
	"github.com/OleG2e/collector/internal/config"
	"github.com/OleG2e/collector/internal/network"
	"github.com/OleG2e/collector/internal/response"
	"github.com/OleG2e/collector/pkg/hashing"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/OleG2e/collector/pkg/logging"
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
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func Logger(l *logging.ZapLogger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
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

			next.ServeHTTP(&lw, r)

			duration := time.Since(start)

			l.InfoCtx(
				r.Context(),
				"request info",
				zap.String("uri", uri),
				zap.String("method", method),
				zap.Duration("duration", duration),
				zap.Int("status", responseData.status),
				zap.Int("size", responseData.size),
			)
		}

		return http.HandlerFunc(fn)
	}
}

func GzipMiddleware(l *logging.ZapLogger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			contentEncoding := r.Header.Get("Content-Encoding")
			sendsGzip := strings.Contains(contentEncoding, "gzip")
			if sendsGzip {
				gzReader, gzipReadErr := gzip.NewReader(r.Body)
				if gzipReadErr != nil {
					l.ErrorCtx(r.Context(), "compress read error", zap.Error(gzipReadErr))
					return
				}
				r.Body = gzReader
				defer func(gzReader *gzip.Reader) {
					gzReaderErr := gzReader.Close()
					if gzReaderErr != nil {
						l.ErrorCtx(r.Context(), "compress close read error", zap.Error(gzReaderErr))
					}
				}(gzReader)
			}

			contentType := r.Header.Get("Content-Type")
			enableEncoding := strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")
			if !enableEncoding || !isSupportedContentType(contentType) {
				next.ServeHTTP(w, r)
				return
			}

			gzWriter, err := gzip.NewWriterLevel(w, gzip.BestSpeed)

			if err != nil {
				l.ErrorCtx(r.Context(), "compress error", zap.Error(err))
				return
			}
			defer func(gz *gzip.Writer) {
				closeGZipErr := gz.Close()
				if closeGZipErr != nil {
					l.ErrorCtx(r.Context(), "close compress error", zap.Error(closeGZipErr))
				}
			}(gzWriter)

			w.Header().Set("Content-Encoding", "gzip")

			next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gzWriter}, r)
		}

		return http.HandlerFunc(fn)
	}
}

func Recover(l *logging.ZapLogger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					l.PanicCtx(r.Context(), "recovered from panic", zap.Any("panic", rec))
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}

func CheckSign(config *config.ServerConfig, l *logging.ZapLogger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if config.HasHashKey() {
				headerHash := r.Header.Get(network.HashHeader)
				l.InfoCtx(r.Context(), "check sign", zap.String("headerHash", headerHash))
				if headerHash != "" {
					var bodyBuffer, bodyData bytes.Buffer
					r.Body = io.NopCloser(io.TeeReader(r.Body, &bodyBuffer))
					_, readErr := bodyData.ReadFrom(r.Body)
					if readErr != nil {
						l.ErrorCtx(r.Context(), "sign error", zap.Error(readErr))
						return
					}
					r.Body = io.NopCloser(&bodyBuffer)

					hashBody := hashing.HashByKey(bodyData.String(), config.GetHashKey())
					if headerHash != hashBody {
						response.New(l, config).BadRequestError(w, http.StatusText(http.StatusBadRequest))
						return
					}
				}
			}

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}

func isSupportedContentType(contentType string) bool {
	if contentType == "" {
		contentType = "text/html"
	}

	return strings.Contains(contentType, "application/json") || strings.Contains(contentType, "text/html")
}
