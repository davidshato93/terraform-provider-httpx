# Example: Using httpx_request data source
# Data sources are read-only and don't create state resources
# They're useful for fetching data during plan/apply without managing lifecycle

# Simple GET request
data "httpx_request" "api_status" {
  url    = "https://httpbin.org/status/200"
  method = "GET"

  expect {
    status_codes = [200]
  }
}

output "api_status_code" {
  value = data.httpx_request.api_status.status_code
}

# GET request with extraction (doesn't store body by default)
data "httpx_request" "json_data" {
  url    = "https://httpbin.org/json"
  method = "GET"

  extract {
    name     = "slideshow_title"
    json_path = "slideshow.title"
  }

  extract {
    name     = "author"
    json_path = "slideshow.author"
  }
}

output "extracted_title" {
  value = data.httpx_request.json_data.outputs["slideshow_title"]
}

output "extracted_author" {
  value = data.httpx_request.json_data.outputs["author"]
}

# Data source with retry
data "httpx_request" "with_retry" {
  url    = "https://httpbin.org/status/500"
  method = "GET"

  retry {
    attempts         = 3
    min_delay_ms     = 1000
    max_delay_ms     = 5000
    backoff          = "exponential"
    retry_on_status_codes = [500]
  }

  expect {
    status_codes = [500] # We expect 500, but want to see retry behavior
  }
}

output "retry_attempts" {
  value = data.httpx_request.with_retry.last_attempt_count
}

# Data source with conditional retry (polling)
# This will poll until the condition is met or timeout
data "httpx_request" "polling" {
  url    = "https://httpbin.org/status/200"
  method = "GET"

  retry_until {
    status_codes = [200]
  }

  retry {
    attempts     = 10
    min_delay_ms = 2000
    max_delay_ms = 5000
    backoff      = "linear"
  }
}

output "polling_status" {
  value = data.httpx_request.polling.status_code
}

# Data source with header extraction
data "httpx_request" "headers" {
  url    = "https://httpbin.org/headers"
  method = "GET"

  extract {
    name   = "content_type"
    header = "Content-Type"
  }
}

output "content_type_header" {
  value = data.httpx_request.headers.outputs["content_type"]
}

