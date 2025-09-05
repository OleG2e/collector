package domain

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestNewFormByRequest_JSONGauge(t *testing.T) {
	body := `{"id":"HeapAlloc","type":"gauge","value":123.45}`
	req, _ := http.NewRequest(
		http.MethodPost,
		"/update/metric",
		io.NopCloser(strings.NewReader(body)),
	)

	form, err := NewFormByRequest(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if form.ID != "HeapAlloc" {
		t.Fatalf("unexpected id: %s", form.ID)
	}
	if !form.IsGaugeType() {
		t.Fatalf("expected gauge type")
	}
	if form.Value == nil || *form.Value != 123.45 {
		t.Fatalf("unexpected value: %v", form.Value)
	}

	b2, _ := io.ReadAll(req.Body)
	if !bytes.Contains(b2, []byte(`"HeapAlloc"`)) {
		t.Fatalf("request body not preserved")
	}
}

func TestNewFormByRequest_JSONCounter(t *testing.T) {
	body := `{"id":"PollCount","type":"counter","delta":10}`
	req, _ := http.NewRequest(
		http.MethodPost,
		"/update/metric",
		io.NopCloser(strings.NewReader(body)),
	)

	form, err := NewFormByRequest(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if form.ID != "PollCount" {
		t.Fatalf("unexpected id: %s", form.ID)
	}
	if !form.IsCounterType() {
		t.Fatalf("expected counter type")
	}
	if form.Delta == nil || *form.Delta != 10 {
		t.Fatalf("unexpected delta: %v", form.Delta)
	}

	b2, _ := io.ReadAll(req.Body)
	if !bytes.Contains(b2, []byte(`"PollCount"`)) {
		t.Fatalf("request body not preserved")
	}
}

func TestNewFormByRequest_EmptyBody_NoError(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "/update/metric", http.NoBody)

	form, err := NewFormByRequest(req)
	if err != nil {
		t.Fatalf("unexpected error on empty body: %v", err)
	}
	if form == nil {
		t.Fatalf("expected non-nil form")
	}
}

func TestNewFormByRequest_InvalidJSON_Error(t *testing.T) {
	req, _ := http.NewRequest(
		http.MethodPost,
		"/update/metric",
		io.NopCloser(strings.NewReader("{not-json")),
	)

	form, err := NewFormByRequest(req)
	if err == nil {
		t.Fatalf("expected error for invalid json, got nil, form: %+v", form)
	}
}

func TestNewFormArrayByRequest(t *testing.T) {
	body := `[{"id":"A","type":"gauge","value":1.5},{"id":"B","type":"counter","delta":2}]`
	req, _ := http.NewRequest(http.MethodPost, "/updates", io.NopCloser(strings.NewReader(body)))

	forms, err := NewFormArrayByRequest(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(forms) != 2 {
		t.Fatalf("unexpected forms len: %d", len(forms))
	}
	if forms[0].ID != "A" || !forms[0].IsGaugeType() || forms[0].Value == nil ||
		*forms[0].Value != 1.5 {
		t.Fatalf("unexpected first form: %+v", forms[0])
	}
	if forms[1].ID != "B" || !forms[1].IsCounterType() || forms[1].Delta == nil ||
		*forms[1].Delta != 2 {
		t.Fatalf("unexpected second form: %+v", forms[1])
	}

	b2, _ := io.ReadAll(req.Body)
	if !bytes.Contains(b2, []byte(`"A"`)) || !bytes.Contains(b2, []byte(`"B"`)) {
		t.Fatalf("request body not preserved")
	}
}

func TestMetricForm_TypeHelpers(t *testing.T) {
	g := MetricForm{MType: MetricTypeGauge}
	if !g.IsGaugeType() || g.IsCounterType() {
		t.Fatalf("gauge helpers invalid")
	}
	c := MetricForm{MType: MetricTypeCounter}
	if !c.IsCounterType() || c.IsGaugeType() {
		t.Fatalf("counter helpers invalid")
	}
}
