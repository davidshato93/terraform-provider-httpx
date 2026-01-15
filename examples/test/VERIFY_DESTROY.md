# Verifying on_destroy Execution

This guide explains how to verify that the `on_destroy` feature is working correctly and see the destroy requests being made.

## Overview

When you run `terraform destroy` on a resource with an `on_destroy` block configured, the provider will:

1. Read the current resource state
2. Build interpolation context from state values
3. Expand templates in the on_destroy configuration
4. Execute the HTTP request with retry/polling logic
5. Log all activity if debug is enabled

## Method 1: Debug Logging (Recommended)

### Enable Debug Mode

```bash
export TF_LOG=DEBUG
export TF_LOG_PATH=terraform-debug.log
cd examples/test
terraform destroy -auto-approve
```

### What to Look For in Logs

```bash
grep -E "Delete method|executing on_destroy|Interpolated|Destroy request|status code" terraform-debug.log
```

**Key log entries:**

```
[INFO] Delete method called - executing on_destroy request
[TRACE] Interpolated ${self.id} -> abc123def456
[TRACE] Interpolated ${self.outputs.user_id} -> user-789
[INFO] Built HTTP request: DELETE https://api.example.com/users/user-789
[INFO] Destroy request succeeded with status code 200
```

### Debug Log Levels

- **DEBUG**: All activity including HTTP requests, retries, templates
- **TRACE**: More detailed trace of internal operations
- **INFO**: High-level operations like "Delete called", "request succeeded"
- **WARN**: Warnings about retries, timeouts
- **ERROR**: Errors during execution

## Method 2: Using the Verification Script

A ready-made script is included:

```bash
cd examples/test
./verify_destroy.sh
```

This script:
1. Enables debug logging automatically
2. Creates a test resource
3. Destroys it
4. Shows you exactly what to look for in the logs

## Method 3: State Inspection

Before destroying, check the state to see what values will be used:

```bash
terraform show

# Look for outputs section:
# outputs = {
#   "user_id" = "user-456"
# }
```

These extracted values will be used for `${self.outputs.user_id}` interpolation.

## Method 4: Manual Testing

Create a test resource with a known on_destroy config:

```hcl
resource "httpx_request" "test" {
  method = "POST"
  url    = "https://httpbin.org/post"
  
  body_json = jsonencode({
    name = "test"
  })

  extract {
    name      = "resource_id"
    json_path = "json.name"
  }

  on_destroy {
    method = "DELETE"
    url    = "https://httpbin.org/delete?id=${self.outputs.resource_id}"
    
    expect {
      status_codes = [200, 404]
    }
  }
}
```

Then:

```bash
# Enable logging
export TF_LOG=DEBUG
export TF_LOG_PATH=test-destroy.log

# Apply to create
terraform apply -target=httpx_request.test -auto-approve

# Destroy to trigger on_destroy
terraform destroy -target=httpx_request.test -auto-approve

# Check logs
tail -100 test-destroy.log
```

## Expected Log Output

### Success Case

```
2024-01-15T10:30:45.123Z [INFO]  Delete method called - executing on_destroy request
2024-01-15T10:30:45.124Z [DEBUG] Building interpolation context from state
2024-01-15T10:30:45.125Z [TRACE] Interpolated ${self.outputs.resource_id} -> test
2024-01-15T10:30:45.126Z [DEBUG] Expanding on_destroy config URL: https://httpbin.org/delete?id=test
2024-01-15T10:30:45.127Z [DEBUG] Building HTTP DELETE request
2024-01-15T10:30:45.128Z [DEBUG] Executing request with retry config: attempts=3, backoff=exponential
2024-01-15T10:30:45.200Z [INFO]  Destroy request succeeded with status code 200
2024-01-15T10:30:45.201Z [INFO]  Expectation validation passed: status code 200 in [200, 404]
```

### Retry Case

```
2024-01-15T10:30:45.123Z [INFO]  Delete method called - executing on_destroy request
2024-01-15T10:30:45.126Z [DEBUG] Building HTTP DELETE request
2024-01-15T10:30:45.127Z [DEBUG] Executing request with retry config
2024-01-15T10:30:45.200Z [WARN]  Request failed with status 503, retrying (attempt 1/3)...
2024-01-15T10:30:45.250Z [DEBUG] Calculating delay: exponential backoff, min=500ms, max=5000ms
2024-01-15T10:30:46.500Z [DEBUG] Retrying request (attempt 2/3)
2024-01-15T10:30:46.600Z [INFO]  Destroy request succeeded with status code 200
```

### Failure Case

```
2024-01-15T10:30:45.123Z [INFO]  Delete method called - executing on_destroy request
2024-01-15T10:30:45.127Z [DEBUG] Executing request with retry config
2024-01-15T10:30:45.200Z [ERROR] Request failed: connection refused
2024-01-15T10:30:45.201Z [ERROR] Destroy request failed: connection refused
2024-01-15T10:30:45.202Z [INFO]  State retained for retry on next destroy attempt
```

## Checking HTTP Requests Made

### Via Logs

```bash
grep "DELETE\|POST\|PUT\|PATCH" terraform-debug.log | grep -v "response_body"
```

### Via External Monitoring

If your on_destroy makes requests to an external API:

1. Check server access logs
2. Enable request tracing on the API side
3. Use tools like `curl -X DELETE https://api.example.com/resource/123 -v` to test manually

### Via httpbin.org (Testing)

The test examples use httpbin.org which echoes requests back:

```bash
# Create a test resource (creates entry at httpbin)
terraform apply -target=httpx_request.test -auto-approve

# Destroy and make DELETE request (httpbin returns the request details)
terraform destroy -target=httpx_request.test -auto-approve

# See the DELETE request that was sent to httpbin in logs
grep -A 20 "DELETE request succeeded" terraform-debug.log
```

## Troubleshooting: on_destroy Not Called

**Problem**: You don't see "Delete method called" in logs

**Possible causes**:

1. **Resource already removed** - State file doesn't exist
   ```bash
   terraform state list  # Check if resource still in state
   ```

2. **Wrong resource name** - Using wrong target
   ```bash
   terraform destroy -target=httpx_request.wrong_name  # Wrong!
   terraform destroy -target=httpx_request.test        # Correct!
   ```

3. **No on_destroy block** - If on_destroy is not configured, destroy is no-op
   ```bash
   terraform show | grep "on_destroy"  # Check state
   ```

4. **Terraform plan-only** - Destroy doesn't run in plan mode
   ```bash
   terraform plan -destroy  # Just shows plan
   terraform destroy        # Actually runs destroy
   ```

## Advanced: Capturing Full Request/Response

Add to logs to see exact HTTP details:

```bash
# Very detailed debugging
export TF_LOG=TRACE
export TF_LOG_PATH=trace.log

terraform destroy -auto-approve

# View all HTTP details
grep -A 5 -B 5 "request\|response" trace.log | head -100
```

## Key Files to Check

- `terraform-debug.log` - Main debug log
- `terraform.tfstate` - State file (contains extracted values)
- `.terraform/logs/` - Additional provider logs (if available)

## Summary

To verify on_destroy is working:

1. ✅ Enable debug logging: `export TF_LOG=DEBUG`
2. ✅ Create a resource: `terraform apply`
3. ✅ Destroy it: `terraform destroy`
4. ✅ Check logs: `grep -i "destroy\|delete" terraform-debug.log`
5. ✅ Verify HTTP requests were made to expected endpoints

The logs will show:
- ✅ Delete method was called
- ✅ on_destroy block was found and executed
- ✅ Templates were expanded correctly
- ✅ HTTP requests were sent
- ✅ Responses were validated
- ✅ Retries occurred (if configured)

