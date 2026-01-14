package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// HttpxRequestDataSourceModel represents the data source state
type HttpxRequestDataSourceModel struct {
	Id                  types.String `tfsdk:"id"`
	Url                 types.String `tfsdk:"url"`
	Method              types.String `tfsdk:"method"`
	Headers             types.Map    `tfsdk:"headers"`
	Query               types.Map    `tfsdk:"query"`
	Body                types.String `tfsdk:"body"`
	BodyJson            types.String `tfsdk:"body_json"`
	BodyFile            types.String `tfsdk:"body_file"`
	BearerToken         types.String `tfsdk:"bearer_token"`
	TimeoutMs           types.Int64  `tfsdk:"timeout_ms"`
	InsecureSkipVerify  types.Bool   `tfsdk:"insecure_skip_verify"`
	ProxyUrl            types.String `tfsdk:"proxy_url"`
	ResponseSensitive   types.Bool   `tfsdk:"response_sensitive"`
	StoreResponseBody   types.Bool   `tfsdk:"store_response_body"`
	StatusCode          types.Int64  `tfsdk:"status_code"`
	ResponseHeaders     types.Map    `tfsdk:"response_headers"`
	ResponseBody        types.String `tfsdk:"response_body"`
	Outputs             types.Map    `tfsdk:"outputs"`
	LastAttemptCount    types.Int64  `tfsdk:"last_attempt_count"`
	LastError           types.String `tfsdk:"last_error"`

	// Blocks
	HeaderBlocks        []HeaderBlockModel        `tfsdk:"header"`
	BasicAuth           *ResourceBasicAuthModel    `tfsdk:"basic_auth"`
	Retry               *RetryModel                `tfsdk:"retry"`
	RetryUntil          *RetryUntilModel           `tfsdk:"retry_until"`
	Expect              *ExpectModel                `tfsdk:"expect"`
	ExtractBlocks       []ExtractBlockModel         `tfsdk:"extract"`
}

