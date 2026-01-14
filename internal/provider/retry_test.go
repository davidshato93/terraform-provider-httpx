package provider

import (
	"errors"
	"testing"
	"time"
)

func TestRetryConfig_ShouldRetry(t *testing.T) {
	tests := []struct {
		name       string
		config     RetryConfig
		err        error
		statusCode int64
		want       bool
	}{
		{
			name: "retry on error",
			config: RetryConfig{
				RetryOnStatusCodes: []int64{500, 502},
			},
			err:        errors.New("connection failed"),
			statusCode: 200,
			want:       true,
		},
		{
			name: "retry on configured status code",
			config: RetryConfig{
				RetryOnStatusCodes: []int64{500, 502, 503},
			},
			err:        nil,
			statusCode: 500,
			want:       true,
		},
		{
			name: "don't retry on success",
			config: RetryConfig{
				RetryOnStatusCodes: []int64{500, 502},
			},
			err:        nil,
			statusCode: 200,
			want:       false,
		},
		{
			name: "don't retry on unconfigured status code",
			config: RetryConfig{
				RetryOnStatusCodes: []int64{500, 502},
			},
			err:        nil,
			statusCode: 404,
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.ShouldRetry(tt.err, tt.statusCode)
			if got != tt.want {
				t.Errorf("ShouldRetry() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRetryConfig_CalculateDelay(t *testing.T) {
	tests := []struct {
		name       string
		config     RetryConfig
		attempt    int64
		retryAfter string
		wantMin    time.Duration
		wantMax    time.Duration
	}{
		{
			name: "fixed backoff",
			config: RetryConfig{
				MinDelayMs: 1000,
				MaxDelayMs: 5000,
				Backoff:    "fixed",
			},
			attempt: 3,
			wantMin: 1000 * time.Millisecond,
			wantMax: 1000 * time.Millisecond,
		},
		{
			name: "linear backoff",
			config: RetryConfig{
				MinDelayMs: 1000,
				MaxDelayMs: 10000,
				Backoff:    "linear",
			},
			attempt: 2,
			wantMin: 2000 * time.Millisecond,
			wantMax: 2000 * time.Millisecond,
		},
		{
			name: "exponential backoff",
			config: RetryConfig{
				MinDelayMs: 1000,
				MaxDelayMs: 10000,
				Backoff:    "exponential",
			},
			attempt: 3,
			wantMin: 4000 * time.Millisecond, // 1000 * 2^(3-1) = 4000
			wantMax: 4000 * time.Millisecond,
		},
		{
			name: "respect max delay",
			config: RetryConfig{
				MinDelayMs: 1000,
				MaxDelayMs: 3000,
				Backoff:    "exponential",
			},
			attempt: 5, // Would be 16000ms without cap
			wantMin: 3000 * time.Millisecond,
			wantMax: 3000 * time.Millisecond,
		},
		{
			name: "respect retry after header",
			config: RetryConfig{
				MinDelayMs:        1000,
				MaxDelayMs:       5000,
				Backoff:          "fixed",
				RespectRetryAfter: true,
			},
			attempt:    1,
			retryAfter: "5",
			wantMin:    5 * time.Second,
			wantMax:    5 * time.Second,
		},
		{
			name: "jitter adds randomness",
			config: RetryConfig{
				MinDelayMs: 1000,
				MaxDelayMs: 5000,
				Backoff:    "fixed",
				Jitter:     true,
			},
			attempt: 1,
			wantMin: 1000 * time.Millisecond,
			wantMax: 1250 * time.Millisecond, // 1000 + 25% = 1250
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delay := tt.config.CalculateDelay(tt.attempt, tt.retryAfter)
			if delay < tt.wantMin {
				t.Errorf("CalculateDelay() = %v, want >= %v", delay, tt.wantMin)
			}
			if delay > tt.wantMax {
				t.Errorf("CalculateDelay() = %v, want <= %v", delay, tt.wantMax)
			}
		})
	}
}

func TestParseRetryAfter(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Duration
		wantErr bool
	}{
		{
			name:    "seconds as integer",
			input:   "5",
			want:    5 * time.Second,
			wantErr: false,
		},
		{
			name:    "seconds with whitespace",
			input:   " 10 ",
			want:    10 * time.Second,
			wantErr: false,
		},
		{
			name:    "HTTP date format (future date)",
			input:   "Wed, 21 Oct 2030 07:28:00 GMT", // Far future date
			wantErr: false, // Will parse to a duration
		},
		{
			name:    "invalid format",
			input:   "invalid",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRetryAfter(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRetryAfter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.want > 0 && got != tt.want {
				t.Errorf("parseRetryAfter() = %v, want %v", got, tt.want)
			}
		})
	}
}

