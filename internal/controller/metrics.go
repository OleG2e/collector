package controller

import (
	"net/http"
	"strconv"

	"github.com/OleG2e/collector/internal/container"
	"github.com/OleG2e/collector/internal/network"
	"github.com/OleG2e/collector/internal/response"
	"github.com/OleG2e/collector/internal/storage"
)

func UpdateMetric() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		form, decodeErr := network.NewFormByRequest(r)
		logger := container.GetLogger()

		if decodeErr != nil {
			logger.Errorln("decodeErr", decodeErr)
			response.BadRequestError(w, decodeErr.Error())
			return
		}

		ms := storage.GetStorage()
		if form.IsGaugeType() {
			ms.SetGaugeValue(form.ID, *form.Value)

			syncStateLogger()

			response.Send(w, http.StatusOK, form)
			return
		}

		if form.IsCounterType() {
			ms.AddCounterValue(form.ID, *form.Delta)
			val, hasVal := ms.GetCounterValue(form.ID)
			if !hasVal {
				http.NotFound(w, r)
				return
			}
			form.Delta = &val

			syncStateLogger()

			response.Send(w, http.StatusOK, form)
			return
		}

		response.BadRequestError(w, "unknown metric type")
	}
}

func GetMetric() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		form, decodeErr := network.NewFormByRequest(r)
		logger := container.GetLogger()

		if decodeErr != nil {
			logger.Errorln("decodeErr", decodeErr)
			response.BadRequestError(w, decodeErr.Error())
			return
		}

		ms := storage.GetStorage()
		if form.IsGaugeType() {
			value, _ := ms.GetGaugeValue(form.ID)
			form.Value = &value
			response.Send(w, http.StatusOK, form)
			return
		}

		if form.IsCounterType() {
			value, _ := ms.GetCounterValue(form.ID)
			form.Delta = &value
			response.Send(w, http.StatusOK, form)
			return
		}
		response.BadRequestError(w, "unknown metric type")
	}
}

const metricReqPathName = "metric"
const valueReqPathName = "value"

func UpdateCounter() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		metric := req.PathValue(metricReqPathName)
		value, convErr := strconv.ParseInt(req.PathValue(valueReqPathName), 10, 64)

		if convErr != nil {
			response.BadRequestError(w, convErr.Error())
		}

		ms := storage.GetStorage()

		ms.AddCounterValue(metric, value)

		syncStateLogger()

		response.Success(w)
	}
}

func UpdateGauge() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		metric := req.PathValue(metricReqPathName)
		value, convErr := strconv.ParseFloat(req.PathValue(valueReqPathName), 64)

		if convErr != nil {
			response.BadRequestError(w, convErr.Error())
		}

		ms := storage.GetStorage()

		ms.SetGaugeValue(metric, value)

		syncStateLogger()

		response.Success(w)
	}
}

func GetCounter() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		metric := req.PathValue(metricReqPathName)

		ms := storage.GetStorage()

		val, hasVal := ms.GetCounterValue(metric)

		if !hasVal {
			http.NotFound(w, req)
			return
		}

		response.Send(w, http.StatusOK, val)
	}
}

func GetGauge() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		metric := req.PathValue(metricReqPathName)

		ms := storage.GetStorage()

		val, hasVal := ms.GetGaugeValue(metric)

		if !hasVal {
			http.NotFound(w, req)
			return
		}

		response.Send(w, http.StatusOK, val)
	}
}

func syncStateLogger() {
	storeInterval := container.GetServerConfig().GetStoreInterval()
	if storeInterval == 0 {
		err := storage.GetStorage().FlushStorage()
		if err != nil {
			container.GetLogger().Error(err)
		}
	}
}
