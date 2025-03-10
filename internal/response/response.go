package response

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/OleG2e/collector/pkg/logging"
	"go.uber.org/zap"
)

type Response struct {
	l *logging.ZapLogger
}

func New(l *logging.ZapLogger) *Response {
	return &Response{
		l: l,
	}
}

func setDefaultHeaders(writer http.ResponseWriter) {
	writer.Header().Add("Content-Type", "application/json")
}

func (resp *Response) Success(writer http.ResponseWriter) {
	setDefaultHeaders(writer)
}

func (resp *Response) BadRequestError(writer http.ResponseWriter, e string) {
	http.Error(writer, e, http.StatusBadRequest)
}

func (resp *Response) ServerError(writer http.ResponseWriter, e string) {
	http.Error(writer, e, http.StatusInternalServerError)
}

func setStatusCode(writer http.ResponseWriter, statusCode int) {
	writer.WriteHeader(statusCode)
}

func (resp *Response) Send(ctx context.Context, writer http.ResponseWriter, statusCode int, data any) {
	setDefaultHeaders(writer)
	setStatusCode(writer, statusCode)

	err := json.NewEncoder(writer).Encode(data)
	if err != nil {
		resp.l.ErrorCtx(ctx, "error encoding response", zap.Error(err))
		return
	}
}
