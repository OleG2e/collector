package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"collector/internal/adapters/store"
	"collector/internal/config"
	"collector/internal/core/domain"
	"collector/pkg/network"
)

type fakeStore struct {
	metrics   *domain.Metrics
	storeType store.Type
}

func (f *fakeStore) GetMetrics() *domain.Metrics     { return f.metrics }
func (f *fakeStore) SetMetrics(m *domain.Metrics)    { f.metrics = m }
func (f *fakeStore) Save(_ context.Context) error    { return nil }
func (f *fakeStore) Restore(_ context.Context) error { return nil }
func (f *fakeStore) Close() error                    { return nil }
func (f *fakeStore) GetStoreType() store.Type        { return f.storeType }

func newTestRouter(st store.Store) *chiRouter {
	logger := newTestLogger()
	conf := &config.ServerConfig{}
	resp := network.NewResponse(logger, conf)
	return &chiRouter{mux: NewRouter(st, logger, conf, resp)}
}

type chiRouter struct{ mux http.Handler }

func (r *chiRouter) do(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	r.mux.ServeHTTP(rr, req)
	return rr
}

func TestRouter_RootOK(t *testing.T) {
	st := &fakeStore{metrics: domain.NewMetrics(), storeType: store.MemoryStoreType}
	rt := newTestRouter(st)

	rr := rt.do(httptest.NewRequest(http.MethodGet, "/", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("GET / expected 200, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "text/html" {
		t.Fatalf("unexpected content-type: %q", ct)
	}
}

func TestRouter_PingDB_DependsOnStoreType(t *testing.T) {
	st1 := &fakeStore{metrics: domain.NewMetrics(), storeType: store.FileStoreType}
	rt1 := newTestRouter(st1)
	rr1 := rt1.do(httptest.NewRequest(http.MethodGet, "/ping", nil))
	if rr1.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for non-db store, got %d", rr1.Code)
	}

	st2 := &fakeStore{metrics: domain.NewMetrics(), storeType: store.DBStoreType}
	rt2 := newTestRouter(st2)
	rr2 := rt2.do(httptest.NewRequest(http.MethodGet, "/ping", nil))
	if rr2.Code != http.StatusOK {
		t.Fatalf("expected 200 for db store, got %d", rr2.Code)
	}
}

func TestRouter_UpdateGauge_PathAndGet(t *testing.T) {
	st := &fakeStore{metrics: domain.NewMetrics(), storeType: store.MemoryStoreType}
	rt := newTestRouter(st)

	rr := rt.do(httptest.NewRequest(http.MethodPost, "/update/gauge/Temp/12.5", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("update gauge expected 200, got %d", rr.Code)
	}
	rr2 := rt.do(httptest.NewRequest(http.MethodGet, "/value/gauge/Temp", nil))
	if rr2.Code != http.StatusOK {
		t.Fatalf("get gauge expected 200, got %d", rr2.Code)
	}

	var got float64
	if err := json.Unmarshal(rr2.Body.Bytes(), &got); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if got != 12.5 {
		t.Fatalf("unexpected gauge value: %v", got)
	}
}

func TestRouter_UpdateCounter_PathAndGet(t *testing.T) {
	st := &fakeStore{metrics: domain.NewMetrics(), storeType: store.MemoryStoreType}
	rt := newTestRouter(st)

	rr := rt.do(httptest.NewRequest(http.MethodPost, "/update/counter/Requests/3", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("update counter expected 200, got %d", rr.Code)
	}
	rr2 := rt.do(httptest.NewRequest(http.MethodGet, "/value/counter/Requests", nil))
	if rr2.Code != http.StatusOK {
		t.Fatalf("get counter expected 200, got %d", rr2.Code)
	}

	var got int64
	if err := json.Unmarshal(rr2.Body.Bytes(), &got); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if got != 3 {
		t.Fatalf("unexpected counter value: %v", got)
	}
}

func TestRouter_GetGauge_NotFound(t *testing.T) {
	st := &fakeStore{metrics: domain.NewMetrics(), storeType: store.MemoryStoreType}
	rt := newTestRouter(st)

	rr := rt.do(httptest.NewRequest(http.MethodGet, "/value/gauge/Unknown", nil))
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown gauge, got %d", rr.Code)
	}
}

func TestRouter_GetCounter_NotFound(t *testing.T) {
	st := &fakeStore{metrics: domain.NewMetrics(), storeType: store.MemoryStoreType}
	rt := newTestRouter(st)

	rr := rt.do(httptest.NewRequest(http.MethodGet, "/value/counter/Unknown", nil))
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown counter, got %d", rr.Code)
	}
}

func TestRouter_UpdateMetric_Body_Gauge(t *testing.T) {
	st := &fakeStore{metrics: domain.NewMetrics(), storeType: store.MemoryStoreType}
	rt := newTestRouter(st)

	body := `{"id":"HeapAlloc","type":"gauge","value":123.45}`
	rr := rt.do(httptest.NewRequest(http.MethodPost, "/update/", bytes.NewBufferString(body)))
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var form domain.MetricForm
	if err := json.Unmarshal(rr.Body.Bytes(), &form); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if !form.IsGaugeType() || form.Value == nil || *form.Value != 123.45 {
		t.Fatalf("unexpected form: %+v", form)
	}
	if v, ok := st.metrics.GetGaugeValue("HeapAlloc"); !ok || v != 123.45 {
		t.Fatalf("value not persisted, got %v ok=%v", v, ok)
	}
}

func TestRouter_UpdateMetric_Body_Counter_ReturnsAggregatedDelta(t *testing.T) {
	st := &fakeStore{metrics: domain.NewMetrics(), storeType: store.MemoryStoreType}
	rt := newTestRouter(st)

	body1 := `{"id":"PollCount","type":"counter","delta":2}`
	rr1 := rt.do(httptest.NewRequest(http.MethodPost, "/update/", bytes.NewBufferString(body1)))
	if rr1.Code != http.StatusOK {
		t.Fatalf("expected 200 on first update, got %d", rr1.Code)
	}
	body2 := `{"id":"PollCount","type":"counter","delta":5}`
	rr2 := rt.do(httptest.NewRequest(http.MethodPost, "/update/", bytes.NewBufferString(body2)))
	if rr2.Code != http.StatusOK {
		t.Fatalf("expected 200 on second update, got %d", rr2.Code)
	}

	var form domain.MetricForm
	if err := json.Unmarshal(rr2.Body.Bytes(), &form); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if !form.IsCounterType() || form.Delta == nil || *form.Delta != 7 {
		t.Fatalf("unexpected form after aggregation: %+v", form)
	}
}

func TestRouter_GetMetric_Body(t *testing.T) {
	st := &fakeStore{metrics: domain.NewMetrics(), storeType: store.MemoryStoreType}
	st.metrics.SetGaugeValue("Alloc", 10.5)
	st.metrics.AddCounterValue("Requests", 4)
	rt := newTestRouter(st)

	reqGauge := domain.MetricForm{ID: "Alloc", MType: domain.MetricTypeGauge}
	gBody, _ := json.Marshal(reqGauge)
	rr1 := rt.do(httptest.NewRequest(http.MethodPost, "/value/", bytes.NewReader(gBody)))
	if rr1.Code != http.StatusOK {
		t.Fatalf("expected 200 for gauge, got %d", rr1.Code)
	}
	var gResp domain.MetricForm
	_ = json.Unmarshal(rr1.Body.Bytes(), &gResp)
	if gResp.Value == nil || *gResp.Value != 10.5 {
		t.Fatalf("unexpected gauge in response: %+v", gResp)
	}

	reqCnt := domain.MetricForm{ID: "Requests", MType: domain.MetricTypeCounter}
	cBody, _ := json.Marshal(reqCnt)
	rr2 := rt.do(httptest.NewRequest(http.MethodPost, "/value/", bytes.NewReader(cBody)))
	if rr2.Code != http.StatusOK {
		t.Fatalf("expected 200 for counter, got %d", rr2.Code)
	}
	var cResp domain.MetricForm
	_ = json.Unmarshal(rr2.Body.Bytes(), &cResp)
	if cResp.Delta == nil || *cResp.Delta != 4 {
		t.Fatalf("unexpected counter in response: %+v", cResp)
	}
}

func TestRouter_UpdateMetrics_Batch(t *testing.T) {
	st := &fakeStore{metrics: domain.NewMetrics(), storeType: store.MemoryStoreType}
	rt := newTestRouter(st)

	forms := []domain.MetricForm{
		{ID: "G1", MType: domain.MetricTypeGauge, Value: float64Ptr(1.25)},
		{ID: "C1", MType: domain.MetricTypeCounter, Delta: int64Ptr(3)},
	}
	body, _ := json.Marshal(forms)

	rr := rt.do(httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewReader(body)))
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if v, ok := st.metrics.GetGaugeValue("G1"); !ok || v != 1.25 {
		t.Fatalf("unexpected stored gauge: %v ok=%v", v, ok)
	}
	if v, ok := st.metrics.GetCounterValue("C1"); !ok || v != 3 {
		t.Fatalf("unexpected stored counter: %v ok=%v", v, ok)
	}
}

func TestRouter_UpdateMetrics_Batch_Empty(t *testing.T) {
	st := &fakeStore{metrics: domain.NewMetrics(), storeType: store.MemoryStoreType}
	rt := newTestRouter(st)

	rr := rt.do(httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewBufferString("[]")))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty batch, got %d", rr.Code)
	}
}

func float64Ptr(v float64) *float64 { return &v }
func int64Ptr(v int64) *int64       { return &v }
