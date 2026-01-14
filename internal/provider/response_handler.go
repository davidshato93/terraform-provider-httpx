package provider

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/davidshato/terraform-provider-httpx/internal/client"
	"github.com/davidshato/terraform-provider-httpx/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// ResponseResult holds the result of an HTTP request
type ResponseResult struct {
	StatusCode      int64
	Headers         map[string]string
	Body            string
	AttemptCount    int64
	Error           string
}

// ExecuteRequest executes an HTTP request and returns the response
func ExecuteRequest(ctx context.Context, req *http.Request, providerConfig *ProviderConfig) (*ResponseResult, error) {
	// Convert to config.ProviderConfig
	cfg := providerConfig.ToConfigProviderConfig()

	// Create HTTP client
	httpClient, err := client.NewHTTPClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Execute request
	httpResp, err := httpClient.Do(req)
	if err != nil {
		return &ResponseResult{
			StatusCode:   0,
			AttemptCount:  1,
			Error:        utils.RedactError(err.Error(), cfg.RedactHeaders),
		}, fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		if err := httpResp.Body.Close(); err != nil {
			tflog.Warn(ctx, "Failed to close response body", map[string]interface{}{"error": err})
		}
	}()

	// Read response body with size limit
	limitedReader := client.LimitReader(httpResp.Body, cfg.MaxResponseBodyBytes)
	bodyBytes, err := io.ReadAll(limitedReader)
	if err != nil {
		return &ResponseResult{
			StatusCode:   int64(httpResp.StatusCode),
			AttemptCount: 1,
			Error:        utils.RedactError(err.Error(), cfg.RedactHeaders),
		}, fmt.Errorf("failed to read response body: %w", err)
	}

	bodyStr := string(bodyBytes)
	
	// Truncate if needed
	if int64(len(bodyBytes)) >= cfg.MaxResponseBodyBytes {
		bodyStr = utils.TruncateString(bodyStr, int(cfg.MaxResponseBodyBytes))
	}

	// Extract headers
	headers := make(map[string]string)
	for k, v := range httpResp.Header {
		// Join multiple values with comma
		headers[k] = strings.Join(v, ", ")
	}

	result := &ResponseResult{
		StatusCode:   int64(httpResp.StatusCode),
		Headers:      headers,
		Body:         bodyStr,
		AttemptCount: 1,
	}

	tflog.Debug(ctx, "HTTP request completed", map[string]interface{}{
		"status_code": result.StatusCode,
		"body_size":   len(bodyBytes),
	})

	return result, nil
}

// ValidateExpectations validates response expectations
func ValidateExpectations(ctx context.Context, result *ResponseResult, expect *ExpectModel) error {
	if expect == nil {
		return nil
	}

	var errors []string

	// Validate status codes
	if !expect.StatusCodes.IsNull() && !expect.StatusCodes.IsUnknown() {
		expectedCodes, err := ConvertTerraformList(ctx, expect.StatusCodes, func(v interface{}) (int64, error) {
			if intVal, ok := v.(types.Int64); ok {
				return intVal.ValueInt64(), nil
			}
			return 0, fmt.Errorf("expected int64, got %T", v)
		})
		if err == nil {
			found := false
			for _, code := range expectedCodes {
				if code == result.StatusCode {
					found = true
					break
				}
			}
			if !found {
				errors = append(errors, fmt.Sprintf("status code %d not in expected codes %v", result.StatusCode, expectedCodes))
			}
		}
	}

	// Validate header presence
	if !expect.HeaderPresent.IsNull() && !expect.HeaderPresent.IsUnknown() {
		requiredHeaders, err := ConvertTerraformList(ctx, expect.HeaderPresent, func(v interface{}) (string, error) {
			if strVal, ok := v.(types.String); ok {
				return strVal.ValueString(), nil
			}
			return "", fmt.Errorf("expected string, got %T", v)
		})
		if err == nil {
			for _, headerName := range requiredHeaders {
				found := false
				for k := range result.Headers {
					if strings.EqualFold(k, headerName) {
						found = true
						break
					}
				}
				if !found {
					errors = append(errors, fmt.Sprintf("required header '%s' not present", headerName))
				}
			}
		}
	}

	// TODO: Implement json_path_exists and json_path_equals in Phase 4/5

	if len(errors) > 0 {
		return fmt.Errorf("expectation validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

