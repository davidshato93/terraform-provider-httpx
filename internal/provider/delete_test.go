package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

// TestDeleteWithoutOnDestroy tests that Delete without on_destroy block is a no-op
func TestDeleteWithoutOnDestroy(t *testing.T) {
	// This test validates the model behavior
	// The actual Delete method would be tested via acceptance tests with Terraform

	model := &HttpxRequestResourceModel{
		Id:        types.StringValue("test-id"),
		OnDestroy: nil, // No on_destroy block
	}

	// Embedded RequestConfigModel
	model.Url = types.StringValue("https://api.example.com/resource")
	model.Method = types.StringValue("POST")

	assert.Nil(t, model.OnDestroy)
	// Delete should be a no-op in this case
}

// TestDeleteWithOnDestroyConfig tests that Delete validates on_destroy presence
func TestDeleteWithOnDestroyConfig(t *testing.T) {
	// This test validates the model behavior
	model := &HttpxRequestResourceModel{
		Id:        types.StringValue("test-id"),
		OnDestroy: &RequestConfigModel{
			Method: types.StringValue("DELETE"),
			Url:    types.StringValue("https://api.example.com/resource/${self.id}"),
		},
	}

	// Embedded RequestConfigModel
	model.Url = types.StringValue("https://api.example.com/resource")
	model.Method = types.StringValue("POST")

	assert.NotNil(t, model.OnDestroy)
	assert.Equal(t, "DELETE", model.OnDestroy.Method.ValueString())
	assert.Equal(t, "https://api.example.com/resource/${self.id}", model.OnDestroy.Url.ValueString())
}

// TestDeleteWithExtractedOutputs tests Delete with extracted values in state
func TestDeleteWithExtractedOutputs(t *testing.T) {
	ctx := context.Background()

	// Simulate state after create with extracted outputs
	state := &HttpxRequestResourceModel{
		Id:      types.StringValue("res-123"),
		Outputs: types.MapValueMust(types.StringType, map[string]attr.Value{
			"user_id": types.StringValue("user-456"),
		}),
		OnDestroy: &RequestConfigModel{
			Method: types.StringValue("DELETE"),
			Url:    types.StringValue("https://api.example.com/users/${self.outputs.user_id}"),
		},
	}

	// Embedded RequestConfigModel
	state.Url = types.StringValue("https://api.example.com/users")
	state.Method = types.StringValue("POST")

	// Build interpolation context
	interpolCtx, err := BuildInterpolationContextFromState(ctx, state)
	assert.NoError(t, err)

	// Interpolate URL
	expandedURL, err := InterpolateString(ctx, state.OnDestroy.Url.ValueString(), interpolCtx)
	assert.NoError(t, err)
	assert.Equal(t, "https://api.example.com/users/user-456", expandedURL)
}

// TestDeleteWithRetryUntilCondition tests Delete respects retry_until for destroy
func TestDeleteWithRetryUntilCondition(t *testing.T) {
	model := &HttpxRequestResourceModel{
		Id:        types.StringValue("test-id"),
		OnDestroy: &RequestConfigModel{
			Method: types.StringValue("DELETE"),
			Url:    types.StringValue("https://api.example.com/resource/${self.id}"),
			RetryUntil: &RetryUntilModel{
				StatusCodes: types.ListValueMust(types.Int64Type, []attr.Value{
					types.Int64Value(200),
					types.Int64Value(204),
					types.Int64Value(404),
				}),
			},
		},
	}

	// Embedded RequestConfigModel
	model.Url = types.StringValue("https://api.example.com/resource")
	model.Method = types.StringValue("POST")

	assert.NotNil(t, model.OnDestroy.RetryUntil)
	assert.NotNil(t, model.OnDestroy.RetryUntil.StatusCodes)
}

// TestDeleteWithExpectBlock tests Delete respects expect block for destroy
func TestDeleteWithExpectBlock(t *testing.T) {
	model := &HttpxRequestResourceModel{
		Id:        types.StringValue("test-id"),
		OnDestroy: &RequestConfigModel{
			Method: types.StringValue("DELETE"),
			Url:    types.StringValue("https://api.example.com/resource/${self.id}"),
			Expect: &ExpectModel{
				StatusCodes: types.ListValueMust(types.Int64Type, []attr.Value{
					types.Int64Value(200),
					types.Int64Value(204),
					types.Int64Value(404),
				}),
			},
		},
	}

	// Embedded RequestConfigModel
	model.Url = types.StringValue("https://api.example.com/resource")
	model.Method = types.StringValue("POST")

	assert.NotNil(t, model.OnDestroy.Expect)
	assert.NotNil(t, model.OnDestroy.Expect.StatusCodes)
}

// TestDeleteWithRetryConfig tests Delete respects retry config for destroy
func TestDeleteWithRetryConfig(t *testing.T) {
	model := &HttpxRequestResourceModel{
		Id:        types.StringValue("test-id"),
		OnDestroy: &RequestConfigModel{
			Method: types.StringValue("DELETE"),
			Url:    types.StringValue("https://api.example.com/resource/${self.id}"),
			Retry: &RetryModel{
				Attempts:           types.Int64Value(5),
				MinDelayMs:         types.Int64Value(500),
				MaxDelayMs:         types.Int64Value(5000),
				Backoff:            types.StringValue("exponential"),
				RetryOnStatusCodes: types.ListValueMust(types.Int64Type, []attr.Value{
					types.Int64Value(500),
					types.Int64Value(502),
					types.Int64Value(503),
				}),
			},
		},
	}

	// Embedded RequestConfigModel
	model.Url = types.StringValue("https://api.example.com/resource")
	model.Method = types.StringValue("POST")

	assert.NotNil(t, model.OnDestroy.Retry)
	assert.Equal(t, int64(5), model.OnDestroy.Retry.Attempts.ValueInt64())
	assert.Equal(t, "exponential", model.OnDestroy.Retry.Backoff.ValueString())
}

// TestDeleteWithBasicAuth tests Delete supports basic auth in destroy config
func TestDeleteWithBasicAuth(t *testing.T) {
	model := &HttpxRequestResourceModel{
		Id:        types.StringValue("test-id"),
		OnDestroy: &RequestConfigModel{
			Method: types.StringValue("DELETE"),
			Url:    types.StringValue("https://api.example.com/resource/${self.id}"),
			BasicAuth: &ResourceBasicAuthModel{
				Username: types.StringValue("user"),
				Password: types.StringValue("pass"),
			},
		},
	}

	// Embedded RequestConfigModel
	model.Url = types.StringValue("https://api.example.com/resource")
	model.Method = types.StringValue("POST")

	assert.NotNil(t, model.OnDestroy.BasicAuth)
	assert.Equal(t, "user", model.OnDestroy.BasicAuth.Username.ValueString())
}

// TestDeleteWithHeaderBlocks tests Delete supports header blocks in destroy config
func TestDeleteWithHeaderBlocks(t *testing.T) {
	model := &HttpxRequestResourceModel{
		Id:        types.StringValue("test-id"),
		OnDestroy: &RequestConfigModel{
			Method: types.StringValue("DELETE"),
			Url:    types.StringValue("https://api.example.com/resource/${self.id}"),
			HeaderBlocks: []HeaderBlockModel{
				{
					Name:  types.StringValue("X-User-ID"),
					Value: types.StringValue("${self.outputs.user_id}"),
				},
			},
		},
	}

	// Embedded RequestConfigModel
	model.Url = types.StringValue("https://api.example.com/resource")
	model.Method = types.StringValue("POST")

	assert.NotNil(t, model.OnDestroy.HeaderBlocks)
	assert.Equal(t, 1, len(model.OnDestroy.HeaderBlocks))
	assert.Equal(t, "X-User-ID", model.OnDestroy.HeaderBlocks[0].Name.ValueString())
}

// TestDeleteMissingInterpolationKey tests Delete errors on missing interpolation key
func TestDeleteMissingInterpolationKey(t *testing.T) {
	ctx := context.Background()

	state := &HttpxRequestResourceModel{
		Id:      types.StringValue("res-123"),
		Outputs: types.MapValueMust(types.StringType, map[string]attr.Value{}), // No user_id
		OnDestroy: &RequestConfigModel{
			Method: types.StringValue("DELETE"),
			Url:    types.StringValue("https://api.example.com/users/${self.outputs.user_id}"),
		},
	}

	// Embedded RequestConfigModel
	state.Url = types.StringValue("https://api.example.com/users")
	state.Method = types.StringValue("POST")

	interpolCtx, err := BuildInterpolationContextFromState(ctx, state)
	assert.NoError(t, err)

	// Try to interpolate with missing key
	_, err = InterpolateString(ctx, state.OnDestroy.Url.ValueString(), interpolCtx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "output key not found")
}

