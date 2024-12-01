package response

import (
	"log"
	"net/http"
)

func setDefaultHeaders(writer http.ResponseWriter) {
	writer.Header().Add("Content-Type", "application/json")
}

func Success(writer http.ResponseWriter) {
	setDefaultHeaders(writer)
}

func BadRequestError(writer http.ResponseWriter, error string) {
	http.Error(writer, error, http.StatusBadRequest)
}

func setStatusCode(writer http.ResponseWriter, statusCode int) {
	writer.WriteHeader(statusCode)
}

func Send(writer http.ResponseWriter, statusCode int, data string) {
	setDefaultHeaders(writer)
	setStatusCode(writer, statusCode)

	_, err := writer.Write([]byte(data))
	if err != nil {
		log.Printf("error encoding response: %v", err)
		return
	}
}
