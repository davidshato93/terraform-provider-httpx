package client

import (
	"bytes"
	"io"
	"testing"
	"time"

	"github.com/davidshato/terraform-provider-httpx/internal/config"
)

func TestNewHTTPClient(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.ProviderConfig
		wantErr bool
	}{
		{
			name: "basic client",
			config: &config.ProviderConfig{
				TimeoutMs:          30000,
				InsecureSkipVerify: false,
			},
			wantErr: false,
		},
		{
			name: "insecure skip verify",
			config: &config.ProviderConfig{
				TimeoutMs:          10000,
				InsecureSkipVerify: true,
			},
			wantErr: false,
		},
		{
			name: "with proxy",
			config: &config.ProviderConfig{
				TimeoutMs:          30000,
				InsecureSkipVerify: false,
				ProxyUrl:           stringPtr("http://proxy.example.com:8080"),
			},
			wantErr: false,
		},
		{
			name: "invalid proxy URL",
			config: &config.ProviderConfig{
				TimeoutMs:          30000,
				InsecureSkipVerify: false,
				ProxyUrl:           stringPtr("://invalid-url"),
			},
			wantErr: true,
		},
		{
			name: "with CA cert",
			config: &config.ProviderConfig{
				TimeoutMs:          30000,
				InsecureSkipVerify: false,
				CaCertPem:          stringPtr("-----BEGIN CERTIFICATE-----\nMOCK\n-----END CERTIFICATE-----"),
			},
			wantErr: true, // Invalid cert will fail to parse
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewHTTPClient(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewHTTPClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Errorf("NewHTTPClient() returned nil client")
			}
			if !tt.wantErr && client != nil {
				if client.GetTimeout() != time.Duration(tt.config.TimeoutMs)*time.Millisecond {
					t.Errorf("GetTimeout() = %v, want %v", client.GetTimeout(), time.Duration(tt.config.TimeoutMs)*time.Millisecond)
				}
			}
		})
	}
}

func TestLimitReader(t *testing.T) {
	tests := []struct {
		name     string
		data     string
		maxBytes int64
		wantRead int
	}{
		{
			name:     "read all data",
			data:     "hello world",
			maxBytes: 20,
			wantRead: 11,
		},
		{
			name:     "limit reading",
			data:     "hello world",
			maxBytes: 5,
			wantRead: 5,
		},
		{
			name:     "zero limit",
			data:     "hello world",
			maxBytes: 0,
			wantRead: 0,
		},
		{
			name:     "exact limit",
			data:     "hello",
			maxBytes: 5,
			wantRead: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader([]byte(tt.data))
			limited := LimitReader(reader, tt.maxBytes)

			buf := make([]byte, len(tt.data)+10) // Buffer larger than data
			n, err := limited.Read(buf)

			if err != nil && err != io.EOF {
				t.Errorf("LimitReader.Read() error = %v", err)
			}
			if n != tt.wantRead {
				t.Errorf("LimitReader.Read() read %d bytes, want %d", n, tt.wantRead)
			}
		})
	}
}

func TestGetTimeout(t *testing.T) {
	cfg := &config.ProviderConfig{
		TimeoutMs: 5000,
	}
	client, err := NewHTTPClient(cfg)
	if err != nil {
		t.Fatalf("NewHTTPClient() error = %v", err)
	}

	timeout := client.GetTimeout()
	expected := 5 * time.Second
	if timeout != expected {
		t.Errorf("GetTimeout() = %v, want %v", timeout, expected)
	}
}

func stringPtr(s string) *string {
	return &s
}
