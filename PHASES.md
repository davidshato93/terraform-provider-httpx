# Implementation Phases

This document tracks the implementation progress of the Terraform HTTP Provider.

## ✅ Phase 0 — Scoping & Conventions (COMPLETE)

- [x] Decided provider name: `httpx`
- [x] Defined MVP schema (finalized blocks: `retry`, `retry_until`, `expect`, `extract`)
- [x] Defined state policy: defaults for response storage/sensitivity
- [x] Decided `httpx_wait` will be in v1.1 (not MVP)

**Deliverable:** README with final HCL interface and examples ✅

## ✅ Phase 1 — Provider Scaffold (COMPLETE)

- [x] Created repo with Terraform Plugin Framework boilerplate
- [x] Provider config schema + validation
- [x] HTTP client factory (timeouts, TLS, proxy)
- [x] Redaction utilities
- [x] Logging conventions (debug vs normal)

**Deliverable:** Provider loads, config validated, unit tests for config ✅

**Files Created:**
- `main.go` - Provider entry point
- `go.mod` - Go module definition
- `internal/provider/provider.go` - Provider implementation with config schema
- `internal/provider/resource_request.go` - Resource schema (stub implementation)
- `internal/provider/datasource_request.go` - Data source schema (stub implementation)
- `internal/client/http_client.go` - HTTP client factory with TLS/proxy support
- `internal/utils/redaction.go` - Header redaction utilities
- `README.md` - Documentation with examples
- `.gitignore` - Git ignore rules

**Status:** Provider compiles successfully. Ready for Phase 2.

## ✅ Phase 2 — Core Resource: httpx_request (COMPLETE)

- [x] Implement resource schema and CRUD skeleton
- [x] Implement request construction:
  - [x] method/url/query
  - [x] headers + repeated headers
  - [x] body/body_json/body_file selection
- [x] Execute request once and store computed attributes
- [x] Implement `expect.status_codes` + `expect.header_present`

**Deliverable:** `httpx_request` works for basic POST/GET with headers/body ✅

## ✅ Phase 3 — Retry Engine (COMPLETE)

- [x] Implement retry loop:
  - [x] attempts, delays, backoff, jitter
  - [x] retry on transport errors
  - [x] retry on configured HTTP codes
  - [x] optional `Retry-After` support
- [x] Plumb attempt count and diagnostics

**Deliverable:** Deterministic retry behavior verified ✅

## ✅ Phase 4 — Conditional Retry (poll-until) (COMPLETE)

- [x] Implement `retry_until` block:
  - [x] `status_codes`
  - [x] `json_path_equals` (dot-path evaluator)
  - [x] `header_equals`
  - [x] `body_regex`
- [x] Integrate Terraform timeouts as a hard deadline
- [x] Improve diagnostics for "condition not met yet"

**Deliverable:** "wait until isAttached=true" supported as first-class behavior ✅

## ✅ Phase 5 — Extraction and Outputs (COMPLETE)

- [x] Implement `extract` blocks:
  - [x] extract from JSON dot-path
  - [x] extract from header
- [x] Expose `outputs` as computed map
- [x] Smart defaults: `store_response_body` defaults to false when extract blocks present

**Deliverable:** Users can chain resources without parsing response bodies in HCL ✅

## ✅ Phase 6 — Data Source (COMPLETE)

- [x] Add `data httpx_request` (read-only)
- [x] Full feature parity with resource (retry, conditional retry, extraction)
- [x] Default `store_response_body = false` for data sources
- [x] Stable ID generation based on request inputs

**Deliverable:** Clean separation of "read" vs "apply-time action" ✅

**Note:** `httpx_wait` resource deferred to v1.1 as per design doc

## ✅ Phase 7 — Hardening and Release Readiness (COMPLETE)

- [x] Add max response size enforcement + truncation
- [x] Add comprehensive sensitive handling
- [x] Add TLS features (CA/client cert)
- [x] Add documentation:
  - [x] examples for auth, retries, conditional retry, extracts
  - [x] gotchas (plan-time calls, state size) - `docs/GOTCHAS.md`
- [x] Add CI:
  - [x] lint, unit tests - `.github/workflows/ci.yml`
  - [x] build artifacts for multiple platforms
- [x] Release process:
  - [x] versioning policy - `docs/RELEASE.md`
  - [x] CHANGELOG.md
  - [x] internal distribution setup documentation

**Deliverable:** v1.0 release ready ✅

**Files Created:**
- `docs/GOTCHAS.md` - Best practices and common gotchas
- `docs/RELEASE.md` - Release process documentation
- `CHANGELOG.md` - Version history
- `.github/workflows/ci.yml` - CI/CD pipeline

