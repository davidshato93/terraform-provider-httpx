# Example: Using on_destroy to clean up created resources

This example demonstrates how to use the `on_destroy` block to execute a DELETE request 
when the resource is destroyed, with template interpolation to reference extracted outputs.

## Setup

This example requires a local HTTP server for testing. You can use httpbin.org for demo purposes,
but for a complete example, you'll want to run a test server locally.

## Example 1: Basic cleanup with extracted ID

```hcl
resource "httpx_request" "create_user" {
  method = "POST"
  url    = "https://httpbin.org/post"

  body_json = jsonencode({
    name  = "Alice"
    email = "alice@example.com"
  })

  extract {
    name      = "user_id"
    json_path = "data.name"  # In real scenario, would extract actual ID
  }

  on_destroy {
    method = "DELETE"
    url    = "https://httpbin.org/delete?user_id=${self.outputs.user_id}"

    expect {
      status_codes = [200, 404]  # 404 is OK if already deleted
    }
  }
}
```

## Example 2: Delete with retry and 404 tolerance

```hcl
resource "httpx_request" "resource" {
  method = "POST"
  url    = "https://api.example.com/resources"

  body_json = jsonencode({
    name = "my-resource"
  })

  extract {
    name      = "resource_id"
    json_path = "id"
  }

  on_destroy {
    method = "DELETE"
    url    = "https://api.example.com/resources/${self.outputs.resource_id}"

    expect {
      status_codes = [200, 204, 404]
    }

    retry {
      attempts               = 3
      min_delay_ms           = 500
      max_delay_ms           = 2000
      backoff                = "exponential"
      retry_on_status_codes  = [429, 500, 502, 503]
    }
  }
}
```

## Example 3: Delete with conditional polling

```hcl
resource "httpx_request" "async_resource" {
  method = "POST"
  url    = "https://api.example.com/async-resources"

  body_json = jsonencode({
    name = "async-res"
  })

  extract {
    name      = "resource_id"
    json_path = "id"
  }

  on_destroy {
    method = "DELETE"
    url    = "https://api.example.com/async-resources/${self.outputs.resource_id}"

    retry_until {
      status_codes = [200, 204]  # Keep polling until we get success or 404
    }

    retry {
      attempts      = 10
      min_delay_ms  = 1000
      max_delay_ms  = 5000
      backoff       = "exponential"
    }
  }
}
```

## Example 4: Delete with request template using self.id

```hcl
resource "httpx_request" "webhook" {
  method = "POST"
  url    = "https://api.example.com/webhooks"

  body_json = jsonencode({
    url = "https://my-app.example.com/webhook"
  })

  on_destroy {
    method = "DELETE"
    url    = "https://api.example.com/webhooks/${self.id}"

    header {
      name  = "X-Delete-Reason"
      value = "Terraform destroy for resource ${self.id}"
    }

    expect {
      status_codes = [200, 204, 404]
    }
  }
}
```

## Example 5: Delete with Basic Auth

```hcl
resource "httpx_request" "protected_resource" {
  method = "POST"
  url    = "https://api.internal.example.com/resources"

  body_json = jsonencode({
    name = "protected"
  })

  extract {
    name      = "resource_id"
    json_path = "id"
  }

  on_destroy {
    method = "DELETE"
    url    = "https://api.internal.example.com/resources/${self.outputs.resource_id}"

    basic_auth {
      username = "admin"
      password = var.admin_password
    }

    expect {
      status_codes = [200, 204, 404]
    }
  }
}
```

## Behavior Notes

1. **No on_destroy block**: Resource is simply removed from Terraform state (no HTTP request).
2. **on_destroy block present**: When resource is destroyed, the HTTP request is executed.
3. **Expectations fail**: If `expect` validation fails, destroy fails and resource state is retained so Terraform can retry.
4. **Timeout**: Delete operations have a default 10-minute timeout, or use `timeouts.delete`.
5. **Template expansion**: Only `${self.outputs.KEY}` and `${self.id}` are available during destroy.
6. **Extraction in destroy**: Any `extract` blocks in on_destroy are evaluated for conditions but NOT persisted to state.

## When to Use on_destroy

- **Resource cleanup**: DELETE operations to remove objects created by this provider
- **Webhook deregistration**: Unregister webhooks or callbacks
- **Cache invalidation**: Trigger cache clear endpoints
- **Audit logging**: POST to audit endpoints when destroying resources
- **Status updates**: PUT requests to update status before deletion

