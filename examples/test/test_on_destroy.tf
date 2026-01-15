# Test example for on_destroy feature
# This example demonstrates on_destroy in action using httpbin.org for testing
# 
# NOTE: The on_destroy block supports template interpolation with ${self.outputs.KEY} and ${self.id}
# but this happens at runtime in the provider (during destroy), not at plan time in Terraform.
# For testing, we use literal values or variables.

# Example 1: POST that extracts an ID, DELETE on destroy with extracted ID
resource "httpx_request" "test_on_destroy_basic" {
  method = "POST"
  url    = "https://httpbin.org/post"

  body_json = jsonencode({
    test_name = "on_destroy_basic"
    timestamp = "2024-01-15T10:00:00Z"
  })

  extract {
    name      = "request_id"
    json_path = "json.test_name"
  }

  on_destroy {
    method = "DELETE"
    # In a real scenario, this would be: "https://httpbin.org/delete?id=${self.outputs.request_id}"
    # The provider expands ${self.outputs.request_id} at destroy time
    url = "https://httpbin.org/delete?id=on_destroy_basic"

    expect {
      status_codes = [200, 404]
    }
  }
}

# Example 2: DELETE with retry config for transient failures
resource "httpx_request" "test_on_destroy_retry" {
  method = "POST"
  url    = "https://httpbin.org/post"

  body_json = jsonencode({
    resource_name = "test_retry_cleanup"
  })

  extract {
    name      = "resource_id"
    json_path = "json.resource_name"
  }

  on_destroy {
    method = "DELETE"
    # In a real scenario: "https://httpbin.org/delete?name=${self.outputs.resource_id}"
    url = "https://httpbin.org/delete?name=test_retry_cleanup"

    expect {
      status_codes = [200, 204, 404]
    }

    retry {
      attempts              = 3
      min_delay_ms          = 500
      max_delay_ms          = 1500
      backoff               = "exponential"
      retry_on_status_codes = [502, 503]
    }
  }
}

# Example 3: DELETE using self.id (resource identifier)
resource "httpx_request" "test_on_destroy_with_id" {
  method = "POST"
  url    = "https://httpbin.org/post"

  body_json = jsonencode({
    action = "create"
  })

  on_destroy {
    method = "DELETE"
    # In a real scenario: "https://httpbin.org/delete?resource_id=${self.id}"
    url = "https://httpbin.org/delete?resource_id=test-resource"

    header {
      name  = "X-Resource-ID"
      # In a real scenario: value = "${self.id}"
      value = "test-resource"
    }

    expect {
      status_codes = [200, 404]
    }
  }
}

# Output the resource IDs for verification
output "test_on_destroy_basic_id" {
  value = httpx_request.test_on_destroy_basic.id
}

output "test_on_destroy_retry_id" {
  value = httpx_request.test_on_destroy_retry.id
}

output "test_on_destroy_with_id_id" {
  value = httpx_request.test_on_destroy_with_id.id
}

output "extracted_request_id" {
  value = httpx_request.test_on_destroy_basic.outputs["request_id"]
}

output "extracted_resource_id" {
  value = httpx_request.test_on_destroy_retry.outputs["resource_id"]
}

