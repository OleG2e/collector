package controller

import (
	"net/http"
	"reflect"
	"testing"
)

func TestBadRequestHandler(t *testing.T) {
	tests := []struct {
		name string
		want http.HandlerFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BadRequestHandler(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BadRequestHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdateCounter(t *testing.T) {
	tests := []struct {
		name string
		want http.HandlerFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UpdateCounter(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UpdateCounter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdateGauge(t *testing.T) {
	tests := []struct {
		name string
		want http.HandlerFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UpdateGauge(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UpdateGauge() = %v, want %v", got, tt.want)
			}
		})
	}
}
