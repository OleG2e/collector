package controller

import (
	"github.com/OleG2e/collector/internal/response"
	"github.com/OleG2e/collector/internal/storage"
	"net/http"
	"strconv"
)

func UpdateCounter() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		metric := req.PathValue("metric")
		value, convErr := strconv.ParseInt(req.PathValue("value"), 10, 64)

		if convErr != nil {
			response.Error(w, convErr.Error())
		}

		ms := storage.GetStorage()

		ms.AddCounterValue(metric, value)

		response.Success(w)
	}
}

func UpdateGauge() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// todo: move to middleware
		//contentType := req.Header.Get("Content-Type")
		//
		//if contentType != "text/plain" {
		//	response.Error(w, http.StatusText(http.StatusUnsupportedMediaType))
		//}

		metric := req.PathValue("metric")
		value, convErr := strconv.ParseFloat(req.PathValue("value"), 10)

		if convErr != nil {
			response.Error(w, convErr.Error())
		}

		ms := storage.GetStorage()

		ms.SetGaugeValue(metric, value)

		response.Success(w)
	}
}

func BadRequestHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		response.Error(w, http.StatusText(http.StatusBadRequest))
	}
}
