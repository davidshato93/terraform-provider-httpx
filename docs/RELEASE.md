# Release Process

## Versioning

This provider follows [Semantic Versioning](https://semver.org/):
- **MAJOR**: Breaking changes (schema changes, removed attributes)
- **MINOR**: New features (new resources, new attributes, new blocks)
- **PATCH**: Bug fixes, documentation updates

## Pre-Release Checklist

- [ ] All tests pass (`go test ./...`)
- [ ] Linting passes (`golangci-lint run`)
- [ ] Documentation is up to date
- [ ] Examples are tested and working
- [ ] CHANGELOG.md is updated
- [ ] Version number is updated in:
  - `main.go` (version constant)
  - `README.md` (if mentioned)
  - `CHANGELOG.md`

## Release Steps

### 1. Update Version

Update the version constant in `main.go`:
```go
const version = "1.0.0"
```

### 2. Update CHANGELOG.md

Add a new section for the release:
```markdown
## [1.0.0] - 2024-01-15

### Added
- Initial release
- httpx_request resource
- httpx_request data source
- Retry and conditional retry support
- Extraction blocks
```

### 3. Create Release Tag

```bash
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

### 4. Build Release Artifacts

For each platform:
```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o terraform-provider-httpx_linux_amd64 .

# macOS
GOOS=darwin GOARCH=amd64 go build -o terraform-provider-httpx_darwin_amd64 .
GOOS=darwin GOARCH=arm64 go build -o terraform-provider-httpx_darwin_arm64 .

# Windows
GOOS=windows GOARCH=amd64 go build -o terraform-provider-httpx_windows_amd64.exe .
```

### 5. Create GitHub Release

1. Go to GitHub Releases
2. Click "Draft a new release"
3. Select the tag (e.g., `v1.0.0`)
4. Title: `v1.0.0`
5. Description: Copy from CHANGELOG.md
6. Upload binaries for each platform
7. Publish release

## Internal Distribution

### Option 1: Terraform Registry (Public)

If publishing to Terraform Registry:
1. Create a GitHub release (as above)
2. The registry will automatically detect the release
3. Follow [Terraform Registry publishing guide](https://www.terraform.io/docs/registry/providers/publishing.html)

### Option 2: Private Registry

For internal/private distribution:

1. **Host binaries** on internal artifact repository
2. **Configure Terraform** to use private registry:
   ```hcl
   terraform {
     required_providers {
       httpx = {
         source  = "internal-registry.example.com/namespace/httpx"
         version = "~> 1.0"
       }
     }
   }
   ```

### Option 3: Local Development Override

For local development/testing:

1. Build the provider:
   ```bash
   go build -o terraform-provider-httpx .
   ```

2. Configure `.terraformrc`:
   ```hcl
   provider_installation {
     dev_overrides {
       "registry.terraform.io/namespace/httpx" = "/path/to/provider/bin"
     }
   }
   ```

3. Skip `terraform init` when using dev_overrides

## Post-Release

- [ ] Announce release (internal channels)
- [ ] Update documentation if needed
- [ ] Monitor for issues
- [ ] Plan next release

## Rollback

If a critical issue is found:

1. **Immediate**: Mark release as deprecated in GitHub
2. **Fix**: Create patch release (e.g., `1.0.1`)
3. **Communicate**: Notify users of the issue and fix

## Version History

See `CHANGELOG.md` for detailed version history.

