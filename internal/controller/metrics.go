package controller

import (
	"github.com/OleG2e/collector/internal/storage"
	"net/http"
	"strconv"

	"go.uber.org/zap"

	"github.com/OleG2e/collector/internal/network"
)

func (c *Controller) UpdateMetric() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		form, decodeErr := network.NewFormByRequest(r)

		if decodeErr != nil {
			c.l.ErrorCtx(r.Context(), "decodeErr", zap.Error(decodeErr))
			c.response.BadRequestError(w, decodeErr.Error())
			return
		}

		if form.IsGaugeType() {
			c.ms.SetGaugeValue(form.ID, *form.Value)

			c.syncStateLogger(r)

			c.response.Send(w, http.StatusOK, form)
			return
		}

		if form.IsCounterType() {
			c.ms.AddCounterValue(form.ID, *form.Delta)
			val, hasVal := c.ms.GetCounterValue(form.ID)
			if !hasVal {
				http.NotFound(w, r)
				return
			}
			form.Delta = &val

			c.syncStateLogger(r)

			c.response.Send(w, http.StatusOK, form)
			return
		}

		c.response.BadRequestError(w, "unknown metric type")
	}
}

func (c *Controller) PingDB() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		storeType := c.ms.GetStoreAlgo().GetStoreType()

		if storeType != storage.DbStoreType {
			c.response.ServerError(w, "connect to db doesn't exist")
			return
		}

		c.response.Success(w)
	}
}

func (c *Controller) GetMetric() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		form, decodeErr := network.NewFormByRequest(r)

		if decodeErr != nil {
			c.l.ErrorCtx(r.Context(), "decodeErr", zap.Error(decodeErr))
			c.response.BadRequestError(w, decodeErr.Error())
			return
		}

		if form.IsGaugeType() {
			value, _ := c.ms.GetGaugeValue(form.ID)
			form.Value = &value
			c.response.Send(w, http.StatusOK, form)
			return
		}

		if form.IsCounterType() {
			value, _ := c.ms.GetCounterValue(form.ID)
			form.Delta = &value
			c.response.Send(w, http.StatusOK, form)
			return
		}
		c.response.BadRequestError(w, "unknown metric type")
	}
}

const metricReqPathName = "metric"
const valueReqPathName = "value"

func (c *Controller) UpdateCounter() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		metric := req.PathValue(metricReqPathName)
		value, convErr := strconv.ParseInt(req.PathValue(valueReqPathName), 10, 64)

		if convErr != nil {
			c.response.BadRequestError(w, convErr.Error())
		}

		c.ms.AddCounterValue(metric, value)

		c.syncStateLogger(req)

		c.response.Success(w)
	}
}

func (c *Controller) UpdateGauge() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		metric := req.PathValue(metricReqPathName)
		value, convErr := strconv.ParseFloat(req.PathValue(valueReqPathName), 64)

		if convErr != nil {
			c.response.BadRequestError(w, convErr.Error())
		}

		c.ms.SetGaugeValue(metric, value)

		c.syncStateLogger(req)

		c.response.Success(w)
	}
}

func (c *Controller) GetCounter() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		metric := req.PathValue(metricReqPathName)

		val, hasVal := c.ms.GetCounterValue(metric)

		if !hasVal {
			http.NotFound(w, req)
			return
		}

		c.response.Send(w, http.StatusOK, val)
	}
}

func (c *Controller) GetGauge() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		metric := req.PathValue(metricReqPathName)

		val, hasVal := c.ms.GetGaugeValue(metric)

		if !hasVal {
			http.NotFound(w, req)
			return
		}

		c.response.Send(w, http.StatusOK, val)
	}
}

func (c *Controller) syncStateLogger(r *http.Request) {
	storeInterval := c.conf.GetStoreInterval()
	if storeInterval == 0 {
		err := c.ms.FlushStorage()
		if err != nil {
			c.l.ErrorCtx(r.Context(), "sync state error", zap.Error(err))
		}
	}
}
