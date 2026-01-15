package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// RequestConfigModel represents the shared request configuration
// Used by both root request and on_destroy block
type RequestConfigModel struct {
	Url                types.String `tfsdk:"url"`
	Method             types.String `tfsdk:"method"`
	Headers            types.Map    `tfsdk:"headers"`
	Query              types.Map    `tfsdk:"query"`
	Body               types.String `tfsdk:"body"`
	BodyJson           types.String `tfsdk:"body_json"`
	BodyFile           types.String `tfsdk:"body_file"`
	BearerToken        types.String `tfsdk:"bearer_token"`
	TimeoutMs          types.Int64  `tfsdk:"timeout_ms"`
	InsecureSkipVerify types.Bool   `tfsdk:"insecure_skip_verify"`
	ProxyUrl           types.String `tfsdk:"proxy_url"`
	ResponseSensitive  types.Bool   `tfsdk:"response_sensitive"`
	StoreResponseBody  types.Bool   `tfsdk:"store_response_body"`

	// Blocks
	HeaderBlocks  []HeaderBlockModel       `tfsdk:"header"`
	BasicAuth     *ResourceBasicAuthModel  `tfsdk:"basic_auth"`
	Retry         *RetryModel              `tfsdk:"retry"`
	RetryUntil    *RetryUntilModel         `tfsdk:"retry_until"`
	Expect        *ExpectModel             `tfsdk:"expect"`
	ExtractBlocks []ExtractBlockModel      `tfsdk:"extract"`
}

// HttpxRequestResourceModel represents the resource state
type HttpxRequestResourceModel struct {
	Id                types.String `tfsdk:"id"`
	ReadMode          types.String `tfsdk:"read_mode"`
	StatusCode        types.Int64  `tfsdk:"status_code"`
	ResponseHeaders   types.Map    `tfsdk:"response_headers"`
	ResponseBody      types.String `tfsdk:"response_body"`
	Outputs           types.Map    `tfsdk:"outputs"`
	LastAttemptCount  types.Int64  `tfsdk:"last_attempt_count"`
	LastError         types.String `tfsdk:"last_error"`

	// Root request configuration (flattened from RequestConfigModel)
	Url                types.String `tfsdk:"url"`
	Method             types.String `tfsdk:"method"`
	Headers            types.Map    `tfsdk:"headers"`
	Query              types.Map    `tfsdk:"query"`
	Body               types.String `tfsdk:"body"`
	BodyJson           types.String `tfsdk:"body_json"`
	BodyFile           types.String `tfsdk:"body_file"`
	BearerToken        types.String `tfsdk:"bearer_token"`
	TimeoutMs          types.Int64  `tfsdk:"timeout_ms"`
	InsecureSkipVerify types.Bool   `tfsdk:"insecure_skip_verify"`
	ProxyUrl           types.String `tfsdk:"proxy_url"`
	ResponseSensitive  types.Bool   `tfsdk:"response_sensitive"`
	StoreResponseBody  types.Bool   `tfsdk:"store_response_body"`

	// Root request blocks
	HeaderBlocks  []HeaderBlockModel       `tfsdk:"header"`
	BasicAuth     *ResourceBasicAuthModel  `tfsdk:"basic_auth"`
	Retry         *RetryModel              `tfsdk:"retry"`
	RetryUntil    *RetryUntilModel         `tfsdk:"retry_until"`
	Expect        *ExpectModel             `tfsdk:"expect"`
	ExtractBlocks []ExtractBlockModel      `tfsdk:"extract"`

	// Destroy configuration
	OnDestroy *RequestConfigModel `tfsdk:"on_destroy"`
	Timeouts  *TimeoutsModel      `tfsdk:"timeouts"`
}

// HeaderBlockModel represents a repeated header block
type HeaderBlockModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

// ResourceBasicAuthModel represents basic auth credentials (for resource models)
type ResourceBasicAuthModel struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

// RetryModel represents retry configuration
type RetryModel struct {
	Attempts            types.Int64   `tfsdk:"attempts"`
	MinDelayMs          types.Int64   `tfsdk:"min_delay_ms"`
	MaxDelayMs          types.Int64   `tfsdk:"max_delay_ms"`
	Backoff             types.String   `tfsdk:"backoff"`
	Jitter              types.Bool    `tfsdk:"jitter"`
	RetryOnStatusCodes  types.List    `tfsdk:"retry_on_status_codes"`
	RespectRetryAfter   types.Bool    `tfsdk:"respect_retry_after"`
}

// RetryUntilModel represents conditional retry configuration
type RetryUntilModel struct {
	StatusCodes     types.List    `tfsdk:"status_codes"`
	JsonPathEquals  types.Map     `tfsdk:"json_path_equals"`
	HeaderEquals    types.Map     `tfsdk:"header_equals"`
	BodyRegex       types.String  `tfsdk:"body_regex"`
}

// ExpectModel represents response expectations
type ExpectModel struct {
	StatusCodes     types.List    `tfsdk:"status_codes"`
	JsonPathExists  types.List    `tfsdk:"json_path_exists"`
	JsonPathEquals  types.Map     `tfsdk:"json_path_equals"`
	HeaderPresent   types.List    `tfsdk:"header_present"`
}

// ExtractBlockModel represents an extract block
type ExtractBlockModel struct {
	Name     types.String `tfsdk:"name"`
	JsonPath types.String `tfsdk:"json_path"`
	Header   types.String `tfsdk:"header"`
}

// TimeoutsModel represents timeout configuration
type TimeoutsModel struct {
	Create types.String `tfsdk:"create"`
	Read   types.String `tfsdk:"read"`
	Update types.String `tfsdk:"update"`
	Delete types.String `tfsdk:"delete"`
}

