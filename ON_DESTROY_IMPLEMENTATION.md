# on_destroy Feature Implementation

## Overview

The `on_destroy` feature allows users to execute HTTP requests when a resource is destroyed (during Terraform destroy). This is useful for cleanup operations like:

- Deleting resources from external APIs
- Deregistering webhooks
- Triggering cache invalidation
- Updating status endpoints
- Audit logging

## Implementation Summary

### 1. Models (Phase 1)

**File**: `internal/provider/models.go`

- Created `RequestConfigModel` struct containing shared request configuration fields
- Updated `HttpxRequestResourceModel` to embed `RequestConfigModel` for root request config
- Added `OnDestroy *RequestConfigModel` field for destroy configuration
- Avoids code duplication by reusing the same configuration model for both root and destroy requests

### 2. Schema (Phase 1)

**File**: `internal/provider/resource_request.go`

- Added `on_destroy` as an optional `SingleNestedBlock` in the resource schema
- Nested schema matches root request configuration fields:
  - Request config: `url`, `method`, `headers`, `query`, `body*`, `bearer_token`
  - Blocks: `header`, `basic_auth`, `retry`, `retry_until`, `expect`, `extract`
  - Security: `timeout_ms`, `insecure_skip_verify`, `proxy_url`, `response_sensitive`, `store_response_body`

### 3. Template Interpolation (Phase 2)

**File**: `internal/provider/interpolation.go`

Implements template expansion for `on_destroy` configuration:

- `${self.id}` - Expands to resource ID
- `${self.outputs.KEY}` - Expands to extracted output value from state
- Applied to: URL, headers (map and blocks), query parameters, body fields
- Error handling: Missing keys generate descriptive errors
- Context building from state: `BuildInterpolationContextFromState()`

### 4. Delete Execution (Phase 3)

**File**: `internal/provider/resource_request.go` - `Delete()` method

Implementation:

1. **No on_destroy**: Simple no-op (resource removed from state)
2. **on_destroy configured**:
   - Build interpolation context from state
   - Expand templates in destroy config fields
   - Build HTTP request using same `BuildRequest()` pipeline
   - Execute with `ExecuteRequestWithRetry()` using same retry/retry_until/expect logic
   - Validate expectations
   - On success: Remove from state
   - On failure: Keep state for Terraform retry

Features:

- Full retry support (exponential, linear, fixed backoff)
- Conditional retry with `retry_until` polling
- Expectation validation (status codes, JSON paths, headers)
- Error handling and logging
- State preservation on failure

### 5. Tests (Phase 4)

**Files**: 
- `internal/provider/interpolation_test.go` - Template interpolation tests
- `internal/provider/delete_test.go` - Delete functionality tests

Test Coverage:

- `TestInterpolateString()` - Template expansion with multiple patterns
- `TestInterpolateStringValue()` - Terraform types integration
- `TestInterpolateMap()` - Map value interpolation
- `TestInterpolateHeaderBlocks()` - Header block interpolation
- `TestBuildInterpolationContextFromState()` - State context building
- `TestDeleteWithoutOnDestroy()` - No-op behavior
- `TestDeleteWithOnDestroyConfig()` - Configuration validation
- `TestDeleteWithExtractedOutputs()` - Output reference resolution
- `TestDeleteWithRetryUntilCondition()` - Polling support
- `TestDeleteWithExpectBlock()` - Expectation validation
- `TestDeleteWithRetryConfig()` - Retry configuration
- `TestDeleteWithBasicAuth()` - Auth support
- `TestDeleteWithHeaderBlocks()` - Header template support
- `TestDeleteMissingInterpolationKey()` - Error handling

### 6. Documentation & Examples (Phase 5)

**Files**:
- `examples/test/on_destroy_example.tf` - Comprehensive examples
- `examples/test/test_on_destroy.tf` - Working test examples
- `README.md` - Updated with on_destroy feature
- `docs/resources/request.md` - Auto-generated schema documentation

Examples included:

1. Basic cleanup with extracted ID
2. DELETE with retry and 404 tolerance
3. DELETE with conditional polling
4. DELETE using self.id reference
5. DELETE with Basic Auth

### 7. Documentation Generation (Phase 6)

- Regenerated provider documentation with `go generate ./...`
- Schema includes full `on_destroy` block documentation
- All templates updated to reflect new capabilities

## Key Design Decisions

### 1. State-Based Interpolation

- Uses **state** (not plan) for interpolation context
- Provides reliable access to previously computed values
- Enables idempotent destroy operations

### 2. Template Patterns

Supported patterns:
- `${self.id}` - Resource identifier
- `${self.outputs.KEY}` - Extracted values (recommended pattern for cleanup)

Rationale:
- Simple and explicit
- Matches common Terraform patterns
- Avoids complex expression syntax

### 3. Error Behavior on Destroy Failure

**Fail-safe design**:
- If destroy HTTP request fails → state is retained
- Terraform will retry destroy on next `terraform destroy`
- User can investigate and fix issues

This prevents accidental data loss from transient failures.

### 4. Request Execution Pipeline

`on_destroy` uses **identical pipeline** to Create/Update:
- Request building
- Retry logic
- Conditional polling (retry_until)
- Expectation validation
- No custom logic

This ensures consistency and reliability.

### 5. Extract Handling in Destroy

- `extract` blocks evaluated during destroy request processing
- Results used for condition evaluation/logging
- **Not persisted** to state (resource is being deleted)

### 6. Default Delete Behavior

- No `on_destroy` block → no HTTP request (safe default)
- Respects `expect.status_codes` (no hardcoded 404 success)
- Timeouts use resource delete timeout

## Usage Examples

### Basic Delete

```hcl
resource "httpx_request" "user" {
  method = "POST"
  url    = "https://api.example.com/users"
  body_json = jsonencode({ name = "alice" })

  extract {
    name      = "user_id"
    json_path = "id"
  }

  on_destroy {
    method = "DELETE"
    url    = "https://api.example.com/users/${self.outputs.user_id}"
    expect {
      status_codes = [200, 204, 404]
    }
  }
}
```

### Delete with Retry

```hcl
on_destroy {
  method = "DELETE"
  url    = "https://api.example.com/resource/${self.outputs.resource_id}"

  retry {
    attempts     = 5
    backoff      = "exponential"
    min_delay_ms = 500
    max_delay_ms = 5000
    retry_on_status_codes = [429, 500, 502, 503]
  }

  expect {
    status_codes = [200, 204, 404]
  }
}
```

### Delete with Polling

```hcl
on_destroy {
  method = "DELETE"
  url    = "https://api.example.com/async/${self.outputs.id}"

  retry_until {
    status_codes = [200, 204, 404]
  }

  retry {
    attempts     = 10
    backoff      = "exponential"
    min_delay_ms = 1000
    max_delay_ms = 5000
  }
}
```

## Testing

All tests pass:

```
go test ./... -v
# 109 tests pass across all packages
```

## Backward Compatibility

✅ **Fully backward compatible**:
- `on_destroy` is optional
- Existing resources work unchanged
- No breaking changes to schema

## Files Modified

1. `internal/provider/models.go` - Shared model refactoring
2. `internal/provider/resource_request.go` - Schema + Delete implementation
3. `internal/provider/interpolation.go` - NEW: Template engine
4. `internal/provider/interpolation_test.go` - NEW: Interpolation tests
5. `internal/provider/delete_test.go` - NEW: Delete functionality tests
6. `examples/test/on_destroy_example.tf` - NEW: Documentation examples
7. `examples/test/test_on_destroy.tf` - NEW: Working test examples
8. `README.md` - Updated with on_destroy feature
9. `docs/resources/request.md` - Auto-generated documentation

## Future Enhancements

Possible additions (not in MVP):

1. **Destroy hooks for data sources** - Read-only data sources don't typically need destroy
2. **Multiple destroy blocks** - Sequence of HTTP calls on destroy
3. **Conditional on_destroy** - Execute destroy only if certain conditions met
4. **Custom error handling** - Override default fail-safe behavior
5. **Destroy-specific timeouts** - Separate timeout for destroy operations

## Validation

✅ Code compiles without errors  
✅ All tests pass (109 tests)  
✅ Documentation generated correctly  
✅ Examples provided and documented  
✅ Backward compatible  
✅ Error cases handled appropriately  
✅ Template interpolation validated  

