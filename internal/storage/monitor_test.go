package storage

import (
	"runtime"
	"testing"
)

//func TestMonitorStorage(t *testing.T) {
//	type want struct {
//		contentType string
//		statusCode  int
//		user        User
//	}
//	tests := []struct {
//		name    string
//		request string
//		users   map[string]User
//		want    want
//	}{
//		{
//			name: "simple test #1",
//			users: map[string]User{
//				"id1": {
//					ID:        "id1",
//					FirstName: "Misha",
//					LastName:  "Popov",
//				},
//			},
//			want: want{
//				contentType: "application/json",
//				statusCode:  200,
//				user: User{ID: "id1",
//					FirstName: "Misha",
//					LastName:  "Popov",
//				},
//			},
//			request: "/users?user_id=id1",
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			request := httptest.NewRequest(http.MethodPost, tt.request, nil)
//			w := httptest.NewRecorder()
//			h := http.HandlerFunc(UserViewHandler(tt.users))
//			h(w, request)
//
//			result := w.Result()
//
//			assert.Equal(t, tt.want.statusCode, result.StatusCode)
//			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
//
//			userResult, err := ioutil.ReadAll(result.Body)
//			require.NoError(t, err)
//			err = result.Body.Close()
//			require.NoError(t, err)
//
//			var user User
//			err = json.Unmarshal(userResult, &user)
//			require.NoError(t, err)
//
//			assert.Equal(t, tt.want.user, user)
//		})
//	}
//}

func TestRunMonitor(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RunMonitor()
		})
	}
}

func Test_initMonitor(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initMonitor()
		})
	}
}

func Test_monitorStorage_initSendTicker(t *testing.T) {
	type fields struct {
		Stats        map[string]any
		RuntimeStats *runtime.MemStats
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &monitorStorage{
				Stats:        tt.fields.Stats,
				RuntimeStats: tt.fields.RuntimeStats,
			}
			s.initSendTicker()
		})
	}
}

func Test_monitorStorage_refreshStats(t *testing.T) {
	type fields struct {
		Stats        map[string]any
		RuntimeStats *runtime.MemStats
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &monitorStorage{
				Stats:        tt.fields.Stats,
				RuntimeStats: tt.fields.RuntimeStats,
			}
			s.refreshStats()
		})
	}
}

func Test_monitorStorage_seedGauge(t *testing.T) {
	type fields struct {
		Stats        map[string]any
		RuntimeStats *runtime.MemStats
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &monitorStorage{
				Stats:        tt.fields.Stats,
				RuntimeStats: tt.fields.RuntimeStats,
			}
			s.seedGauge()
		})
	}
}

func Test_monitorStorage_sendCounterData(t *testing.T) {
	type fields struct {
		Stats        map[string]any
		RuntimeStats *runtime.MemStats
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &monitorStorage{
				Stats:        tt.fields.Stats,
				RuntimeStats: tt.fields.RuntimeStats,
			}
			s.sendCounterData()
		})
	}
}

func Test_monitorStorage_sendGaugeData(t *testing.T) {
	type fields struct {
		Stats        map[string]any
		RuntimeStats *runtime.MemStats
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &monitorStorage{
				Stats:        tt.fields.Stats,
				RuntimeStats: tt.fields.RuntimeStats,
			}
			s.sendGaugeData()
		})
	}
}
