# Example: Testing extraction functionality
# This demonstrates extracting values from responses for use in downstream resources

# Example 1: Extract values from JSON response
resource "httpx_request" "extract_json" {
  url    = "https://httpbin.org/json"
  method = "GET"

  expect {
    status_codes = [200]
  }

  # Extract values from JSON response
  extract {
    name      = "slide_title"
    json_path = "slideshow.title"
  }

  extract {
    name      = "slide_count"
    json_path = "slideshow.slides"
  }
}

# Example 2: Extract from headers
resource "httpx_request" "extract_headers" {
  url    = "https://httpbin.org/get"
  method = "GET"

  expect {
    status_codes = [200]
  }

  extract {
    name   = "content_type"
    header = "Content-Type"
  }

  extract {
    name   = "server"
    header = "Server"
  }
}

# Example 3: Extract both JSON and headers
resource "httpx_request" "extract_mixed" {
  url    = "https://httpbin.org/json"
  method = "GET"

  expect {
    status_codes = [200]
  }

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
    header = "Content-Type"
  }
}

# Example 4: Chain resources using extracted values
resource "httpx_request" "use_extracted" {
  url    = "https://httpbin.org/get"
  method = "GET"

  # Use extracted value from previous resource
  query = {
    title = httpx_request.extract_json.outputs["slide_title"]
  }

  expect {
    status_codes = [200]
  }
}

# Outputs to see extracted values
output "extracted_json_values" {
  value = {
    title = httpx_request.extract_json.outputs["slide_title"]
    # Note: slide_count will be a JSON array string
  }
  description = "Values extracted from JSON response"
}

output "extracted_header_values" {
  value = {
    content_type = httpx_request.extract_headers.outputs["content_type"]
    server       = httpx_request.extract_headers.outputs["server"]
  }
  description = "Values extracted from response headers"
}

output "all_outputs" {
  value = {
    json_resource   = httpx_request.extract_json.outputs
    header_resource = httpx_request.extract_headers.outputs
    mixed_resource  = httpx_request.extract_mixed.outputs
  }
  description = "All extracted outputs from all resources"
}

output "body_resource" {
  value = {
    body = jsondecode(httpx_request.use_extracted.response_body).args.title
  }
  description = "Response body from mixed resource"
}
