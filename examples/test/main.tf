# Configure the provider
provider "httpx" {
  default_headers = {
    "User-Agent" = "terraform-httpx-test/1.0"
  }
  
  timeout_ms           = 30000
  insecure_skip_verify = false
  max_response_body_bytes = 1048576
}

# Test 1: Simple GET request
resource "httpx_request" "test_get" {
  url    = "https://httpbin.org/get"
  method = "GET"

  query = {
    test = "value"
    foo  = "bar"
  }

  expect {
    status_codes = [200]
    header_present = ["Content-Type"]
  }
}

# Test 2: POST request with JSON body
resource "httpx_request" "test_post" {
  url    = "https://httpbin.org/post"
  method = "POST"

  headers = {
    "Content-Type" = "application/json"
  }

  body_json = jsonencode({
    name    = "test"
    message = "Hello from Terraform!"
    number  = 42
  })

  expect {
    status_codes = [200]
  }

  depends_on = [httpx_request.test_get]
}

# Test 3: GET request with basic auth (using httpbin.org which accepts any credentials)
resource "httpx_request" "test_auth" {
  url    = "https://httpbin.org/basic-auth/user/pass"
  method = "GET"

  basic_auth {
    username = "user"
    password = "pass"
  }

  expect {
    status_codes = [200]
  }

  depends_on = [httpx_request.test_post]
}

# Outputs to see the results
output "get_status_code" {
  value     = httpx_request.test_get.status_code
  sensitive = false
}

output "get_response_body" {
  value     = httpx_request.test_get.response_body
  sensitive = false
}

output "post_status_code" {
  value     = httpx_request.test_post.status_code
  sensitive = false
}

output "auth_status_code" {
  value     = httpx_request.test_auth.status_code
  sensitive = false
}

