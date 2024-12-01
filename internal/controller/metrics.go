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
			response.BadRequestError(w, convErr.Error())
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
		//	response.BadRequestError(w, http.StatusText(http.StatusUnsupportedMediaType))
		//}

		metric := req.PathValue("metric")
		value, convErr := strconv.ParseFloat(req.PathValue("value"), 10)

		if convErr != nil {
			response.BadRequestError(w, convErr.Error())
		}

		ms := storage.GetStorage()

		ms.SetGaugeValue(metric, value)

		response.Success(w)
	}
}

func GetCounter() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		metric := req.PathValue("metric")

		ms := storage.GetStorage()

		val, hasVal := ms.GetCounterValue(metric)

		if !hasVal {
			http.NotFound(w, req)
			return
		}

		response.Send(w, http.StatusOK, strconv.FormatInt(val, 10))
	}
}

func GetGauge() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		metric := req.PathValue("metric")

		ms := storage.GetStorage()

		val, hasVal := ms.GetGaugeValue(metric)

		if !hasVal {
			http.NotFound(w, req)
			return
		}

		response.Send(w, http.StatusOK, strconv.FormatFloat(val, 'g', -1, 64))
	}
}
