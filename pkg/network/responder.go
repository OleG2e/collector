package network

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"collector/internal/config"
	"collector/internal/core/domain"
	"collector/pkg/hashing"
)

type Response struct {
	logger *slog.Logger
	conf   *config.ServerConfig
}

func NewResponse(logger *slog.Logger, conf *config.ServerConfig) *Response {
	return &Response{
		logger: logger,
		conf:   conf,
	}
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

func (resp *Response) Send(
	ctx context.Context,
	writer http.ResponseWriter,
	statusCode int,
	data any,
) {
	resp.setDefaultHeaders(writer)
	resp.setStatusCode(writer, statusCode)

	marshData, err := json.Marshal(data)
	if err != nil {
		resp.logger.ErrorContext(ctx, "error encoding response", slog.Any("error", err))

		return
	}

	if resp.conf.HasHashKey() {
		hashBody := hashing.HashByKey(string(marshData), resp.conf.GetHashKey())
		writer.Header().Add(domain.HashHeader, hashBody)
	}

	_, err = writer.Write(marshData)
	if err != nil {
		resp.logger.ErrorContext(ctx, "write response error", slog.Any("error", err))

		return
	}
}

func (resp *Response) setStatusCode(writer http.ResponseWriter, statusCode int) {
	writer.WriteHeader(statusCode)
}

func (resp *Response) setDefaultHeaders(writer http.ResponseWriter) {
	writer.Header().Add("Content-Type", "application/json")
}
