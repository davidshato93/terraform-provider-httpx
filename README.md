# Terraform Provider: httpx

A Terraform provider for executing HTTP requests with retry logic, conditional polling, resource cleanup (on_destroy), and safe secret handling.

## Features

- Support for all HTTP methods (GET, POST, PUT, PATCH, DELETE, etc.)
- Configurable retry and backoff strategies
- Conditional retry (poll-until) based on response properties
- Flexible request headers (including repeated headers)
- Strong secret hygiene (avoids leaking tokens into state/logs)
- JSON path extraction for downstream resources
- Response validation with expectations
- **Resource cleanup with `on_destroy` block** (execute HTTP requests when resource is deleted, with template interpolation)

## Provider Configuration

```hcl
provider "httpx" {
  # Optional defaults applied to resources unless overridden
  default_headers = {
    "User-Agent" = "terraform-httpx/1.0"
  }

  # Auth helpers (all sensitive)
  basic_auth {
    username = var.user
    password = var.pass
  }

  # Or:
  # bearer_token = var.token

  # Transport
  timeout_ms           = 30000
  insecure_skip_verify = false
  proxy_url            = null

  # TLS customization
  ca_cert_pem       = null
  client_cert_pem   = null
  client_key_pem    = null

  # Safety
  redact_headers         = ["Authorization", "Proxy-Authorization", "X-Api-Key"]
  max_response_body_bytes = 1048576
  debug                  = false
}
```

## Resource: httpx_request

### Basic Example

```hcl
resource "httpx_request" "example" {
  url    = "https://api.example.com/v1/items"
  method = "POST"

  headers = {
    "Content-Type" = "application/json"
  }

  body_json = {
    name = "example"
  }

  expect {
    status_codes = [200, 201]
  }
}
```

### Advanced Example with Conditional Retry

```hcl
resource "httpx_request" "attach" {
  url      = "${local.environment_api_url}/attach"
  method   = "POST"
  headers  = { Authorization = "Basic ${local.cp_jenkins_credentials}" }
  
  body_json = {
    clusterName = var.environment_id
    accountId   = local.account_id
    licenseId   = local.license_id
  }

  retry {
    attempts     = 20
    min_delay_ms = 250
    max_delay_ms = 5000
    backoff      = "exponential"
    jitter       = true
    retry_on_status_codes = [408, 429, 500, 502, 503, 504]
  }

  retry_until {
    status_codes = [200]
    json_path_equals = {
      "data.isAttached" = true
    }
  }

  expect {
    status_codes = [200, 201]
  }

  extract {
    name      = "env_id"
    json_path = "data.environmentId"
  }

  timeouts {
    create = "10m"
  }
}
```

## Resource Schema

### Required Arguments

- `url` (string) - The URL to make the request to
- `method` (string) - HTTP method (GET, POST, PUT, PATCH, DELETE, etc.)

### Optional Arguments

- `headers` (map(string)) - Request headers as a map
- `header` (block) - Repeated header blocks for multiple values with the same name
- `query` (map(string)) - Query parameters
- `body` (string) - Raw request body
- `body_json` (any) - JSON-encodable object (mutually exclusive with `body` and `body_file`)
- `body_file` (string) - Path to file to read and send (mutually exclusive with `body` and `body_json`)
- `basic_auth` (block) - Basic authentication credentials
- `bearer_token` (string, sensitive) - Bearer token for authentication
- `timeout_ms` (number) - Request timeout in milliseconds
- `insecure_skip_verify` (bool) - Skip TLS certificate verification
- `proxy_url` (string) - Proxy URL
- `retry` (block) - Retry configuration
- `retry_until` (block) - Conditional retry (poll-until) configuration
- `expect` (block) - Response expectations/validation
- `extract` (block) - Extract values from response
- `response_sensitive` (bool) - Mark response body as sensitive
- `store_response_body` (bool) - Whether to store response body in state
- `read_mode` (string) - Read behavior: "none" or "refresh"
- `timeouts` (block) - Timeout configuration
- `on_destroy` (block) - Execute HTTP request when resource is destroyed (template interpolation with `${self.outputs.KEY}` and `${self.id}` supported)

### Computed Attributes

- `status_code` (number) - HTTP status code
- `response_headers` (map(string)) - Response headers
- `response_body` (string, optionally sensitive) - Response body
- `outputs` (map(string)) - Extracted values from `extract` blocks
- `last_attempt_count` (number) - Number of attempts made
- `last_error` (string) - Last error message (redacted)
- `id` (string) - Resource identifier

## Data Source: httpx_request

Same schema as the resource, but read-only. Defaults `store_response_body = false`.

```hcl
data "httpx_request" "status" {
  url    = "https://api.example.com/v1/status"
  method = "GET"

  expect {
    status_codes = [200]
  }
}
```

## Documentation

### For Users

- [Provider Configuration](./README.md#provider-configuration) - Configure the provider
- [Resource Documentation](./docs/resources/request.md) - Full schema for `httpx_request` resource
- [Data Source Documentation](./docs/data-sources/request.md) - Full schema for data source
- [Examples](./examples/test/README.md) - Real-world usage examples
- [Best Practices](./docs/GOTCHAS.md) - Common pitfalls and recommendations

### For Developers

- [Implementation Details](./docs/ON_DESTROY_IMPLEMENTATION.md) - How the on_destroy feature works
- [Testing Guide](./TESTING.md) - How to test the provider locally
- [Verifying on_destroy](./examples/test/VERIFY_DESTROY.md) - How to verify destroy requests
- [Release Process](./docs/RELEASE.md) - How to cut a release
- [Changelog](./CHANGELOG.md) - Version history

## Development

### Building

```bash
go build -o terraform-provider-httpx
```

### Testing

```bash
go test ./...
```

### Local Testing with Terraform

```bash
# Set up provider override
export TF_CLI_CONFIG_FILE=examples/test/.terraformrc

# Build provider
go build -o terraform-provider-httpx

# Run examples
cd examples/test
terraform plan
terraform apply
terraform destroy
```

For detailed local testing instructions, see [TESTING.md](./TESTING.md).

### Verifying on_destroy Feature

```bash
cd examples/test
./verify_destroy.sh
```

See [VERIFY_DESTROY.md](./examples/test/VERIFY_DESTROY.md) for detailed verification procedures.

### CI/CD

The project includes GitHub Actions workflows for:
- Linting (`golangci-lint`)
- Testing (`go test`)
- Building binaries for multiple platforms
- Automated releases with GPG signing

## Release Process

See [`docs/RELEASE.md`](./docs/RELEASE.md) for the release process and versioning policy.

## License

Internal use only.