package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// ExtractValues extracts values from response based on extract blocks
func ExtractValues(ctx context.Context, result *ResponseResult, extractBlocks []ExtractBlockModel) (map[string]string, error) {
	outputs := make(map[string]string)

	if len(extractBlocks) == 0 {
		return outputs, nil
	}

	// Parse JSON body if needed for JSON path extraction
	var jsonData interface{}
	hasJsonData := false
	if result.Body != "" {
		if err := json.Unmarshal([]byte(result.Body), &jsonData); err == nil {
			hasJsonData = true
		}
	}

	for _, extract := range extractBlocks {
		if extract.Name.IsNull() || extract.Name.IsUnknown() {
			continue
		}

		name := extract.Name.ValueString()
		if name == "" {
			continue
		}

		var value string

		// Extract from JSON path
		if !extract.JsonPath.IsNull() && !extract.JsonPath.IsUnknown() {
			jsonPath := extract.JsonPath.ValueString()
			if jsonPath != "" {
				if !hasJsonData {
					tflog.Debug(ctx, "Cannot extract JSON path, body is not valid JSON", map[string]interface{}{
						"name": name,
						"path": jsonPath,
					})
					outputs[name] = ""
					continue
				}

				extractedValue, extractErr := evaluateJsonPath(jsonData, jsonPath)
				if extractErr != nil {
					tflog.Debug(ctx, "Failed to extract JSON path", map[string]interface{}{
						"name":  name,
						"path":  jsonPath,
						"error": extractErr.Error(),
					})
					outputs[name] = ""
					continue
				}

				// Convert extracted value to string
				// Handle different types appropriately
				switch v := extractedValue.(type) {
				case string:
					value = v
				case bool:
					value = fmt.Sprintf("%t", v)
				case float64:
					// JSON numbers are float64
					value = fmt.Sprintf("%g", v)
				case nil:
					value = ""
				default:
					// For complex types, marshal to JSON string
					if jsonBytes, marshalErr := json.Marshal(v); marshalErr == nil {
						value = string(jsonBytes)
					} else {
						value = fmt.Sprintf("%v", v)
					}
				}
			}
		}

		// Extract from header (takes precedence if both are specified)
		if !extract.Header.IsNull() && !extract.Header.IsUnknown() {
			headerName := extract.Header.ValueString()
			if headerName != "" {
				found := false
				for k, v := range result.Headers {
					if strings.EqualFold(k, headerName) {
						value = v
						found = true
						break
					}
				}
				if !found {
					tflog.Debug(ctx, "Header not found for extraction", map[string]interface{}{
						"name":        name,
						"header_name": headerName,
					})
					value = ""
				}
			}
		}

		outputs[name] = value
		tflog.Debug(ctx, "Extracted value", map[string]interface{}{
			"name":  name,
			"value": value,
		})
	}

	return outputs, nil
}
