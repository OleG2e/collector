package response

import (
	"encoding/json"
	"net/http"

	"github.com/OleG2e/collector/internal/container"
)

func setDefaultHeaders(writer http.ResponseWriter) {
	writer.Header().Add("Content-Type", "application/json")
}

func Success(writer http.ResponseWriter) {
	setDefaultHeaders(writer)
}

func BadRequestError(writer http.ResponseWriter, e string) {
	http.Error(writer, e, http.StatusBadRequest)
}

func setStatusCode(writer http.ResponseWriter, statusCode int) {
	writer.WriteHeader(statusCode)
}

func Send(writer http.ResponseWriter, statusCode int, data any) {
	setDefaultHeaders(writer)
	setStatusCode(writer, statusCode)

	err := json.NewEncoder(writer).Encode(data)
	if err != nil {
		logger := container.GetLogger().Sugar()
		logger.Errorln("error encoding response", err)
		return
	}
}
