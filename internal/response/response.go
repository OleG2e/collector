package response

import (
	"net/http"
)

func setDefaultHeaders(writer http.ResponseWriter) {
	writer.Header().Add("Content-Type", "application/json")
}

func Success(writer http.ResponseWriter) {
	setDefaultHeaders(writer)
}

func Error(writer http.ResponseWriter, error string) {
	http.Error(writer, error, http.StatusBadRequest)
}
