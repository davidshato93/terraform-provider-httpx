# Gotchas and Best Practices

## Plan-Time vs Apply-Time Execution

### Data Sources (Plan-Time)
Data sources execute during `terraform plan` and `terraform apply`. This means:
- They run **every time** Terraform evaluates the configuration
- They can cause delays during planning
- Use them for read-only operations that don't create resources

```hcl
# ✅ Good: Fetching API status during plan
data "httpx_request" "api_status" {
  url    = "https://api.example.com/status"
  method = "GET"
}

# ⚠️ Warning: This runs on every plan
data "httpx_request" "expensive_query" {
  url    = "https://api.example.com/expensive-operation"
  method = "GET"
  retry {
    attempts = 10
    min_delay_ms = 5000
  }
}
```

### Resources (Apply-Time)
Resources execute during `terraform apply`:
- They create/update state
- They only run when configuration changes
- Use them for operations that create or modify external resources

```hcl
# ✅ Good: Creating a resource via API
resource "httpx_request" "create_item" {
  url    = "https://api.example.com/items"
  method = "POST"
  body_json = {
    name = "example"
  }
}
```

## State Size Management

### The Problem
Large response bodies stored in state can:
- Slow down Terraform operations
- Increase state file size
- Potentially expose sensitive data

### Solutions

#### 1. Use `extract` Blocks (Recommended)
Extract only what you need:

```hcl
resource "httpx_request" "example" {
  url    = "https://api.example.com/data"
  method = "GET"
  
  # Automatically sets store_response_body = false
  extract {
    name     = "id"
    json_path = "data.id"
  }
  
  extract {
    name     = "status"
    json_path = "data.status"
  }
}

# Use extracted values
output "item_id" {
  value = httpx_request.example.outputs["id"]
}
```

#### 2. Explicitly Disable Body Storage
```hcl
resource "httpx_request" "example" {
  url                = "https://api.example.com/data"
  method             = "GET"
  store_response_body = false  # Don't store body
  
  extract {
    name     = "id"
    json_path = "data.id"
  }
}
```

#### 3. Use `jsondecode()` for One-Time Parsing
If you only need to parse once and don't need extraction:

```hcl
resource "httpx_request" "example" {
  url    = "https://api.example.com/data"
  method = "GET"
  store_response_body = true
}

locals {
  parsed = jsondecode(httpx_request.example.response_body)
}

output "item_id" {
  value = local.parsed.data.id
}
```

**Note:** `extract` is better for:
- Multiple values
- Header extraction
- Minimizing state size
- Cleaner syntax

## Sensitive Data Handling

### Marking Response Bodies as Sensitive

```hcl
resource "httpx_request" "secret_data" {
  url              = "https://api.example.com/secrets"
  method           = "GET"
  response_sensitive = true  # Marks response_body as sensitive
}
```

### Provider-Level Redaction

Configure which headers to redact in logs and errors:

```hcl
provider "httpx" {
  redact_headers = [
    "Authorization",
    "Proxy-Authorization",
    "X-Api-Key",
    "Cookie"
  ]
}
```

### Best Practices
1. **Never store sensitive data unnecessarily**
   ```hcl
   # ❌ Bad: Storing full response with secrets
   resource "httpx_request" "bad" {
     url = "https://api.example.com/secrets"
     method = "GET"
     store_response_body = true  # Stores secrets in state!
   }
   
   # ✅ Good: Extract only what you need
   resource "httpx_request" "good" {
     url = "https://api.example.com/secrets"
     method = "GET"
     extract {
       name     = "token"
       json_path = "data.token"
     }
   }
   ```

2. **Use `response_sensitive` for sensitive responses**
   ```hcl
   resource "httpx_request" "sensitive" {
     url               = "https://api.example.com/user-data"
     method            = "GET"
     response_sensitive = true
   }
   ```

## Retry Behavior

### Retry vs Conditional Retry

**`retry` block**: Retries on failures (errors or specific status codes)
```hcl
retry {
  attempts = 3
  retry_on_status_codes = [500, 502, 503]
}
```

**`retry_until` block**: Polls until a condition is met
```hcl
retry_until {
  status_codes = [200]
  json_path_equals = {
    "data.status" = "ready"
  }
}
```

### Timeout Considerations

Always set timeouts for long-running operations:

```hcl
resource "httpx_request" "polling" {
  url    = "https://api.example.com/operation"
  method = "POST"
  
  retry_until {
    status_codes = [200]
  }
  
  retry {
    attempts     = 30
    min_delay_ms = 2000
    max_delay_ms = 10000
  }
  
  timeouts {
    create = "10m"  # Hard limit
  }
}
```

## Common Mistakes

### 1. Forgetting to Set Timeouts
```hcl
# ❌ Bad: Could hang indefinitely
resource "httpx_request" "bad" {
  retry_until {
    status_codes = [200]
  }
  retry {
    attempts = 1000  # Too many!
  }
}

# ✅ Good: Has timeout protection
resource "httpx_request" "good" {
  retry_until {
    status_codes = [200]
  }
  retry {
    attempts = 30
  }
  timeouts {
    create = "5m"
  }
}
```

### 2. Storing Large Bodies Unnecessarily
```hcl
# ❌ Bad: Stores 10MB response in state
resource "httpx_request" "bad" {
  url = "https://api.example.com/large-data"
  method = "GET"
  # store_response_body defaults to true
}

# ✅ Good: Extract only needed values
resource "httpx_request" "good" {
  url = "https://api.example.com/large-data"
  method = "GET"
  extract {
    name     = "count"
    json_path = "data.count"
  }
}
```

### 3. Not Handling Errors Properly
```hcl
# ❌ Bad: No error handling
resource "httpx_request" "bad" {
  url = "https://api.example.com/data"
  method = "GET"
}

# ✅ Good: Validates response
resource "httpx_request" "good" {
  url = "https://api.example.com/data"
  method = "GET"
  
  expect {
    status_codes = [200]
  }
  
  retry {
    attempts = 3
    retry_on_status_codes = [500, 502, 503]
  }
}
```

## Performance Tips

1. **Use data sources for read-only operations**
2. **Extract only what you need** to minimize state size
3. **Set appropriate timeouts** to avoid hanging operations
4. **Use retry for transient failures**, not for polling
5. **Use conditional retry (`retry_until`) for polling scenarios**

## Debugging

Enable debug logging:

```hcl
provider "httpx" {
  debug = true
}
```

Then run Terraform with debug logging:
```bash
TF_LOG=DEBUG terraform plan
```

This will show detailed request/response information (with sensitive data redacted).

