package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestExtractValues(t *testing.T) {
	tests := []struct {
		name         string
		result       *ResponseResult
		extractBlocks []ExtractBlockModel
		want         map[string]string
		wantErr      bool
	}{
		{
			name: "extract JSON path",
			result: &ResponseResult{
				Body: `{"id": "123", "name": "test"}`,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
			},
			extractBlocks: []ExtractBlockModel{
				{
					Name:    types.StringValue("id"),
					JsonPath: types.StringValue("id"),
				},
			},
			want: map[string]string{
				"id": "123",
			},
			wantErr: false,
		},
		{
			name: "extract nested JSON path",
			result: &ResponseResult{
				Body: `{"data": {"id": "123"}}`,
			},
			extractBlocks: []ExtractBlockModel{
				{
					Name:    types.StringValue("id"),
					JsonPath: types.StringValue("data.id"),
				},
			},
			want: map[string]string{
				"id": "123",
			},
			wantErr: false,
		},
		{
			name: "extract header",
			result: &ResponseResult{
				Body: `{}`,
				Headers: map[string]string{
					"X-Request-ID": "abc123",
				},
			},
			extractBlocks: []ExtractBlockModel{
				{
					Name:   types.StringValue("request_id"),
					Header: types.StringValue("X-Request-ID"),
				},
			},
			want: map[string]string{
				"request_id": "abc123",
			},
			wantErr: false,
		},
		{
			name: "extract multiple values",
			result: &ResponseResult{
				Body: `{"id": "123", "status": "ready"}`,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
			},
			extractBlocks: []ExtractBlockModel{
				{
					Name:    types.StringValue("id"),
					JsonPath: types.StringValue("id"),
				},
				{
					Name:    types.StringValue("status"),
					JsonPath: types.StringValue("status"),
				},
			},
			want: map[string]string{
				"id":     "123",
				"status": "ready",
			},
			wantErr: false,
		},
		{
			name: "invalid JSON path",
			result: &ResponseResult{
				Body: `{"id": "123"}`,
			},
			extractBlocks: []ExtractBlockModel{
				{
					Name:    types.StringValue("missing"),
					JsonPath: types.StringValue("nonexistent.path"),
				},
			},
			want: map[string]string{
				"missing": "",
			},
			wantErr: false, // Errors are logged but don't fail extraction
		},
		{
			name: "missing header",
			result: &ResponseResult{
				Body: `{}`,
				Headers: map[string]string{
					"Other-Header": "value",
				},
			},
			extractBlocks: []ExtractBlockModel{
				{
					Name:   types.StringValue("missing"),
					Header: types.StringValue("X-Missing"),
				},
			},
			want: map[string]string{
				"missing": "",
			},
			wantErr: false,
		},
		{
			name: "non-JSON body with JSON path",
			result: &ResponseResult{
				Body: "plain text",
			},
			extractBlocks: []ExtractBlockModel{
				{
					Name:    types.StringValue("id"),
					JsonPath: types.StringValue("id"),
				},
			},
			want: map[string]string{
				"id": "",
			},
			wantErr: false,
		},
		{
			name: "empty extract blocks",
			result: &ResponseResult{
				Body: `{"id": "123"}`,
			},
			extractBlocks: []ExtractBlockModel{},
			want:          map[string]string{},
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractValues(context.Background(), tt.result, tt.extractBlocks)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractValues() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("ExtractValues() [%s] = %v, want %v", k, got[k], v)
				}
			}
			// Check that we don't have extra keys
			for k := range got {
				if _, exists := tt.want[k]; !exists {
					t.Errorf("ExtractValues() returned unexpected key: %s", k)
				}
			}
		})
	}
}

