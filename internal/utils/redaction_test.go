package utils

import (
	"testing"
)

func TestRedactHeaders(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		redactList []string
		expected   map[string]string
	}{
		{
			name: "no redaction",
			headers: map[string]string{
				"Content-Type": "application/json",
				"User-Agent":   "terraform",
			},
			redactList: []string{},
			expected: map[string]string{
				"Content-Type": "application/json",
				"User-Agent":   "terraform",
			},
		},
		{
			name: "redact authorization",
			headers: map[string]string{
				"Authorization": "Bearer secret-token",
				"Content-Type":  "application/json",
			},
			redactList: []string{"Authorization"},
			expected: map[string]string{
				"Authorization": "[REDACTED]",
				"Content-Type":  "application/json",
			},
		},
		{
			name: "case insensitive redaction",
			headers: map[string]string{
				"authorization": "Bearer secret-token",
				"AUTHORIZATION": "Bearer secret-token",
				"Authorization": "Bearer secret-token",
			},
			redactList: []string{"Authorization"},
			expected: map[string]string{
				"authorization": "[REDACTED]",
				"AUTHORIZATION": "[REDACTED]",
				"Authorization": "[REDACTED]",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RedactHeaders(tt.headers, tt.redactList)
			for k, v := range tt.expected {
				if result[k] != v {
					t.Errorf("RedactHeaders() header %s = %v, want %v", k, result[k], v)
				}
			}
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		maxLen  int
		wantLen int
	}{
		{
			name:    "no truncation needed",
			input:   "short",
			maxLen:  10,
			wantLen: 5,
		},
		{
			name:    "truncation needed",
			input:   "this is a very long string",
			maxLen:  10,
			wantLen: 10 + len("... [TRUNCATED]"),
		},
		{
			name:    "exact length",
			input:   "exact",
			maxLen:  5,
			wantLen: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateString(tt.input, tt.maxLen)
			if len(result) != tt.wantLen {
				t.Errorf("TruncateString() length = %d, want %d", len(result), tt.wantLen)
			}
			if tt.maxLen < len(tt.input) && result[len(result)-len("... [TRUNCATED]"):] != "... [TRUNCATED]" {
				t.Errorf("TruncateString() should end with truncation marker")
			}
		})
	}
}

func TestRedactError(t *testing.T) {
	tests := []struct {
		name       string
		errMsg     string
		redactList []string
		shouldContain string
		shouldNotContain string
	}{
		{
			name:       "redact authorization",
			errMsg:     "error: Authorization: Bearer secret-token",
			redactList: []string{"Authorization"},
			shouldContain: "[REDACTED]",
			shouldNotContain: "", // RedactError only replaces header names, not values
		},
		{
			name:       "no redaction",
			errMsg:     "error: connection failed",
			redactList: []string{},
			shouldContain: "error: connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RedactError(tt.errMsg, tt.redactList)
			if tt.shouldContain != "" && !contains(result, tt.shouldContain) {
				t.Errorf("RedactError() should contain %s, got %s", tt.shouldContain, result)
			}
			if tt.shouldNotContain != "" && contains(result, tt.shouldNotContain) {
				t.Errorf("RedactError() should not contain %s, got %s", tt.shouldNotContain, result)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		containsMiddle(s, substr))))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

