# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial implementation of httpx provider
- httpx_request resource with full CRUD support
- httpx_request data source
- Retry engine with configurable backoff strategies
- Conditional retry (poll-until) support
- Extraction blocks for JSON path and header extraction
- Response validation with expect blocks
- TLS configuration (CA certs, client certs)
- Proxy support
- Sensitive data handling and redaction
- Response body size limits and truncation

### Changed
- Default `store_response_body` behavior: defaults to `false` when `extract` blocks are present
- Data sources default `store_response_body` to `false`

### Security
- Header redaction for sensitive headers
- Response body sensitivity marking
- Error message redaction

## [1.0.0] - TBD

### Added
- Initial release

