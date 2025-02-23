package controller

import (
	"net/http"
	"strconv"

	"github.com/OleG2e/collector/internal/storage"

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
			c.st.SetGaugeValue(form.ID, *form.Value)

			c.syncStateLogger(r)

			c.response.Send(r.Context(), w, http.StatusOK, form)
			return
		}

		if form.IsCounterType() {
			c.st.AddCounterValue(form.ID, *form.Delta)
			val, hasVal := c.st.GetCounterValue(form.ID)
			if !hasVal {
				http.NotFound(w, r)
				return
			}
			form.Delta = &val

			c.syncStateLogger(r)

			c.response.Send(r.Context(), w, http.StatusOK, form)
			return
		}

		c.response.BadRequestError(w, "unknown metric type")
	}
}

func (c *Controller) UpdateMetrics() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		forms, decodeErr := network.NewFormArrayByRequest(r)

		if decodeErr != nil {
			c.l.ErrorCtx(r.Context(), "decodeErr", zap.Error(decodeErr))
			c.response.BadRequestError(w, decodeErr.Error())
			return
		}

		if len(forms) == 0 {
			c.response.BadRequestError(w, "no metrics found")
			return
		}

		for _, form := range forms {
			if form.IsGaugeType() {
				c.st.SetGaugeValue(form.ID, *form.Value)
			}

			if form.IsCounterType() {
				c.st.AddCounterValue(form.ID, *form.Delta)
			}
		}

		c.syncStateLogger(r)

		c.response.Success(w)
	}
}

func (c *Controller) PingDB() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		storeType := c.st.GetStoreAlgo().GetStoreType()

		if storeType != storage.DBStoreType {
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
			value, _ := c.st.GetGaugeValue(form.ID)
			form.Value = &value
			c.response.Send(r.Context(), w, http.StatusOK, form)
			return
		}

		if form.IsCounterType() {
			value, _ := c.st.GetCounterValue(form.ID)
			form.Delta = &value
			c.response.Send(r.Context(), w, http.StatusOK, form)
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

		c.st.AddCounterValue(metric, value)

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

		c.st.SetGaugeValue(metric, value)

		c.syncStateLogger(req)

		c.response.Success(w)
	}
}

func (c *Controller) GetCounter() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metric := r.PathValue(metricReqPathName)

		val, hasVal := c.st.GetCounterValue(metric)

		if !hasVal {
			http.NotFound(w, r)
			return
		}

		c.response.Send(r.Context(), w, http.StatusOK, val)
	}
}

func (c *Controller) GetGauge() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metric := r.PathValue(metricReqPathName)

		val, hasVal := c.st.GetGaugeValue(metric)

		if !hasVal {
			http.NotFound(w, r)
			return
		}

		c.response.Send(r.Context(), w, http.StatusOK, val)
	}
}

func (c *Controller) syncStateLogger(r *http.Request) {
	storeInterval := c.conf.GetStoreInterval()
	if storeInterval == 0 {
		err := c.st.FlushStorage(r.Context())
		if err != nil {
			c.l.ErrorCtx(r.Context(), "sync state error", zap.Error(err))
		}
	}
}
