package provider

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &HttpxRequestDataSource{}
var _ datasource.DataSourceWithConfigure = &HttpxRequestDataSource{}

type HttpxRequestDataSource struct {
	config *ProviderConfig
}

func NewHttpxRequestDataSource() datasource.DataSource {
	return &HttpxRequestDataSource{}
}

func (d *HttpxRequestDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_request"
}

func (d *HttpxRequestDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	// Same schema as resource, but read-only
	resp.Schema = schema.Schema{
		Description: "Data source for executing HTTP requests (read-only)",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Data source identifier",
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
				Description: "Whether to store response body in state (defaults to false for data sources)",
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
		},
	}
}

func (d *HttpxRequestDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	config, ok := req.ProviderData.(*ProviderConfig)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			"Expected *ProviderConfig, got something else",
		)
		return
	}

	d.config = config
}

func (d *HttpxRequestDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model HttpxRequestDataSourceModel

	// Read Terraform configuration into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
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
		ProviderDefaults: d.config,
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to build request", err.Error())
		return
	}

	// Build retry configs
	retryConfig := BuildRetryConfig(ctx, model.Retry)
	retryUntilConfig := BuildRetryUntilConfig(ctx, model.RetryUntil)

	// Execute request with retry and conditional retry
	result, err := ExecuteRequestWithRetry(ctx, httpReq, d.config, retryConfig, retryUntilConfig)
	if err != nil {
		resp.Diagnostics.AddError("Request failed", err.Error())
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
	id := generateDataSourceID(model)

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

	// Set response body (default to false for data sources to avoid polluting state)
	storeBody := false
	if !model.StoreResponseBody.IsNull() && !model.StoreResponseBody.IsUnknown() {
		storeBody = model.StoreResponseBody.ValueBool()
	} else if len(model.ExtractBlocks) == 0 {
		// If no extract blocks and not explicitly set, default to false for data sources
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

// generateDataSourceID generates a stable ID for the data source
func generateDataSourceID(model HttpxRequestDataSourceModel) string {
	// Create a hash of key request attributes for stability
	hashInput := fmt.Sprintf("%s|%s|%s",
		model.Url.ValueString(),
		model.Method.ValueString(),
		model.Body.ValueString())

	hash := sha256.Sum256([]byte(hashInput))
	return hex.EncodeToString(hash[:])[:16] // Use first 16 chars
}

