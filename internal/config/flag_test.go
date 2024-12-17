package config

import (
	"reflect"
	"testing"
)

func TestConfig_GetPollInterval(t *testing.T) {
	type fields struct {
		ServerHostPort string
		ReportInterval int
		PollInterval   int
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Config{
				ServerHostPort: tt.fields.ServerHostPort,
				ReportInterval: tt.fields.ReportInterval,
				PollInterval:   tt.fields.PollInterval,
			}
			if got := c.GetPollInterval(); got != tt.want {
				t.Errorf("GetPollInterval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_GetReportInterval(t *testing.T) {
	type fields struct {
		ServerHostPort string
		ReportInterval int
		PollInterval   int
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Config{
				ServerHostPort: tt.fields.ServerHostPort,
				ReportInterval: tt.fields.ReportInterval,
				PollInterval:   tt.fields.PollInterval,
			}
			if got := c.GetReportInterval(); got != tt.want {
				t.Errorf("GetReportInterval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_GetServerHostPort(t *testing.T) {
	type fields struct {
		ServerHostPort string
		ReportInterval int
		PollInterval   int
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Config{
				ServerHostPort: tt.fields.ServerHostPort,
				ReportInterval: tt.fields.ReportInterval,
				PollInterval:   tt.fields.PollInterval,
			}
			if got := c.GetServerHostPort(); got != tt.want {
				t.Errorf("GetServerHostPort() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetConfig(t *testing.T) {
	tests := []struct {
		name string
		want *Config
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetConfig(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_initAppConfig(t *testing.T) {
	tests := []struct {
		name    string
		want    *Config
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := initAppConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("initAppConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("initAppConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}
