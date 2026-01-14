package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// RetryUntilConfig holds conditional retry configuration
type RetryUntilConfig struct {
	StatusCodes    []int64
	JsonPathEquals map[string]string
	HeaderEquals   map[string]string
	BodyRegex      string
}

// EvaluateRetryUntil checks if all retry_until conditions are satisfied
func (ruc *RetryUntilConfig) EvaluateRetryUntil(ctx context.Context, result *ResponseResult) (bool, []string) {
	if ruc == nil {
		return true, nil // No conditions means always satisfied
	}

	var unsatisfied []string

	// Check status codes
	if len(ruc.StatusCodes) > 0 {
		found := false
		for _, code := range ruc.StatusCodes {
			if result.StatusCode == code {
				found = true
				break
			}
		}
		if !found {
			unsatisfied = append(unsatisfied, fmt.Sprintf("status code %d not in required codes %v", result.StatusCode, ruc.StatusCodes))
		}
	}

	// Check JSON path conditions
	if len(ruc.JsonPathEquals) > 0 {
		if !checkJsonPathConditions(ctx, result.Body, ruc.JsonPathEquals) {
			unsatisfied = append(unsatisfied, "JSON path conditions not satisfied")
		}
	}

	// Check header conditions
	if len(ruc.HeaderEquals) > 0 {
		if !checkHeaderConditions(result.Headers, ruc.HeaderEquals) {
			unsatisfied = append(unsatisfied, "header conditions not satisfied")
		}
	}

	// Check body regex
	if ruc.BodyRegex != "" {
		matched, err := regexp.MatchString(ruc.BodyRegex, result.Body)
		if err != nil {
			unsatisfied = append(unsatisfied, fmt.Sprintf("invalid regex pattern: %v", err))
		} else if !matched {
			unsatisfied = append(unsatisfied, fmt.Sprintf("body does not match regex: %s", ruc.BodyRegex))
		}
	}

	return len(unsatisfied) == 0, unsatisfied
}

// checkJsonPathConditions evaluates JSON path conditions
func checkJsonPathConditions(ctx context.Context, body string, conditions map[string]string) bool {
	if body == "" {
		return false
	}

	var jsonData interface{}
	if err := json.Unmarshal([]byte(body), &jsonData); err != nil {
		tflog.Debug(ctx, "Failed to parse JSON for path evaluation", map[string]interface{}{
			"error": err.Error(),
		})
		return false
	}

	for path, expectedValue := range conditions {
		actualValue, err := evaluateJsonPath(jsonData, path)
		if err != nil {
			tflog.Debug(ctx, "JSON path evaluation failed", map[string]interface{}{
				"path": path,
				"error": err.Error(),
			})
			return false
		}

		// Convert actual value to string for comparison
		actualStr := fmt.Sprintf("%v", actualValue)
		
		// Try to parse expected value as JSON to handle booleans/numbers properly
		var expectedParsed interface{}
		if err := json.Unmarshal([]byte(expectedValue), &expectedParsed); err == nil {
			// Successfully parsed as JSON, compare parsed values
			if fmt.Sprintf("%v", expectedParsed) != fmt.Sprintf("%v", actualValue) {
				return false
			}
		} else {
			// Not valid JSON, compare as strings
			if actualStr != expectedValue {
				return false
			}
		}
	}

	return true
}

// evaluateJsonPath evaluates a dot-path expression on JSON data
// Supports simple dot notation: "data.isAttached", "items[0].id"
func evaluateJsonPath(data interface{}, path string) (interface{}, error) {
	if path == "" {
		return data, nil
	}

	parts := strings.Split(path, ".")
	current := data

	for i, part := range parts {
		// Check for array index notation: "items[0]"
		if idx := strings.Index(part, "["); idx != -1 {
			key := part[:idx]
			idxStr := part[idx+1 : len(part)-1] // Extract index between [ and ]
			
			// Navigate to the array
			if key != "" {
				if m, ok := current.(map[string]interface{}); ok {
					if val, exists := m[key]; exists {
						current = val
					} else {
						return nil, fmt.Errorf("key '%s' not found at path '%s'", key, strings.Join(parts[:i+1], "."))
					}
				} else {
					return nil, fmt.Errorf("expected object at path '%s'", strings.Join(parts[:i], "."))
				}
			}

			// Access array element
			if arr, ok := current.([]interface{}); ok {
				var idx int
				if _, err := fmt.Sscanf(idxStr, "%d", &idx); err != nil {
					return nil, fmt.Errorf("invalid array index '%s'", idxStr)
				}
				if idx < 0 || idx >= len(arr) {
					return nil, fmt.Errorf("array index %d out of bounds (length: %d)", idx, len(arr))
				}
				current = arr[idx]
			} else {
				return nil, fmt.Errorf("expected array at path '%s'", strings.Join(parts[:i], "."))
			}
		} else {
			// Regular key access
			if m, ok := current.(map[string]interface{}); ok {
				if val, exists := m[part]; exists {
					current = val
				} else {
					return nil, fmt.Errorf("key '%s' not found at path '%s'", part, strings.Join(parts[:i+1], "."))
				}
			} else {
				return nil, fmt.Errorf("expected object at path '%s', got %T", strings.Join(parts[:i], "."), current)
			}
		}
	}

	return current, nil
}

// checkHeaderConditions checks if header conditions are satisfied
func checkHeaderConditions(headers map[string]string, conditions map[string]string) bool {
	for headerName, expectedValue := range conditions {
		found := false
		for k, v := range headers {
			if strings.EqualFold(k, headerName) {
				if v == expectedValue {
					found = true
					break
				}
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// BuildRetryUntilConfig converts RetryUntilModel to RetryUntilConfig
func BuildRetryUntilConfig(ctx context.Context, retryUntilModel *RetryUntilModel) *RetryUntilConfig {
	if retryUntilModel == nil {
		return nil
	}

	config := &RetryUntilConfig{
		StatusCodes:    []int64{},
		JsonPathEquals: make(map[string]string),
		HeaderEquals:   make(map[string]string),
		BodyRegex:      "",
	}

	// Parse status codes
	if !retryUntilModel.StatusCodes.IsNull() && !retryUntilModel.StatusCodes.IsUnknown() {
		codes, err := ConvertTerraformList(ctx, retryUntilModel.StatusCodes, func(v interface{}) (int64, error) {
			if intVal, ok := v.(types.Int64); ok {
				return intVal.ValueInt64(), nil
			}
			return 0, fmt.Errorf("expected int64, got %T", v)
		})
		if err == nil {
			config.StatusCodes = codes
		}
	}

	// Parse JSON path conditions
	if !retryUntilModel.JsonPathEquals.IsNull() && !retryUntilModel.JsonPathEquals.IsUnknown() {
		elements := retryUntilModel.JsonPathEquals.Elements()
		for k, v := range elements {
			if strVal, ok := v.(types.String); ok {
				config.JsonPathEquals[k] = strVal.ValueString()
			}
		}
	}

	// Parse header conditions
	if !retryUntilModel.HeaderEquals.IsNull() && !retryUntilModel.HeaderEquals.IsUnknown() {
		elements := retryUntilModel.HeaderEquals.Elements()
		for k, v := range elements {
			if strVal, ok := v.(types.String); ok {
				config.HeaderEquals[k] = strVal.ValueString()
			}
		}
	}

	// Parse body regex
	if !retryUntilModel.BodyRegex.IsNull() && !retryUntilModel.BodyRegex.IsUnknown() {
		config.BodyRegex = retryUntilModel.BodyRegex.ValueString()
	}

	return config
}

