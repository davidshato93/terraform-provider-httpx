# Quick test to verify on_destroy is working
# Note: provider and required_providers defined in versions.tf

# Create a simple test resource that we'll destroy
resource "httpx_request" "verify_destroy" {
  method = "POST"
  url    = "https://httpbin.org/post"

  body_json = jsonencode({
    test = "destroy_verification"
  })

  extract {
    name      = "test_id"
    json_path = "json.test"
  }

  on_destroy {
    method = "DELETE"
    url    = "https://httpbin.org/delete?test=destroy_verification"

    expect {
      status_codes = [200, 404]
    }

    # Add a small retry to show retry behavior in logs
    retry {
      attempts      = 2
      min_delay_ms  = 100
      max_delay_ms  = 200
      backoff       = "fixed"
    }
  }
}

output "resource_id" {
  value = httpx_request.verify_destroy.id
  description = "The ID of the created resource - will be used in destroy logs"
}

output "extracted_test_id" {
  value = httpx_request.verify_destroy.outputs["test_id"]
  description = "The extracted test_id value"
}

