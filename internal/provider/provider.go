package provider

import (
	"context"

	"github.com/davidshato/terraform-provider-httpx/internal/config"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ provider.Provider = &HttpxProvider{}

type HttpxProvider struct {
	version string
}

type HttpxProviderModel struct {
	DefaultHeaders       map[string]string `tfsdk:"default_headers"`
	BasicAuth            *BasicAuthModel   `tfsdk:"basic_auth"`
	BearerToken          *string           `tfsdk:"bearer_token"`
	TimeoutMs            *int64            `tfsdk:"timeout_ms"`
	InsecureSkipVerify   *bool             `tfsdk:"insecure_skip_verify"`
	ProxyUrl             *string           `tfsdk:"proxy_url"`
	CaCertPem            *string           `tfsdk:"ca_cert_pem"`
	ClientCertPem        *string           `tfsdk:"client_cert_pem"`
	ClientKeyPem         *string           `tfsdk:"client_key_pem"`
	RedactHeaders        []string          `tfsdk:"redact_headers"`
	MaxResponseBodyBytes *int64            `tfsdk:"max_response_body_bytes"`
	Debug                *bool             `tfsdk:"debug"`
}

type BasicAuthModel struct {
	Username string `tfsdk:"username"`
	Password string `tfsdk:"password"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &HttpxProvider{
			version: version,
		}
	}
}

func (p *HttpxProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "httpx"
	resp.Version = p.version
}

func (p *HttpxProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"default_headers": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Optional defaults applied to resources unless overridden",
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
			"ca_cert_pem": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "CA certificate in PEM format",
			},
			"client_cert_pem": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Client certificate in PEM format",
			},
			"client_key_pem": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Client key in PEM format",
			},
			"redact_headers": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Headers to redact in logs and diagnostics",
			},
			"max_response_body_bytes": schema.Int64Attribute{
				Optional:    true,
				Description: "Maximum response body size in bytes",
			},
			"debug": schema.BoolAttribute{
				Optional:    true,
				Description: "Enable debug logging",
			},
		},
		Blocks: map[string]schema.Block{
			"basic_auth": schema.SingleNestedBlock{
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
				Description: "Basic authentication credentials",
			},
		},
		Description: "Provider for executing HTTP requests with retry logic and conditional polling",
	}
}

func (p *HttpxProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config HttpxProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set defaults
	redactHeaders := config.RedactHeaders
	if len(redactHeaders) == 0 {
		redactHeaders = []string{"Authorization", "Proxy-Authorization", "X-Api-Key"}
	}

	maxResponseBodyBytes := int64(1048576) // 1MB default
	if config.MaxResponseBodyBytes != nil {
		maxResponseBodyBytes = *config.MaxResponseBodyBytes
	}

	timeoutMs := int64(30000) // 30 seconds default
	if config.TimeoutMs != nil {
		timeoutMs = *config.TimeoutMs
	}

	insecureSkipVerify := false
	if config.InsecureSkipVerify != nil {
		insecureSkipVerify = *config.InsecureSkipVerify
	}

	// Create provider configuration
	var basicAuthModel *BasicAuthModel
	if config.BasicAuth != nil {
		basicAuthModel = &BasicAuthModel{
			Username: config.BasicAuth.Username,
			Password: config.BasicAuth.Password,
		}
	}

	providerConfig := &ProviderConfig{
		DefaultHeaders:       config.DefaultHeaders,
		BasicAuth:            basicAuthModel,
		BearerToken:          config.BearerToken,
		TimeoutMs:            timeoutMs,
		InsecureSkipVerify:   insecureSkipVerify,
		ProxyUrl:             config.ProxyUrl,
		CaCertPem:            config.CaCertPem,
		ClientCertPem:        config.ClientCertPem,
		ClientKeyPem:         config.ClientKeyPem,
		RedactHeaders:        redactHeaders,
		MaxResponseBodyBytes: maxResponseBodyBytes,
		Debug:                config.Debug != nil && *config.Debug,
	}

	// Enable debug logging if requested
	if providerConfig.Debug {
		tflog.SetField(ctx, "httpx_debug", true)
	}

	resp.ResourceData = providerConfig
	resp.DataSourceData = providerConfig

	tflog.Info(ctx, "Provider configured successfully")
}

func (p *HttpxProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewHttpxRequestResource,
	}
}

func (p *HttpxProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewHttpxRequestDataSource,
	}
}

// ProviderConfig wraps config.ProviderConfig for provider-specific use
// This allows us to keep the config package independent
//
//nolint:revive // ProviderConfig is the correct name for Terraform provider configuration
type ProviderConfig struct {
	DefaultHeaders       map[string]string
	BasicAuth            *BasicAuthModel
	BearerToken          *string
	TimeoutMs            int64
	InsecureSkipVerify   bool
	ProxyUrl             *string
	CaCertPem            *string
	ClientCertPem        *string
	ClientKeyPem         *string
	RedactHeaders        []string
	MaxResponseBodyBytes int64
	Debug                bool
}

// ToConfigProviderConfig converts ProviderConfig to config.ProviderConfig
func (p *ProviderConfig) ToConfigProviderConfig() *config.ProviderConfig {
	var basicAuth *config.BasicAuthModel
	if p.BasicAuth != nil {
		basicAuth = &config.BasicAuthModel{
			Username: p.BasicAuth.Username,
			Password: p.BasicAuth.Password,
		}
	}

	return &config.ProviderConfig{
		DefaultHeaders:       p.DefaultHeaders,
		BasicAuth:            basicAuth,
		BearerToken:          p.BearerToken,
		TimeoutMs:            p.TimeoutMs,
		InsecureSkipVerify:   p.InsecureSkipVerify,
		ProxyUrl:             p.ProxyUrl,
		CaCertPem:            p.CaCertPem,
		ClientCertPem:        p.ClientCertPem,
		ClientKeyPem:         p.ClientKeyPem,
		RedactHeaders:        p.RedactHeaders,
		MaxResponseBodyBytes: p.MaxResponseBodyBytes,
		Debug:                p.Debug,
	}
}
