---
page_title: "Provider: httpx"
subcategory: ""
description: |-
  The httpx provider allows Terraform to manage HTTP requests with retry logic, conditional polling, and safe secret handling.
---

# httpx Provider

The httpx provider enables Terraform to execute HTTP requests with advanced features including:

- Support for all HTTP methods (GET, POST, PUT, PATCH, DELETE, etc.)
- Configurable retry and backoff strategies (fixed, linear, exponential, jitter)
- Conditional retry (poll-until) based on response properties
- Flexible request headers (including repeated headers)
- Strong secret hygiene (avoids leaking tokens into state/logs)
- JSON path extraction for downstream resources
- Response validation with expectations
- TLS configuration and custom CA certificates
- Proxy support
- Basic authentication and bearer token support

## Example Usage

```hcl
terraform {
  required_providers {
    httpx = {
      source  = "davidshato93/httpx"
      version = "~> 1.0"
    }
  }
}

provider "httpx" {
  # Optional provider-level defaults
  default_headers = {
    "User-Agent" = "terraform-httpx/1.0"
  }
  
  # Optional timeout (default: 30s)
  timeout_ms = 30000
}

# Simple GET request
resource "httpx_request" "example" {
  url    = "https://api.example.com/endpoint"
  method = "GET"
}
```

## Configuration

The provider supports the following top-level arguments:

- **`default_headers`** (Optional) - A map of HTTP headers to include in every request
- **`bearer_token`** (Optional, Sensitive) - Default bearer token for authentication
- **`basic_auth`** (Optional) - Default basic authentication (username and password)
- **`timeout_ms`** (Optional) - Default request timeout in milliseconds (default: 30000)
- **`insecure_skip_verify`** (Optional) - Skip TLS certificate verification (default: false)
- **`ca_cert_pem`** (Optional) - Custom CA certificate in PEM format
- **`client_cert_pem`** (Optional) - Client certificate in PEM format
- **`client_key_pem`** (Optional) - Client private key in PEM format
- **`proxy_url`** (Optional) - HTTP proxy URL
- **`max_response_body_bytes`** (Optional) - Maximum response body size in bytes (default: 1MB)

## Resources and Data Sources

Please refer to the documentation for the individual resources and data sources:

- [httpx_request (resource)](resources/httpx_request.md)
- [httpx_request (data source)](data-sources/httpx_request.md)

## Documentation

For more information, visit the [GitHub repository](https://github.com/davidshato93/terraform-provider-httpx).


