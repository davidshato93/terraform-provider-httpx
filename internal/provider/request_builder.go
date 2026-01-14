package provider

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// RequestConfig holds the configuration for building an HTTP request
type RequestConfig struct {
	Url                string
	Method             string
	Headers            map[string]string
	HeaderBlocks       []HeaderBlockModel
	Query              map[string]string
	Body               types.String
	BodyJson           types.String
	BodyFile           types.String
	BasicAuth          *ResourceBasicAuthModel
	BearerToken        types.String
	ProviderDefaults   *ProviderConfig
}

// BuildRequest constructs an HTTP request from the configuration
func BuildRequest(ctx context.Context, config *RequestConfig) (*http.Request, error) {
	// Parse URL
	reqURL, err := url.Parse(config.Url)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Add query parameters
	if len(config.Query) > 0 {
		q := reqURL.Query()
		for k, v := range config.Query {
			q.Add(k, v)
		}
		reqURL.RawQuery = q.Encode()
	}

	// Determine request body
	var bodyReader io.Reader
	contentTypeSet := false

	// Check for body conflicts
	bodyCount := 0
	if !config.Body.IsNull() && !config.Body.IsUnknown() && config.Body.ValueString() != "" {
		bodyCount++
	}
	if !config.BodyJson.IsNull() && !config.BodyJson.IsUnknown() && config.BodyJson.ValueString() != "" {
		bodyCount++
	}
	if !config.BodyFile.IsNull() && !config.BodyFile.IsUnknown() && config.BodyFile.ValueString() != "" {
		bodyCount++
	}

	if bodyCount > 1 {
		return nil, fmt.Errorf("only one of body, body_json, or body_file can be set")
	}

	// Set body
	if !config.Body.IsNull() && !config.Body.IsUnknown() && config.Body.ValueString() != "" {
		bodyReader = strings.NewReader(config.Body.ValueString())
	} else if !config.BodyJson.IsNull() && !config.BodyJson.IsUnknown() && config.BodyJson.ValueString() != "" {
		// Parse JSON to validate and pretty-print
		var jsonData interface{}
		if err := json.Unmarshal([]byte(config.BodyJson.ValueString()), &jsonData); err != nil {
			return nil, fmt.Errorf("invalid JSON in body_json: %w", err)
		}
		jsonBytes, err := json.Marshal(jsonData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBytes)
		// Set Content-Type if not already set
		if config.Headers["Content-Type"] == "" {
			contentTypeSet = true
		}
	} else if !config.BodyFile.IsNull() && !config.BodyFile.IsUnknown() && config.BodyFile.ValueString() != "" {
		filePath := config.BodyFile.ValueString()
		file, err := os.Open(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to open body_file: %w", err)
		}
		defer func() {
			if err := file.Close(); err != nil {
				tflog.Warn(ctx, "Failed to close file", map[string]interface{}{"error": err})
			}
		}()
		bodyBytes, err := io.ReadAll(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read body_file: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// Create request
	var req *http.Request
	if bodyReader != nil {
		req, err = http.NewRequestWithContext(ctx, config.Method, reqURL.String(), bodyReader)
	} else {
		req, err = http.NewRequestWithContext(ctx, config.Method, reqURL.String(), nil)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Merge headers: provider defaults first, then resource headers, then header blocks
	headers := make(map[string][]string)

	// Add provider default headers
	if config.ProviderDefaults != nil && config.ProviderDefaults.DefaultHeaders != nil {
		for k, v := range config.ProviderDefaults.DefaultHeaders {
			headers[strings.ToLower(k)] = []string{v}
		}
	}

	// Add resource headers (overrides provider defaults)
	if config.Headers != nil {
		for k, v := range config.Headers {
			headers[strings.ToLower(k)] = []string{v}
		}
	}

	// Add header blocks (allows multiple values for same header)
	for _, hb := range config.HeaderBlocks {
		if !hb.Name.IsNull() && !hb.Value.IsNull() {
			key := strings.ToLower(hb.Name.ValueString())
			if existing, ok := headers[key]; ok {
				headers[key] = append(existing, hb.Value.ValueString())
			} else {
				headers[key] = []string{hb.Value.ValueString()}
			}
		}
	}

	// Set Content-Type for JSON if needed
	if contentTypeSet {
		headers["content-type"] = []string{"application/json"}
	}

	// Apply headers to request
	for k, values := range headers {
		for _, v := range values {
			req.Header.Add(k, v)
		}
	}

	// Set authentication
	// config.BasicAuth uses BasicAuthModel from models.go which has types.String fields
	if config.BasicAuth != nil {
		username := ""
		password := ""
		if !config.BasicAuth.Username.IsNull() && !config.BasicAuth.Username.IsUnknown() {
			username = config.BasicAuth.Username.ValueString()
		}
		if !config.BasicAuth.Password.IsNull() && !config.BasicAuth.Password.IsUnknown() {
			password = config.BasicAuth.Password.ValueString()
		}
		if username != "" && password != "" {
			auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
			req.Header.Set("Authorization", "Basic "+auth)
		}
	} else if config.ProviderDefaults != nil && config.ProviderDefaults.BasicAuth != nil {
		// ProviderDefaults.BasicAuth uses BasicAuthModel from provider.go (string fields)
		username := config.ProviderDefaults.BasicAuth.Username
		password := config.ProviderDefaults.BasicAuth.Password
		if username != "" && password != "" {
			auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
			req.Header.Set("Authorization", "Basic "+auth)
		}
	}

	if !config.BearerToken.IsNull() && !config.BearerToken.IsUnknown() && config.BearerToken.ValueString() != "" {
		req.Header.Set("Authorization", "Bearer "+config.BearerToken.ValueString())
	} else if config.ProviderDefaults != nil && config.ProviderDefaults.BearerToken != nil {
		req.Header.Set("Authorization", "Bearer "+*config.ProviderDefaults.BearerToken)
	}

	tflog.Debug(ctx, "Built HTTP request", map[string]interface{}{
		"method": req.Method,
		"url":    req.URL.String(),
	})

	return req, nil
}

// ConvertTerraformMap converts a Terraform types.Map to a Go map[string]string
func ConvertTerraformMap(ctx context.Context, tfMap types.Map) (map[string]string, error) {
	if tfMap.IsNull() || tfMap.IsUnknown() {
		return nil, nil
	}

	result := make(map[string]string)
	elements := tfMap.Elements()
	for k, v := range elements {
		if strVal, ok := v.(types.String); ok {
			result[k] = strVal.ValueString()
		}
	}
	return result, nil
}

// ConvertTerraformList converts a Terraform types.List to a Go slice
func ConvertTerraformList[T any](ctx context.Context, tfList types.List, converter func(interface{}) (T, error)) ([]T, error) {
	if tfList.IsNull() || tfList.IsUnknown() {
		return nil, nil
	}

	var result []T
	elements := tfList.Elements()
	for _, elem := range elements {
		converted, err := converter(elem)
		if err != nil {
			return nil, err
		}
		result = append(result, converted)
	}
	return result, nil
}

