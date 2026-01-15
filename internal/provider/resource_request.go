package provider

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &HttpxRequestResource{}
var _ resource.ResourceWithConfigure = &HttpxRequestResource{}

type HttpxRequestResource struct {
	config *ProviderConfig
}

func NewHttpxRequestResource() resource.Resource {
	return &HttpxRequestResource{}
}

func (r *HttpxRequestResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_request"
}

func (r *HttpxRequestResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Resource for executing HTTP requests with retry logic and conditional polling",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Resource identifier",
			},
			"url": schema.StringAttribute{
				Required:    true,
				Description: "The URL to make the request to",
			},
			"method": schema.StringAttribute{
				Required:    true,
				Description: "HTTP method (GET, POST, PUT, PATCH, DELETE, etc.)",
			},
			"headers": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Request headers as a map",
			},
			"query": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Query parameters",
			},
			"body": schema.StringAttribute{
				Optional:    true,
				Description: "Raw request body (mutually exclusive with body_json and body_file)",
			},
			"body_json": schema.StringAttribute{
				Optional:    true,
				Description: "JSON-encodable object (mutually exclusive with body and body_file)",
			},
			"body_file": schema.StringAttribute{
				Optional:    true,
				Description: "Path to file to read and send (mutually exclusive with body and body_json)",
			},
			"bearer_token": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Bearer token for authentication",
			},
			"timeout_ms": schema.Int64Attribute{
				Optional:    true,
				Description: "Request timeout in milliseconds",
			},
			"insecure_skip_verify": schema.BoolAttribute{
				Optional:    true,
				Description: "Skip TLS certificate verification",
			},
			"proxy_url": schema.StringAttribute{
				Optional:    true,
				Description: "Proxy URL",
			},
			"response_sensitive": schema.BoolAttribute{
				Optional:    true,
				Description: "Mark response body as sensitive",
			},
			"store_response_body": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether to store response body in state. Defaults to true, but defaults to false if extract blocks are present (unless explicitly set to true).",
			},
			"read_mode": schema.StringAttribute{
				Optional:    true,
				Description: "Read behavior: 'none' or 'refresh'",
			},
			"status_code": schema.Int64Attribute{
				Computed:    true,
				Description: "HTTP status code",
			},
			"response_headers": schema.MapAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Response headers",
			},
			"response_body": schema.StringAttribute{
				Computed:    true,
				Sensitive:   false, // Will be set dynamically based on response_sensitive
				Description: "Response body",
			},
			"outputs": schema.MapAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Extracted values from extract blocks",
			},
			"last_attempt_count": schema.Int64Attribute{
				Computed:    true,
				Description: "Number of attempts made",
			},
			"last_error": schema.StringAttribute{
				Computed:    true,
				Description: "Last error message (redacted)",
			},
		},
		Blocks: map[string]schema.Block{
			"header": schema.ListNestedBlock{
				Description: "Repeated header blocks for multiple values with the same name",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "Header name",
						},
						"value": schema.StringAttribute{
							Required:    true,
							Description: "Header value",
						},
					},
				},
			},
			"basic_auth": schema.SingleNestedBlock{
				Description: "Basic authentication credentials",
				Attributes: map[string]schema.Attribute{
					"username": schema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: "Basic auth username",
					},
					"password": schema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: "Basic auth password",
					},
				},
			},
			"retry": schema.SingleNestedBlock{
				Description: "Retry configuration",
				Attributes: map[string]schema.Attribute{
					"attempts": schema.Int64Attribute{
						Optional:    true,
						Description: "Maximum number of retry attempts",
					},
					"min_delay_ms": schema.Int64Attribute{
						Optional:    true,
						Description: "Minimum delay between retries in milliseconds",
					},
					"max_delay_ms": schema.Int64Attribute{
						Optional:    true,
						Description: "Maximum delay between retries in milliseconds",
					},
					"backoff": schema.StringAttribute{
						Optional:    true,
						Description: "Backoff strategy: 'fixed', 'linear', or 'exponential'",
					},
					"jitter": schema.BoolAttribute{
						Optional:    true,
						Description: "Add jitter to retry delays",
					},
					"retry_on_status_codes": schema.ListAttribute{
						ElementType: types.Int64Type,
						Optional:    true,
						Description: "HTTP status codes that should trigger a retry",
					},
					"respect_retry_after": schema.BoolAttribute{
						Optional:    true,
						Description: "Respect Retry-After header if present",
					},
				},
			},
			"retry_until": schema.SingleNestedBlock{
				Description: "Conditional retry (poll-until) configuration",
				Attributes: map[string]schema.Attribute{
					"status_codes": schema.ListAttribute{
						ElementType: types.Int64Type,
						Optional:    true,
						Description: "Status codes that satisfy the condition",
					},
					"json_path_equals": schema.MapAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Description: "JSON path conditions that must equal specified values",
					},
					"header_equals": schema.MapAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Description: "Header conditions that must equal specified values",
					},
					"body_regex": schema.StringAttribute{
						Optional:    true,
						Description: "Regex pattern that must match the response body",
					},
				},
			},
			"expect": schema.SingleNestedBlock{
				Description: "Response expectations/validation",
				Attributes: map[string]schema.Attribute{
					"status_codes": schema.ListAttribute{
						ElementType: types.Int64Type,
						Optional:    true,
						Description: "Expected HTTP status codes",
					},
					"json_path_exists": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Description: "JSON paths that must exist",
					},
					"json_path_equals": schema.MapAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Description: "JSON path conditions that must equal specified values",
					},
					"header_present": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Description: "Headers that must be present",
					},
				},
			},
			"extract": schema.ListNestedBlock{
				Description: "Extract values from response",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "Name of the extracted value",
						},
						"json_path": schema.StringAttribute{
							Optional:    true,
							Description: "JSON path to extract from",
						},
						"header": schema.StringAttribute{
							Optional:    true,
							Description: "Header name to extract from",
						},
					},
				},
			},
			"timeouts": schema.SingleNestedBlock{
				Description: "Timeout configuration",
				Attributes: map[string]schema.Attribute{
					"create": schema.StringAttribute{
						Optional:    true,
						Description: "Timeout for create operation",
					},
					"read": schema.StringAttribute{
						Optional:    true,
						Description: "Timeout for read operation",
					},
					"update": schema.StringAttribute{
						Optional:    true,
						Description: "Timeout for update operation",
					},
					"delete": schema.StringAttribute{
						Optional:    true,
						Description: "Timeout for delete operation",
					},
				},
			},
			"on_destroy": schema.SingleNestedBlock{
				Description: "HTTP request to execute when resource is destroyed. Supports template interpolation with ${self.outputs.KEY} and ${self.id}",
				Attributes: map[string]schema.Attribute{
					"url": schema.StringAttribute{
						Optional:    true,
						Description: "The URL to make the destroy request to (supports ${self.outputs.KEY} and ${self.id} interpolation)",
					},
					"method": schema.StringAttribute{
						Optional:    true,
						Description: "HTTP method for destroy request",
					},
					"headers": schema.MapAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Description: "Request headers for destroy request",
					},
					"query": schema.MapAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Description: "Query parameters for destroy request",
					},
					"body": schema.StringAttribute{
						Optional:    true,
						Description: "Raw request body for destroy request",
					},
					"body_json": schema.StringAttribute{
						Optional:    true,
						Description: "JSON request body for destroy request",
					},
					"body_file": schema.StringAttribute{
						Optional:    true,
						Description: "Path to file to read for destroy request body",
					},
					"bearer_token": schema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: "Bearer token for destroy request",
					},
					"timeout_ms": schema.Int64Attribute{
						Optional:    true,
						Description: "Request timeout for destroy request in milliseconds",
					},
					"insecure_skip_verify": schema.BoolAttribute{
						Optional:    true,
						Description: "Skip TLS certificate verification for destroy request",
					},
					"proxy_url": schema.StringAttribute{
						Optional:    true,
						Description: "Proxy URL for destroy request",
					},
					"response_sensitive": schema.BoolAttribute{
						Optional:    true,
						Description: "Mark destroy response body as sensitive",
					},
					"store_response_body": schema.BoolAttribute{
						Optional:    true,
						Description: "Whether to store destroy response body (not persisted to state since resource is deleted)",
					},
				},
				Blocks: map[string]schema.Block{
					"header": schema.ListNestedBlock{
						Description: "Repeated header blocks for destroy request",
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Required:    true,
									Description: "Header name",
								},
								"value": schema.StringAttribute{
									Required:    true,
									Description: "Header value (supports ${self.outputs.KEY} and ${self.id} interpolation)",
								},
							},
						},
					},
					"basic_auth": schema.SingleNestedBlock{
						Description: "Basic authentication credentials for destroy request",
						Attributes: map[string]schema.Attribute{
							"username": schema.StringAttribute{
								Optional:    true,
								Sensitive:   true,
								Description: "Basic auth username",
							},
							"password": schema.StringAttribute{
								Optional:    true,
								Sensitive:   true,
								Description: "Basic auth password",
							},
						},
					},
					"retry": schema.SingleNestedBlock{
						Description: "Retry configuration for destroy request",
						Attributes: map[string]schema.Attribute{
							"attempts": schema.Int64Attribute{
								Optional:    true,
								Description: "Maximum number of retry attempts",
							},
							"min_delay_ms": schema.Int64Attribute{
								Optional:    true,
								Description: "Minimum delay between retries in milliseconds",
							},
							"max_delay_ms": schema.Int64Attribute{
								Optional:    true,
								Description: "Maximum delay between retries in milliseconds",
							},
							"backoff": schema.StringAttribute{
								Optional:    true,
								Description: "Backoff strategy: 'fixed', 'linear', or 'exponential'",
							},
							"jitter": schema.BoolAttribute{
								Optional:    true,
								Description: "Add jitter to retry delays",
							},
							"retry_on_status_codes": schema.ListAttribute{
								ElementType: types.Int64Type,
								Optional:    true,
								Description: "HTTP status codes that should trigger a retry",
							},
							"respect_retry_after": schema.BoolAttribute{
								Optional:    true,
								Description: "Respect Retry-After header if present",
							},
						},
					},
					"retry_until": schema.SingleNestedBlock{
						Description: "Conditional retry configuration for destroy request",
						Attributes: map[string]schema.Attribute{
							"status_codes": schema.ListAttribute{
								ElementType: types.Int64Type,
								Optional:    true,
								Description: "Status codes that satisfy the condition",
							},
							"json_path_equals": schema.MapAttribute{
								ElementType: types.StringType,
								Optional:    true,
								Description: "JSON path conditions that must equal specified values",
							},
							"header_equals": schema.MapAttribute{
								ElementType: types.StringType,
								Optional:    true,
								Description: "Header conditions that must equal specified values",
							},
							"body_regex": schema.StringAttribute{
								Optional:    true,
								Description: "Regex pattern that must match the response body",
							},
						},
					},
					"expect": schema.SingleNestedBlock{
						Description: "Response expectations for destroy request",
						Attributes: map[string]schema.Attribute{
							"status_codes": schema.ListAttribute{
								ElementType: types.Int64Type,
								Optional:    true,
								Description: "Expected HTTP status codes",
							},
							"json_path_exists": schema.ListAttribute{
								ElementType: types.StringType,
								Optional:    true,
								Description: "JSON paths that must exist",
							},
							"json_path_equals": schema.MapAttribute{
								ElementType: types.StringType,
								Optional:    true,
								Description: "JSON path conditions that must equal specified values",
							},
							"header_present": schema.ListAttribute{
								ElementType: types.StringType,
								Optional:    true,
								Description: "Headers that must be present",
							},
						},
					},
					"extract": schema.ListNestedBlock{
						Description: "Extract values from destroy response (for condition evaluation only, not persisted)",
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Required:    true,
									Description: "Name of the extracted value",
								},
								"json_path": schema.StringAttribute{
									Optional:    true,
									Description: "JSON path to extract from",
								},
								"header": schema.StringAttribute{
									Optional:    true,
									Description: "Header name to extract from",
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *HttpxRequestResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	config, ok := req.ProviderData.(*ProviderConfig)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			"Expected *ProviderConfig, got something else",
		)
		return
	}

	r.config = config
}

func (r *HttpxRequestResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model HttpxRequestResourceModel

	// Read Terraform configuration into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build request configuration
	headers, err := ConvertTerraformMap(ctx, model.Headers)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Headers", err.Error())
		return
	}

	query, err := ConvertTerraformMap(ctx, model.Query)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Query", err.Error())
		return
	}

	// Build HTTP request
	httpReq, err := BuildRequest(ctx, &RequestConfig{
		Url:              model.Url.ValueString(),
		Method:           model.Method.ValueString(),
		Headers:          headers,
		HeaderBlocks:     model.HeaderBlocks,
		Query:            query,
		Body:             model.Body,
		BodyJson:         model.BodyJson,
		BodyFile:         model.BodyFile,
		BasicAuth:        model.BasicAuth,
		BearerToken:      model.BearerToken,
		ProviderDefaults: r.config,
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to build request", err.Error())
		return
	}

	// Build retry configs
	retryConfig := BuildRetryConfig(ctx, model.Retry)
	retryUntilConfig := BuildRetryUntilConfig(ctx, model.RetryUntil)

	// Handle timeouts if configured
	createCtx := ctx
	if model.Timeouts != nil && !model.Timeouts.Create.IsNull() && !model.Timeouts.Create.IsUnknown() {
		timeoutStr := model.Timeouts.Create.ValueString()
		if timeout, err := time.ParseDuration(timeoutStr); err == nil {
			var cancel context.CancelFunc
			createCtx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}
	}

	// Execute request with retry and conditional retry
	result, err := ExecuteRequestWithRetry(createCtx, httpReq, r.config, retryConfig, retryUntilConfig)
	if err != nil {
		if createCtx.Err() == context.DeadlineExceeded {
			resp.Diagnostics.AddError("Request timeout", fmt.Sprintf("Request exceeded timeout, last error: %s", err.Error()))
		} else {
			resp.Diagnostics.AddError("Request failed", err.Error())
		}
		return
	}

	// Validate expectations
	if model.Expect != nil {
		if err := ValidateExpectations(ctx, result, model.Expect); err != nil {
			resp.Diagnostics.AddError("Expectation validation failed", err.Error())
			return
		}
	}

	// Generate ID (hash of request inputs for stability)
	id := generateResourceID(model)

	// Set computed attributes
	model.Id = types.StringValue(id)
	model.StatusCode = types.Int64Value(result.StatusCode)
	model.LastAttemptCount = types.Int64Value(result.AttemptCount)
	if result.Error != "" {
		model.LastError = types.StringValue(result.Error)
	} else {
		model.LastError = types.StringNull()
	}

	// Set response headers
	responseHeaders := make(map[string]attr.Value)
	for k, v := range result.Headers {
		responseHeaders[k] = types.StringValue(v)
	}
	model.ResponseHeaders = types.MapValueMust(types.StringType, responseHeaders)

	// Set response body (respect store_response_body)
	// Default: true for resources (users may need the body)
	// But if extract blocks are present, default to false to save state space
	storeBody := true
	if !model.StoreResponseBody.IsNull() && !model.StoreResponseBody.IsUnknown() {
		storeBody = model.StoreResponseBody.ValueBool()
	} else if len(model.ExtractBlocks) > 0 {
		// If extract blocks are present and store_response_body not explicitly set,
		// default to false to save state space (user can override)
		storeBody = false
	}

	if storeBody {
		model.ResponseBody = types.StringValue(result.Body)
	} else {
		model.ResponseBody = types.StringNull()
	}

	// Extract values from response
	extractedOutputs, err := ExtractValues(ctx, result, model.ExtractBlocks)
	if err != nil {
		resp.Diagnostics.AddWarning("Extraction warnings", fmt.Sprintf("Some values could not be extracted: %v", err))
	}

	// Convert extracted outputs to Terraform map
	outputsMap := make(map[string]attr.Value)
	for k, v := range extractedOutputs {
		outputsMap[k] = types.StringValue(v)
	}
	model.Outputs = types.MapValueMust(types.StringType, outputsMap)

	// Save state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// generateResourceID generates a stable ID for the resource
func generateResourceID(model HttpxRequestResourceModel) string {
	// Create a hash of key request attributes for stability
	hashInput := fmt.Sprintf("%s|%s|%s",
		model.Url.ValueString(),
		model.Method.ValueString(),
		model.Body.ValueString())

	hash := sha256.Sum256([]byte(hashInput))
	return hex.EncodeToString(hash[:])[:16] // Use first 16 chars
}

func (r *HttpxRequestResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model HttpxRequestResourceModel

	// Read current state
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check read_mode
	readMode := "none"
	if !model.ReadMode.IsNull() && !model.ReadMode.IsUnknown() {
		readMode = model.ReadMode.ValueString()
	}

	if readMode == "none" {
		// No-op: just return current state
		resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
		return
	}

	// readMode == "refresh": re-execute the request
	headers, err := ConvertTerraformMap(ctx, model.Headers)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Headers", err.Error())
		return
	}

	query, err := ConvertTerraformMap(ctx, model.Query)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Query", err.Error())
		return
	}

	// Build and execute request
	httpReq, err := BuildRequest(ctx, &RequestConfig{
		Url:              model.Url.ValueString(),
		Method:           model.Method.ValueString(),
		Headers:          headers,
		HeaderBlocks:     model.HeaderBlocks,
		Query:            query,
		Body:             model.Body,
		BodyJson:         model.BodyJson,
		BodyFile:         model.BodyFile,
		BasicAuth:        model.BasicAuth,
		BearerToken:      model.BearerToken,
		ProviderDefaults: r.config,
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to build request", err.Error())
		return
	}

	// Build retry configs
	retryConfig := BuildRetryConfig(ctx, model.Retry)
	retryUntilConfig := BuildRetryUntilConfig(ctx, model.RetryUntil)

	// Handle timeouts if configured
	updateCtx := ctx
	if model.Timeouts != nil && !model.Timeouts.Update.IsNull() && !model.Timeouts.Update.IsUnknown() {
		timeoutStr := model.Timeouts.Update.ValueString()
		if timeout, err := time.ParseDuration(timeoutStr); err == nil {
			var cancel context.CancelFunc
			updateCtx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}
	}

	// Execute request with retry and conditional retry
	result, err := ExecuteRequestWithRetry(updateCtx, httpReq, r.config, retryConfig, retryUntilConfig)
	if err != nil {
		if updateCtx.Err() == context.DeadlineExceeded {
			resp.Diagnostics.AddError("Request timeout", fmt.Sprintf("Request exceeded timeout, last error: %s", err.Error()))
		} else {
			resp.Diagnostics.AddError("Request failed", err.Error())
		}
		return
	}

	// Update state with fresh response
	model.StatusCode = types.Int64Value(result.StatusCode)
	model.LastAttemptCount = types.Int64Value(result.AttemptCount)
	if result.Error != "" {
		model.LastError = types.StringValue(result.Error)
	} else {
		model.LastError = types.StringNull()
	}

	responseHeaders := make(map[string]attr.Value)
	for k, v := range result.Headers {
		responseHeaders[k] = types.StringValue(v)
	}
	model.ResponseHeaders = types.MapValueMust(types.StringType, responseHeaders)

	// Default: true, but false if extract blocks present (unless explicitly set)
	storeBody := true
	if !model.StoreResponseBody.IsNull() && !model.StoreResponseBody.IsUnknown() {
		storeBody = model.StoreResponseBody.ValueBool()
	} else if len(model.ExtractBlocks) > 0 {
		// If extract blocks are present and store_response_body not explicitly set,
		// default to false to save state space
		storeBody = false
	}

	if storeBody {
		model.ResponseBody = types.StringValue(result.Body)
	}

	// Extract values from response
	extractedOutputs, err := ExtractValues(ctx, result, model.ExtractBlocks)
	if err != nil {
		resp.Diagnostics.AddWarning("Extraction warnings", fmt.Sprintf("Some values could not be extracted: %v", err))
	}

	// Convert extracted outputs to Terraform map
	outputsMap := make(map[string]attr.Value)
	for k, v := range extractedOutputs {
		outputsMap[k] = types.StringValue(v)
	}
	model.Outputs = types.MapValueMust(types.StringType, outputsMap)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *HttpxRequestResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Update is essentially the same as Create - re-execute the request
	var model HttpxRequestResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	headers, err := ConvertTerraformMap(ctx, model.Headers)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Headers", err.Error())
		return
	}

	query, err := ConvertTerraformMap(ctx, model.Query)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Query", err.Error())
		return
	}

	httpReq, err := BuildRequest(ctx, &RequestConfig{
		Url:              model.Url.ValueString(),
		Method:           model.Method.ValueString(),
		Headers:          headers,
		HeaderBlocks:     model.HeaderBlocks,
		Query:            query,
		Body:             model.Body,
		BodyJson:         model.BodyJson,
		BodyFile:         model.BodyFile,
		BasicAuth:        model.BasicAuth,
		BearerToken:      model.BearerToken,
		ProviderDefaults: r.config,
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to build request", err.Error())
		return
	}

	// Build retry configs
	retryConfig := BuildRetryConfig(ctx, model.Retry)
	retryUntilConfig := BuildRetryUntilConfig(ctx, model.RetryUntil)

	// Handle timeouts if configured
	readCtx := ctx
	if model.Timeouts != nil && !model.Timeouts.Read.IsNull() && !model.Timeouts.Read.IsUnknown() {
		timeoutStr := model.Timeouts.Read.ValueString()
		if timeout, err := time.ParseDuration(timeoutStr); err == nil {
			var cancel context.CancelFunc
			readCtx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}
	}

	// Execute request with retry and conditional retry
	result, err := ExecuteRequestWithRetry(readCtx, httpReq, r.config, retryConfig, retryUntilConfig)
	if err != nil {
		resp.Diagnostics.AddError("Request failed", err.Error())
		return
	}

	if model.Expect != nil {
		if err := ValidateExpectations(ctx, result, model.Expect); err != nil {
			resp.Diagnostics.AddError("Expectation validation failed", err.Error())
			return
		}
	}

	// Update computed attributes
	model.StatusCode = types.Int64Value(result.StatusCode)
	model.LastAttemptCount = types.Int64Value(result.AttemptCount)
	if result.Error != "" {
		model.LastError = types.StringValue(result.Error)
	} else {
		model.LastError = types.StringNull()
	}

	responseHeaders := make(map[string]attr.Value)
	for k, v := range result.Headers {
		responseHeaders[k] = types.StringValue(v)
	}
	model.ResponseHeaders = types.MapValueMust(types.StringType, responseHeaders)

	// Default: true, but false if extract blocks present (unless explicitly set)
	storeBody := true
	if !model.StoreResponseBody.IsNull() && !model.StoreResponseBody.IsUnknown() {
		storeBody = model.StoreResponseBody.ValueBool()
	} else if len(model.ExtractBlocks) > 0 {
		// If extract blocks are present and store_response_body not explicitly set,
		// default to false to save state space
		storeBody = false
	}

	if storeBody {
		model.ResponseBody = types.StringValue(result.Body)
	} else {
		model.ResponseBody = types.StringNull()
	}

	// Extract values from response
	extractedOutputs, err := ExtractValues(ctx, result, model.ExtractBlocks)
	if err != nil {
		resp.Diagnostics.AddWarning("Extraction warnings", fmt.Sprintf("Some values could not be extracted: %v", err))
	}

	// Convert extracted outputs to Terraform map
	outputsMap := make(map[string]attr.Value)
	for k, v := range extractedOutputs {
		outputsMap[k] = types.StringValue(v)
	}
	model.Outputs = types.MapValueMust(types.StringType, outputsMap)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *HttpxRequestResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model HttpxRequestResourceModel

	// Read current state
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If no on_destroy configured, just remove from state (no-op delete)
	if model.OnDestroy == nil {
		tflog.Info(ctx, "Delete method called - no on_destroy block configured, removing from state")
		return
	}

	tflog.Info(ctx, "Delete method called - executing on_destroy request")

	// Build interpolation context from current state
	interpolCtx, err := BuildInterpolationContextFromState(ctx, &model)
	if err != nil {
		resp.Diagnostics.AddError("Failed to build interpolation context", err.Error())
		return
	}

	// Expand templates in on_destroy config
	destroyConfig := model.OnDestroy

	// Interpolate URL
	if !destroyConfig.Url.IsNull() {
		expandedURL, err := InterpolateString(ctx, destroyConfig.Url.ValueString(), interpolCtx)
		if err != nil {
			resp.Diagnostics.AddError("Failed to interpolate destroy URL", err.Error())
			return
		}
		destroyConfig.Url = types.StringValue(expandedURL)
	}

	// Interpolate headers (map)
	if !destroyConfig.Headers.IsNull() {
		headersMap, err := ConvertTerraformMap(ctx, destroyConfig.Headers)
		if err != nil {
			resp.Diagnostics.AddError("Invalid destroy headers", err.Error())
			return
		}
		expandedHeaders, err := InterpolateMap(ctx, headersMap, interpolCtx)
		if err != nil {
			resp.Diagnostics.AddError("Failed to interpolate destroy headers", err.Error())
			return
		}
		// Convert back to Terraform map
		headersAttrMap := make(map[string]attr.Value)
		for k, v := range expandedHeaders {
			headersAttrMap[k] = types.StringValue(v)
		}
		destroyConfig.Headers = types.MapValueMust(types.StringType, headersAttrMap)
	}

	// Interpolate header blocks (repeated)
	if len(destroyConfig.HeaderBlocks) > 0 {
		expandedBlocks, err := InterpolateHeaderBlocks(ctx, destroyConfig.HeaderBlocks, interpolCtx)
		if err != nil {
			resp.Diagnostics.AddError("Failed to interpolate destroy headers", err.Error())
			return
		}
		destroyConfig.HeaderBlocks = expandedBlocks
	}

	// Interpolate query parameters
	if !destroyConfig.Query.IsNull() {
		queryMap, err := ConvertTerraformMap(ctx, destroyConfig.Query)
		if err != nil {
			resp.Diagnostics.AddError("Invalid destroy query", err.Error())
			return
		}
		expandedQuery, err := InterpolateMap(ctx, queryMap, interpolCtx)
		if err != nil {
			resp.Diagnostics.AddError("Failed to interpolate destroy query", err.Error())
			return
		}
		queryAttrMap := make(map[string]attr.Value)
		for k, v := range expandedQuery {
			queryAttrMap[k] = types.StringValue(v)
		}
		destroyConfig.Query = types.MapValueMust(types.StringType, queryAttrMap)
	}

	// Interpolate body fields
	if !destroyConfig.Body.IsNull() {
		expandedBody, err := InterpolateString(ctx, destroyConfig.Body.ValueString(), interpolCtx)
		if err != nil {
			resp.Diagnostics.AddError("Failed to interpolate destroy body", err.Error())
			return
		}
		destroyConfig.Body = types.StringValue(expandedBody)
	}

	if !destroyConfig.BodyJson.IsNull() {
		expandedBodyJson, err := InterpolateString(ctx, destroyConfig.BodyJson.ValueString(), interpolCtx)
		if err != nil {
			resp.Diagnostics.AddError("Failed to interpolate destroy body_json", err.Error())
			return
		}
		destroyConfig.BodyJson = types.StringValue(expandedBodyJson)
	}

	// Build HTTP request from destroy config
	headers, err := ConvertTerraformMap(ctx, destroyConfig.Headers)
	if err != nil {
		resp.Diagnostics.AddError("Invalid destroy headers", err.Error())
		return
	}

	query, err := ConvertTerraformMap(ctx, destroyConfig.Query)
	if err != nil {
		resp.Diagnostics.AddError("Invalid destroy query", err.Error())
		return
	}

	httpReq, err := BuildRequest(ctx, &RequestConfig{
		Url:              destroyConfig.Url.ValueString(),
		Method:           destroyConfig.Method.ValueString(),
		Headers:          headers,
		HeaderBlocks:     destroyConfig.HeaderBlocks,
		Query:            query,
		Body:             destroyConfig.Body,
		BodyJson:         destroyConfig.BodyJson,
		BodyFile:         destroyConfig.BodyFile,
		BasicAuth:        destroyConfig.BasicAuth,
		BearerToken:      destroyConfig.BearerToken,
		ProviderDefaults: r.config,
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to build destroy request", err.Error())
		return
	}

	// Build retry configs
	retryConfig := BuildRetryConfig(ctx, destroyConfig.Retry)
	retryUntilConfig := BuildRetryUntilConfig(ctx, destroyConfig.RetryUntil)

	// Execute request with retry logic
	result, err := ExecuteRequestWithRetry(ctx, httpReq, r.config, retryConfig, retryUntilConfig)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Destroy request failed: %s", err.Error()))
		resp.Diagnostics.AddError("Destroy request failed", err.Error())
		// Keep state on destroy failure so Terraform can retry
		return
	}

	// Validate expectations
	if destroyConfig.Expect != nil {
		if err := ValidateExpectations(ctx, result, destroyConfig.Expect); err != nil {
			tflog.Error(ctx, fmt.Sprintf("Destroy expectation validation failed: %s", err.Error()))
			resp.Diagnostics.AddError("Destroy expectation validation failed", err.Error())
			// Keep state on expectation failure
			return
		}
	}

	// Log successful destroy execution
	tflog.Info(ctx, fmt.Sprintf("Destroy request succeeded with status code %d", result.StatusCode))

	// Successfully removed - state will be cleared by Terraform framework
}
