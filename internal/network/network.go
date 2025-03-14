package network

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type MetricType string

const (
	MetricTypeGauge   = MetricType("gauge")
	MetricTypeCounter = MetricType("counter")
)

const HashHeader = "HashSHA256"

type MetricForm struct {
	ID    string     `json:"id"`              // имя метрики
	MType MetricType `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64     `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64   `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func (f *MetricForm) IsGaugeType() bool {
	return f.MType == MetricTypeGauge
}

func (f *MetricForm) IsCounterType() bool {
	return f.MType == MetricTypeCounter
}

func NewFormByRequest(r *http.Request) (*MetricForm, error) {
	var bodyBuffer bytes.Buffer
	r.Body = io.NopCloser(io.TeeReader(r.Body, &bodyBuffer))

	form := MetricForm{}

	decodeErr := json.NewDecoder(r.Body).Decode(&form)

	r.Body = io.NopCloser(&bodyBuffer)

	if errors.Is(decodeErr, io.EOF) {
		decodeErr = nil
	}

	return &form, decodeErr
}

func NewFormArrayByRequest(r *http.Request) ([]MetricForm, error) {
	var bodyBuffer bytes.Buffer
	r.Body = io.NopCloser(io.TeeReader(r.Body, &bodyBuffer))

	var forms []MetricForm

	decodeErr := json.NewDecoder(r.Body).Decode(&forms)

	r.Body = io.NopCloser(&bodyBuffer)

	if errors.Is(decodeErr, io.EOF) {
		decodeErr = nil
	}

	return forms, decodeErr
}
