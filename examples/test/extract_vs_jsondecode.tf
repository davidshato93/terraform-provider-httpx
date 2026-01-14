# Comparison: extract vs jsondecode()
# This file demonstrates when to use each approach

# ============================================
# APPROACH 1: Using jsondecode() (your example)
# ============================================
# Pros: Simple, works if you're storing response body
# Cons: Requires storing full response body, only works for JSON

resource "httpx_request" "with_jsondecode" {
  url    = "https://httpbin.org/json"
  method = "GET"

  # Must store response body to use jsondecode()
  store_response_body = true

  expect {
    status_codes = [200]
  }
}

# Extract value using jsondecode() - works but requires response_body
output "jsondecode_example" {
  value = jsondecode(httpx_request.with_jsondecode.response_body).slideshow.title
}

# ============================================
# APPROACH 2: Using extract (recommended)
# ============================================
# Pros: Don't need to store body, works for headers too, cleaner syntax
# Cons: Slightly more configuration

resource "httpx_request" "with_extract" {
  url    = "https://httpbin.org/json"
  method = "GET"

  # Don't need to store response body!
  store_response_body = false  # Saves state space

  extract {
    name      = "title"
    json_path = "slideshow.title"
  }

  extract {
    name      = "author"
    json_path = "slideshow.author"
  }

  extract {
    name   = "content_type"
    header = "Content-Type"  # Can't do this with jsondecode()!
  }

  expect {
    status_codes = [200]
  }
}

# Use extracted values - cleaner and doesn't require response_body
output "extract_example" {
  value = {
    title        = httpx_request.with_extract.outputs["title"]
    author       = httpx_request.with_extract.outputs["author"]
    content_type = httpx_request.with_extract.outputs["content_type"]
  }
}

# ============================================
# REAL-WORLD USE CASE: Large API Response
# ============================================
# Imagine an API that returns a 10MB JSON response with thousands of fields
# You only need 2-3 values

resource "httpx_request" "large_api_response" {
  url    = "https://httpbin.org/json"
  method = "GET"

  # With extract: Don't store the large body, just extract what you need
  store_response_body = false  # Saves 10MB from state!

  extract {
    name      = "important_id"
    json_path = "slideshow.title"  # Just get what you need
  }

  extract {
    name   = "request_id"
    header = "X-Request-Id"
  }
}

# Use the extracted values without storing 10MB in state
resource "httpx_request" "downstream" {
  url    = "https://httpbin.org/get"
  method = "GET"

  query = {
    id = httpx_request.large_api_response.outputs["important_id"]
  }

  headers = {
    "X-Request-Id" = httpx_request.large_api_response.outputs["request_id"]
  }
}

# ============================================
# SUMMARY
# ============================================
# Use jsondecode() when:
# - Response is small and you're already storing it
# - You only need JSON values (not headers)
# - You're comfortable with Terraform JSON parsing
#
# Use extract when:
# - Response is large (save state space)
# - You need header values
# - You want cleaner, more declarative syntax
# - You're extracting multiple values
# - Response contains sensitive data (extract only what you need)

