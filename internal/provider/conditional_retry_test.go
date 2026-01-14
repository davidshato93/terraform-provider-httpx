package provider

import (
	"context"
	"encoding/json"
	"regexp"
	"testing"
)

func TestCheckJsonPathConditions(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		conditions map[string]string
		want       bool
	}{
		{
			name: "simple match",
			body: `{"status": "ready"}`,
			conditions: map[string]string{
				"status": "ready",
			},
			want: true,
		},
		{
			name: "nested path match",
			body: `{"data": {"status": "ready"}}`,
			conditions: map[string]string{
				"data.status": "ready",
			},
			want: true,
		},
		{
			name: "array index match",
			body: `{"items": [{"id": "123"}]}`,
			conditions: map[string]string{
				"items[0].id": "123",
			},
			want: true,
		},
		{
			name: "multiple conditions all match",
			body: `{"status": "ready", "count": "5"}`,
			conditions: map[string]string{
				"status": "ready",
				"count":  "5",
			},
			want: true,
		},
		{
			name: "condition mismatch",
			body: `{"status": "pending"}`,
			conditions: map[string]string{
				"status": "ready",
			},
			want: false,
		},
		{
			name: "invalid JSON",
			body: `{invalid json}`,
			conditions: map[string]string{
				"status": "ready",
			},
			want: false,
		},
		{
			name:       "no conditions",
			body:       `{"status": "ready"}`,
			conditions: map[string]string{},
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkJsonPathConditions(context.Background(), tt.body, tt.conditions)
			if got != tt.want {
				t.Errorf("checkJsonPathConditions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvaluateJsonPath(t *testing.T) {
	tests := []struct {
		name    string
		data    interface{}
		path    string
		want    interface{}
		wantErr bool
	}{
		{
			name:    "empty path returns root",
			data:    map[string]interface{}{"key": "value"},
			path:    "",
			want:    map[string]interface{}{"key": "value"},
			wantErr: false,
		},
		{
			name: "simple key",
			data: map[string]interface{}{
				"status": "ready",
			},
			path:    "status",
			want:    "ready",
			wantErr: false,
		},
		{
			name: "nested path",
			data: map[string]interface{}{
				"data": map[string]interface{}{
					"status": "ready",
				},
			},
			path:    "data.status",
			want:    "ready",
			wantErr: false,
		},
		{
			name: "array index",
			data: map[string]interface{}{
				"items": []interface{}{
					map[string]interface{}{"id": "123"},
				},
			},
			path:    "items[0].id",
			want:    "123",
			wantErr: false,
		},
		{
			name: "non-existent key",
			data: map[string]interface{}{
				"status": "ready",
			},
			path:    "missing",
			wantErr: true,
		},
		{
			name: "invalid path",
			data: map[string]interface{}{
				"status": "ready",
			},
			path:    "status[invalid]",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := evaluateJsonPath(tt.data, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("evaluateJsonPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// Convert to JSON for comparison
				gotJSON, _ := json.Marshal(got)
				wantJSON, _ := json.Marshal(tt.want)
				if string(gotJSON) != string(wantJSON) {
					t.Errorf("evaluateJsonPath() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestCheckHeaderEquals(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		conditions map[string]string
		want       bool
	}{
		{
			name: "match single header",
			headers: map[string]string{
				"Content-Type": "application/json",
			},
			conditions: map[string]string{
				"Content-Type": "application/json",
			},
			want: true,
		},
		{
			name: "case insensitive match",
			headers: map[string]string{
				"content-type": "application/json",
			},
			conditions: map[string]string{
				"Content-Type": "application/json",
			},
			want: true,
		},
		{
			name: "multiple headers all match",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Request-ID": "12345",
			},
			conditions: map[string]string{
				"Content-Type": "application/json",
				"X-Request-ID":  "12345",
			},
			want: true,
		},
		{
			name: "header mismatch",
			headers: map[string]string{
				"Content-Type": "text/plain",
			},
			conditions: map[string]string{
				"Content-Type": "application/json",
			},
			want: false,
		},
		{
			name: "missing header",
			headers: map[string]string{
				"Other-Header": "value",
			},
			conditions: map[string]string{
				"Content-Type": "application/json",
			},
			want: false,
		},
		{
			name:       "no conditions",
			headers:    map[string]string{"Content-Type": "application/json"},
			conditions: map[string]string{},
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkHeaderConditions(tt.headers, tt.conditions)
			if got != tt.want {
				t.Errorf("checkHeaderEquals() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckBodyRegex(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		pattern string
		want    bool
	}{
		{
			name:    "simple match",
			body:    "hello world",
			pattern: "hello",
			want:    true,
		},
		{
			name:    "regex match",
			body:    "status: ready",
			pattern: "status:\\s+ready",
			want:    true,
		},
		{
			name:    "no match",
			body:    "hello world",
			pattern: "goodbye",
			want:    false,
		},
		{
			name:    "empty pattern",
			body:    "hello world",
			pattern: "",
			want:    true, // Empty pattern matches everything
		},
		{
			name:    "empty body",
			body:    "",
			pattern: "hello",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, err := regexp.MatchString(tt.pattern, tt.body)
			if err != nil {
				t.Errorf("regexp.MatchString() error = %v", err)
				return
			}
			if matched != tt.want {
				t.Errorf("regexp.MatchString() = %v, want %v", matched, tt.want)
			}
		})
	}
}

