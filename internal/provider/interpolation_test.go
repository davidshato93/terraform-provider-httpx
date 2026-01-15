package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestInterpolateString(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		text         string
		interpolCtx  *InterpolationContext
		expected     string
		expectError  bool
	}{
		{
			name: "interpolate self.id",
			text: "https://api.example.com/users/${self.id}",
			interpolCtx: &InterpolationContext{
				ID:      "user-123",
				Outputs: make(map[string]string),
			},
			expected: "https://api.example.com/users/user-123",
		},
		{
			name: "interpolate self.outputs.KEY",
			text: "https://api.example.com/users/${self.outputs.user_id}",
			interpolCtx: &InterpolationContext{
				ID: "resource-123",
				Outputs: map[string]string{
					"user_id": "456",
				},
			},
			expected: "https://api.example.com/users/456",
		},
		{
			name: "interpolate multiple values",
			text: "${self.id}:${self.outputs.user_id}",
			interpolCtx: &InterpolationContext{
				ID: "res-1",
				Outputs: map[string]string{
					"user_id": "usr-2",
				},
			},
			expected: "res-1:usr-2",
		},
		{
			name: "no interpolation",
			text: "https://api.example.com/users/fixed",
			interpolCtx: &InterpolationContext{
				ID:      "user-123",
				Outputs: make(map[string]string),
			},
			expected: "https://api.example.com/users/fixed",
		},
		{
			name: "empty text",
			text: "",
			interpolCtx: &InterpolationContext{
				ID:      "user-123",
				Outputs: make(map[string]string),
			},
			expected: "",
		},
		{
			name: "missing output key",
			text: "${self.outputs.missing}",
			interpolCtx: &InterpolationContext{
				ID:      "res-1",
				Outputs: make(map[string]string),
			},
			expectError: true,
		},
		{
			name:        "nil context",
			text:        "${self.id}",
			interpolCtx: nil,
			expected:    "${self.id}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := InterpolateString(ctx, tt.text, tt.interpolCtx)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestInterpolateStringValue(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		val         types.String
		interpolCtx *InterpolationContext
		expected    string
		expectError bool
	}{
		{
			name: "interpolate string value",
			val:  types.StringValue("https://api.example.com/users/${self.outputs.user_id}"),
			interpolCtx: &InterpolationContext{
				ID: "resource-123",
				Outputs: map[string]string{
					"user_id": "456",
				},
			},
			expected: "https://api.example.com/users/456",
		},
		{
			name: "null value",
			val:  types.StringNull(),
			interpolCtx: &InterpolationContext{
				ID:      "user-123",
				Outputs: make(map[string]string),
			},
			expected: "",
		},
		{
			name: "missing key",
			val:  types.StringValue("${self.outputs.missing}"),
			interpolCtx: &InterpolationContext{
				ID:      "res-1",
				Outputs: make(map[string]string),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := InterpolateStringValue(ctx, tt.val, tt.interpolCtx)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.expected == "" && tt.val.IsNull() {
					assert.True(t, result.IsNull())
				} else {
					assert.Equal(t, tt.expected, result.ValueString())
				}
			}
		})
	}
}

func TestInterpolateMap(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		m            map[string]string
		interpolCtx  *InterpolationContext
		expected     map[string]string
		expectError  bool
	}{
		{
			name: "interpolate map values",
			m: map[string]string{
				"X-User-ID": "${self.outputs.user_id}",
				"X-Req-ID":  "request-${self.id}",
			},
			interpolCtx: &InterpolationContext{
				ID: "res-123",
				Outputs: map[string]string{
					"user_id": "user-456",
				},
			},
			expected: map[string]string{
				"X-User-ID": "user-456",
				"X-Req-ID":  "request-res-123",
			},
		},
		{
			name:         "empty map",
			m:            make(map[string]string),
			interpolCtx:  &InterpolationContext{ID: "res-1", Outputs: make(map[string]string)},
			expected:     make(map[string]string),
			expectError:  false,
		},
		{
			name: "missing key in map",
			m: map[string]string{
				"header": "${self.outputs.missing}",
			},
			interpolCtx: &InterpolationContext{
				ID:      "res-1",
				Outputs: make(map[string]string),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := InterpolateMap(ctx, tt.m, tt.interpolCtx)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestInterpolateHeaderBlocks(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		blocks      []HeaderBlockModel
		interpolCtx *InterpolationContext
		expected    map[string]string
		expectError bool
	}{
		{
			name: "interpolate header block values",
			blocks: []HeaderBlockModel{
				{
					Name:  types.StringValue("X-User-ID"),
					Value: types.StringValue("${self.outputs.user_id}"),
				},
				{
					Name:  types.StringValue("X-Req-ID"),
					Value: types.StringValue("request-${self.id}"),
				},
			},
			interpolCtx: &InterpolationContext{
				ID: "res-123",
				Outputs: map[string]string{
					"user_id": "user-456",
				},
			},
			expected: map[string]string{
				"X-User-ID": "user-456",
				"X-Req-ID":  "request-res-123",
			},
		},
		{
			name:        "empty blocks",
			blocks:      []HeaderBlockModel{},
			interpolCtx: &InterpolationContext{ID: "res-1", Outputs: make(map[string]string)},
			expected:    make(map[string]string),
		},
		{
			name: "missing key",
			blocks: []HeaderBlockModel{
				{
					Name:  types.StringValue("X-Missing"),
					Value: types.StringValue("${self.outputs.missing}"),
				},
			},
			interpolCtx: &InterpolationContext{
				ID:      "res-1",
				Outputs: make(map[string]string),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := InterpolateHeaderBlocks(ctx, tt.blocks, tt.interpolCtx)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if len(tt.expected) == 0 {
					assert.Equal(t, 0, len(result))
				} else {
			for _, block := range result {
				assert.Equal(t, tt.expected[block.Name.ValueString()], block.Value.ValueString())
			}
				}
			}
		})
	}
}

func TestBuildInterpolationContextFromState(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		state       *HttpxRequestResourceModel
		expectID    string
		expectOut   map[string]string
		expectError bool
	}{
		{
			name: "build context from state",
			state: &HttpxRequestResourceModel{
				Id:         types.StringValue("resource-123"),
				StatusCode: types.Int64Value(200),
				Outputs: types.MapValueMust(types.StringType, map[string]attr.Value{
					"user_id": types.StringValue("user-456"),
					"org_id":  types.StringValue("org-789"),
				}),
				ResponseBody: types.StringValue(`{"key":"value"}`),
			},
			expectID: "resource-123",
			expectOut: map[string]string{
				"user_id": "user-456",
				"org_id":  "org-789",
			},
		},
		{
			name: "null outputs",
			state: &HttpxRequestResourceModel{
				Id:         types.StringValue("resource-123"),
				StatusCode: types.Int64Value(200),
				Outputs:    types.MapNull(types.StringType),
			},
			expectID: "resource-123",
			expectOut: make(map[string]string),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := BuildInterpolationContextFromState(ctx, tt.state)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectID, result.ID)
				assert.Equal(t, tt.expectOut, result.Outputs)
			}
		})
	}
}

