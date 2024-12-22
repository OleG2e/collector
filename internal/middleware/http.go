package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/OleG2e/collector/internal/container"
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

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		logger := container.GetLogger().Sugar()
		logger.Infoln(
			"uri", uri,
			"method", method,
			"duration", duration,
			"status", responseData.status,
			"size", responseData.size,
		)
	})
}

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := container.GetLogger().Sugar()

		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			gzReader, gzipReadErr := gzip.NewReader(r.Body)
			if gzipReadErr != nil {
				logger.Errorln("compress read error", gzipReadErr)
				return
			}
			r.Body = gzReader
			defer func(gzReader *gzip.Reader) {
				gzReaderErr := gzReader.Close()
				if gzReaderErr != nil {
					logger.Errorln("compress close read error", gzReaderErr)
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
			logger.Errorln("compress error", err)
			return
		}
		defer func(gz *gzip.Writer) {
			closeGZipErr := gz.Close()
			if closeGZipErr != nil {
				logger.Errorln("close compress error", err)
			}
		}(gzWriter)

		w.Header().Set("Content-Encoding", "gzip")

		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gzWriter}, r)
	})
}

func isSupportedContentType(contentType string) bool {
	if contentType == "" {
		contentType = "text/html"
	}

	return strings.Contains(contentType, "application/json") || strings.Contains(contentType, "text/html")
}
