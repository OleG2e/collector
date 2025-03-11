package response

import (
	"context"
	"encoding/json"
	"github.com/OleG2e/collector/internal/config"
	"github.com/OleG2e/collector/internal/network"
	"github.com/OleG2e/collector/pkg/hashing"
	"net/http"

	"github.com/OleG2e/collector/pkg/logging"
	"go.uber.org/zap"
)

type Response struct {
	l    *logging.ZapLogger
	conf *config.ServerConfig
}

func New(l *logging.ZapLogger, conf *config.ServerConfig) *Response {
	return &Response{
		l:    l,
		conf: conf,
	}
}

func (resp *Response) setDefaultHeaders(writer http.ResponseWriter) {
	writer.Header().Add("Content-Type", "application/json")
}

func (resp *Response) Success(writer http.ResponseWriter) {
	resp.setDefaultHeaders(writer)
}

func (resp *Response) BadRequestError(writer http.ResponseWriter, e string) {
	http.Error(writer, e, http.StatusBadRequest)
}

func (resp *Response) ServerError(writer http.ResponseWriter, e string) {
	http.Error(writer, e, http.StatusInternalServerError)
}

func (resp *Response) setStatusCode(writer http.ResponseWriter, statusCode int) {
	writer.WriteHeader(statusCode)
}

func (resp *Response) Send(ctx context.Context, writer http.ResponseWriter, statusCode int, data any) {
	resp.setDefaultHeaders(writer)
	resp.setStatusCode(writer, statusCode)

	marshData, err := json.Marshal(data)

	if err != nil {
		resp.l.ErrorCtx(ctx, "error encoding response", zap.Error(err))
		return
	}

	if resp.conf.HasHashKey() {
		hashBody := hashing.HashByKey(string(marshData), resp.conf.GetHashKey())
		writer.Header().Add(network.HashHeader, hashBody)
	}

	_, err = writer.Write(marshData)
	if err != nil {
		resp.l.ErrorCtx(ctx, "write response error", zap.Error(err))
		return
	}
}
