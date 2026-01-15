package provider

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// InterpolationContext holds state values available for template expansion
type InterpolationContext struct {
	ID           string            // self.id
	Outputs      map[string]string // self.outputs.KEY
	ResponseBody string            // self.response_body
	StatusCode   int64             // self.status_code
}

// InterpolateString replaces ${self.KEY} patterns with values from state context
// Supported patterns:
//   - ${self.id}
//   - ${self.outputs.KEY}
//   - ${self.response_body}
//   - ${self.status_code}
func InterpolateString(ctx context.Context, text string, interpolCtx *InterpolationContext) (string, error) {
	if text == "" || interpolCtx == nil {
		return text, nil
	}

	result := text
	var lastErr error

	// Pattern: ${self.outputs.KEY}
	outputsRegex := regexp.MustCompile(`\$\{self\.outputs\.([a-zA-Z0-9_]+)\}`)
	result = outputsRegex.ReplaceAllStringFunc(result, func(match string) string {
		submatches := outputsRegex.FindStringSubmatch(match)
		if len(submatches) < 2 {
			return match
		}
		key := submatches[1]
		if val, ok := interpolCtx.Outputs[key]; ok {
			tflog.Trace(ctx, fmt.Sprintf("Interpolated ${self.outputs.%s} -> %s", key, val))
			return val
		}
		lastErr = fmt.Errorf("output key not found: %s", key)
		return match
	})

	if lastErr != nil {
		return "", lastErr
	}

	// Pattern: ${self.id}
	result = strings.ReplaceAll(result, "${self.id}", interpolCtx.ID)
	if strings.Contains(text, "${self.id}") {
		tflog.Trace(ctx, fmt.Sprintf("Interpolated ${self.id} -> %s", interpolCtx.ID))
	}

	return result, nil
}

// InterpolateStringValue applies interpolation to a Terraform StringValue
func InterpolateStringValue(ctx context.Context, val types.String, interpolCtx *InterpolationContext) (types.String, error) {
	if val.IsNull() || val.IsUnknown() {
		return val, nil
	}

	expanded, err := InterpolateString(ctx, val.ValueString(), interpolCtx)
	if err != nil {
		return types.StringNull(), err
	}

	return types.StringValue(expanded), nil
}

// InterpolateMap applies interpolation to all values in a map
func InterpolateMap(ctx context.Context, m map[string]string, interpolCtx *InterpolationContext) (map[string]string, error) {
	if len(m) == 0 || interpolCtx == nil {
		return m, nil
	}

	result := make(map[string]string)
	for key, val := range m {
		expanded, err := InterpolateString(ctx, val, interpolCtx)
		if err != nil {
			return nil, fmt.Errorf("failed to interpolate map value for key %q: %w", key, err)
		}
		result[key] = expanded
	}

	return result, nil
}

// InterpolateHeaderBlocks applies interpolation to header block values
func InterpolateHeaderBlocks(ctx context.Context, blocks []HeaderBlockModel, interpolCtx *InterpolationContext) ([]HeaderBlockModel, error) {
	if len(blocks) == 0 || interpolCtx == nil {
		return blocks, nil
	}

	result := make([]HeaderBlockModel, 0, len(blocks))
	for _, block := range blocks {
		expandedValue, err := InterpolateString(ctx, block.Value.ValueString(), interpolCtx)
		if err != nil {
			return nil, fmt.Errorf("failed to interpolate header %q: %w", block.Name.ValueString(), err)
		}
		result = append(result, HeaderBlockModel{
			Name:  block.Name,
			Value: types.StringValue(expandedValue),
		})
	}

	return result, nil
}

// BuildInterpolationContextFromState creates an InterpolationContext from resource state
func BuildInterpolationContextFromState(ctx context.Context, state *HttpxRequestResourceModel) (*InterpolationContext, error) {
	interpolCtx := &InterpolationContext{
		ID:         state.Id.ValueString(),
		Outputs:    make(map[string]string),
		StatusCode: state.StatusCode.ValueInt64(),
	}

	// Extract outputs from state
	if !state.Outputs.IsNull() {
		outputsMap := make(map[string]types.String)
		diags := state.Outputs.ElementsAs(ctx, &outputsMap, false)
		if diags.HasError() {
			return nil, fmt.Errorf("failed to parse outputs from state")
		}
		for key, val := range outputsMap {
			interpolCtx.Outputs[key] = val.ValueString()
		}
	}

	// Extract response body if available
	if !state.ResponseBody.IsNull() {
		interpolCtx.ResponseBody = state.ResponseBody.ValueString()
	}

	return interpolCtx, nil
}

