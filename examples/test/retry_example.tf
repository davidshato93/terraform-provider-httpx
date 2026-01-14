# Example: Testing retry functionality
# This example demonstrates retry behavior with different configurations

# Test retry on 5xx errors (httpbin.org/status/500 returns 500)
# Note: This will retry 3 times, then return 500 status
# The last_attempt_count will show how many retries were attempted
resource "httpx_request" "test_retry_500" {
  url    = "https://httpbin.org/status/500"
  method = "GET"

  retry {
    attempts             = 3
    min_delay_ms         = 500
    max_delay_ms         = 2000
    backoff              = "exponential"
    jitter               = true
    retry_on_status_codes = [500, 502, 503, 504]
    respect_retry_after  = true
  }

  # Expect 500 after retries are exhausted (retries will still happen)
  expect {
    status_codes = [500]
  }
}

# Test retry with exponential backoff (successful request)
resource "httpx_request" "test_retry_exponential" {
  url    = "https://httpbin.org/get"
  method = "GET"

  retry {
    attempts     = 5
    min_delay_ms = 100
    max_delay_ms = 1000
    backoff      = "exponential"
    jitter       = false
  }

  expect {
    status_codes = [200]
  }
}

# Test retry with linear backoff (successful request)
resource "httpx_request" "test_retry_linear" {
  url    = "https://httpbin.org/get"
  method = "GET"

  retry {
    attempts     = 3
    min_delay_ms = 200
    max_delay_ms = 1000
    backoff      = "linear"
    jitter       = true
  }

  expect {
    status_codes = [200]
  }
}

# Test retry with fixed delay (successful request)
resource "httpx_request" "test_retry_fixed" {
  url    = "https://httpbin.org/get"
  method = "GET"

  retry {
    attempts     = 3
    min_delay_ms = 500
    max_delay_ms = 500
    backoff      = "fixed"
    jitter       = false
  }

  expect {
    status_codes = [200]
  }
}

# Outputs to verify retry behavior
output "retry_attempt_counts" {
  value = {
    retry_500    = httpx_request.test_retry_500.last_attempt_count
    exponential  = httpx_request.test_retry_exponential.last_attempt_count
    linear       = httpx_request.test_retry_linear.last_attempt_count
    fixed        = httpx_request.test_retry_fixed.last_attempt_count
  }
  description = "Shows how many attempts were made for each retry configuration"
}

output "retry_500_status" {
  value     = httpx_request.test_retry_500.status_code
  sensitive = false
  description = "Status code after retries (should be 500)"
}
