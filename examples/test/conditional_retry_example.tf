# Example: Testing conditional retry (retry_until / poll-until)
# This demonstrates polling until conditions are met

# Example 1: Poll until JSON path condition is met
# Note: httpbin.org/json returns a fixed JSON structure
# This example shows the syntax, but httpbin always returns the same data
resource "httpx_request" "poll_json_path" {
  url    = "https://httpbin.org/json"
  method = "GET"

  retry {
    attempts     = 10
    min_delay_ms = 500
    max_delay_ms = 2000
    backoff      = "exponential"
    jitter       = true
  }

  retry_until {
    status_codes = [200]
    json_path_equals = {
      "slideshow.title" = "Sample Slide Show"
    }
  }

  timeouts {
    create = "5m"
  }

  expect {
    status_codes = [200]
  }
}

# Example 2: Poll until header condition is met
resource "httpx_request" "poll_header" {
  url    = "https://httpbin.org/get"
  method = "GET"

  retry {
    attempts     = 5
    min_delay_ms = 1000
    max_delay_ms = 3000
    backoff      = "linear"
  }

  retry_until {
    status_codes = [200]
    header_equals = {
      "Content-Type" = "application/json"
    }
  }

  expect {
    status_codes = [200]
  }
}

# Example 3: Poll until body regex matches
resource "httpx_request" "poll_regex" {
  url    = "https://httpbin.org/get"
  method = "GET"

  retry {
    attempts     = 5
    min_delay_ms = 500
    max_delay_ms = 2000
    backoff      = "fixed"
  }

  retry_until {
    status_codes = [200]
    body_regex = "\"url\":"
  }

  expect {
    status_codes = [200]
  }
}

# Example 4: Demonstrate actual polling behavior
# This will poll until a condition that will NEVER be met
# It will timeout after 1 minute, showing that polling actually happened
# resource "httpx_request" "poll_until_timeout" {
#   url    = "https://httpbin.org/json"
#   method = "GET"

#   retry {
#     attempts     = 100  # High number of attempts
#     min_delay_ms = 2000 # 2 seconds between attempts
#     max_delay_ms = 2000
#     backoff      = "fixed"
#     jitter       = false
#   }

#   # This condition will NEVER be met (httpbin.org/json always returns the same data)
#   # So it will keep polling until timeout
#   retry_until {
#     status_codes = [200]
#     json_path_equals = {
#       "slideshow.title" = "This Will Never Match"  # This value doesn't exist in httpbin.org/json
#     }
#   }

#   # Timeout after 1 minute - this will cause the polling to stop
#   timeouts {
#     create = "1m"
#   }

#   # Remove expect block to allow timeout error (or expect 200 to see it fail on timeout)
#   # expect {
#   #   status_codes = [200]
#   # }
# }

output "poll_attempts" {
  value = {
    json_path = httpx_request.poll_json_path.last_attempt_count
    header    = httpx_request.poll_header.last_attempt_count
    regex     = httpx_request.poll_regex.last_attempt_count
    # timeout   = httpx_request.poll_until_timeout.last_attempt_count
  }
  description = "Number of attempts made for each conditional retry"
}

# output "timeout_demo" {
#   value = {
#     attempts      = httpx_request.poll_until_timeout.last_attempt_count
#     last_status   = httpx_request.poll_until_timeout.status_code
#     last_error    = httpx_request.poll_until_timeout.last_error
#   }
#   description = "Shows polling behavior - should have multiple attempts before timeout"
# }

