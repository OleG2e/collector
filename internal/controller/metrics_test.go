package controller

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"github.com/OleG2e/collector/internal/config"
	"github.com/OleG2e/collector/internal/response"
	"github.com/OleG2e/collector/internal/storage"
	"github.com/OleG2e/collector/pkg/logging"
	"github.com/go-chi/chi/v5"
)

func TestController_GetCounter(t *testing.T) {
	type fields struct {
		l        *logging.ZapLogger
		ctx      context.Context
		router   chi.Router
		response *response.Response
		ms       *storage.MemStorage
		conf     *config.ServerConfig
	}
	tests := []struct {
		name   string
		fields fields
		want   http.HandlerFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Controller{
				l:        tt.fields.l,
				router:   tt.fields.router,
				response: tt.fields.response,
				ms:       tt.fields.ms,
				conf:     tt.fields.conf,
			}
			if got := c.GetCounter(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetCounter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestController_GetGauge(t *testing.T) {
	type fields struct {
		l        *logging.ZapLogger
		ctx      context.Context
		router   chi.Router
		response *response.Response
		ms       *storage.MemStorage
		conf     *config.ServerConfig
	}
	tests := []struct {
		name   string
		fields fields
		want   http.HandlerFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Controller{
				l:        tt.fields.l,
				router:   tt.fields.router,
				response: tt.fields.response,
				ms:       tt.fields.ms,
				conf:     tt.fields.conf,
			}
			if got := c.GetGauge(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetGauge() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestController_GetMetric(t *testing.T) {
	type fields struct {
		l        *logging.ZapLogger
		ctx      context.Context
		router   chi.Router
		response *response.Response
		ms       *storage.MemStorage
		conf     *config.ServerConfig
	}
	tests := []struct {
		name   string
		fields fields
		want   http.HandlerFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Controller{
				l:        tt.fields.l,
				router:   tt.fields.router,
				response: tt.fields.response,
				ms:       tt.fields.ms,
				conf:     tt.fields.conf,
			}
			if got := c.GetMetric(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMetric() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestController_UpdateCounter(t *testing.T) {
	type fields struct {
		l        *logging.ZapLogger
		ctx      context.Context
		router   chi.Router
		response *response.Response
		ms       *storage.MemStorage
		conf     *config.ServerConfig
	}
	tests := []struct {
		name   string
		fields fields
		want   http.HandlerFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Controller{
				l:        tt.fields.l,
				router:   tt.fields.router,
				response: tt.fields.response,
				ms:       tt.fields.ms,
				conf:     tt.fields.conf,
			}
			if got := c.UpdateCounter(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UpdateCounter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestController_UpdateGauge(t *testing.T) {
	type fields struct {
		l        *logging.ZapLogger
		ctx      context.Context
		router   chi.Router
		response *response.Response
		ms       *storage.MemStorage
		conf     *config.ServerConfig
	}
	tests := []struct {
		name   string
		fields fields
		want   http.HandlerFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Controller{
				l:        tt.fields.l,
				router:   tt.fields.router,
				response: tt.fields.response,
				ms:       tt.fields.ms,
				conf:     tt.fields.conf,
			}
			if got := c.UpdateGauge(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UpdateGauge() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestController_UpdateMetric(t *testing.T) {
	type fields struct {
		l        *logging.ZapLogger
		ctx      context.Context
		router   chi.Router
		response *response.Response
		ms       *storage.MemStorage
		conf     *config.ServerConfig
	}
	tests := []struct {
		name   string
		fields fields
		want   http.HandlerFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Controller{
				l:        tt.fields.l,
				router:   tt.fields.router,
				response: tt.fields.response,
				ms:       tt.fields.ms,
				conf:     tt.fields.conf,
			}
			if got := c.UpdateMetric(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UpdateMetric() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestController_syncStateLogger(t *testing.T) {
	type fields struct {
		l        *logging.ZapLogger
		ctx      context.Context
		router   chi.Router
		response *response.Response
		ms       *storage.MemStorage
		conf     *config.ServerConfig
	}
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Controller{
				l:        tt.fields.l,
				router:   tt.fields.router,
				response: tt.fields.response,
				ms:       tt.fields.ms,
				conf:     tt.fields.conf,
			}
			c.syncStateLogger(tt.args.r)
		})
	}
}
