package response

import (
	"net/http"
	"testing"
)

func TestBadRequestError(t *testing.T) {
	type args struct {
		writer http.ResponseWriter
		e      string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			BadRequestError(tt.args.writer, tt.args.e)
		})
	}
}

func TestSend(t *testing.T) {
	type args struct {
		writer     http.ResponseWriter
		statusCode int
		data       string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Send(tt.args.writer, tt.args.statusCode, tt.args.data)
		})
	}
}

func TestSuccess(t *testing.T) {
	type args struct {
		writer http.ResponseWriter
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Success(tt.args.writer)
		})
	}
}

func Test_setDefaultHeaders(t *testing.T) {
	type args struct {
		writer http.ResponseWriter
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setDefaultHeaders(tt.args.writer)
		})
	}
}

func Test_setStatusCode(t *testing.T) {
	type args struct {
		writer     http.ResponseWriter
		statusCode int
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setStatusCode(tt.args.writer, tt.args.statusCode)
		})
	}
}
