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

### Automated Release (Recommended)

The release process is automated via GitHub Actions. When you push a tag, the workflow will:
1. Build binaries for all platforms (linux/amd64, windows/amd64, darwin/amd64, darwin/arm64)
2. Create a GitHub Release with all artifacts attached
3. Mark as pre-release if the tag contains a hyphen (e.g., `v1.0.0-beta.1`)

**To create a release:**

1. Ensure all changes are committed and pushed:
   ```bash
   git add .
   git commit -m "Prepare release v1.0.0"
   git push
   ```

2. Create and push a tag:
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

3. The GitHub Actions workflow will automatically:
   - Build all platform binaries
   - Create a GitHub Release
   - Attach all artifacts to the release

**Note:** The version is automatically extracted from the tag name, so you don't need to manually update `main.go` for releases.

### Manual Release (Alternative)

If you prefer to build and release manually:

#### 1. Update Version

Update the version constant in `main.go`:
```go
var version string = "1.0.0"
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

## Publishing to Terraform Registry

### Prerequisites

Before publishing to Terraform Registry, ensure:

1. ✅ **GPG key added to GitHub**: https://github.com/settings/keys
2. ✅ **GPG key added to Terraform Registry**: https://registry.terraform.io/settings/gpg-keys
3. ✅ **GitHub Secrets configured**: `GPG_PRIVATE_KEY`, `GPG_PASSPHRASE`, `GPG_KEY_ID`
4. ✅ **Provider repository follows naming**: `terraform-provider-<name>`
5. ✅ **Repository is public** (required for Terraform Registry)

See [GPG_SETUP.md](GPG_SETUP.md) for detailed GPG key setup instructions.

### Publishing Steps

1. **Create a release** by pushing a tag:
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

2. **GitHub Actions automatically**:
   - Builds binaries for all platforms
   - Signs all binaries with your GPG key
   - Creates a GitHub Release with signed artifacts

3. **Terraform Registry automatically**:
   - Detects your GitHub release (usually within minutes)
   - Verifies GPG signatures
   - Publishes your provider

4. **Verify your provider**:
   - Go to: `https://registry.terraform.io/providers/davidshato93/httpx`
   - Check that releases show as verified

### Terraform Registry Requirements

- **Repository naming**: Must be `terraform-provider-<name>`
- **Releases**: Must be signed with GPG
- **Version tags**: Must follow semantic versioning (e.g., `v1.0.0`)
- **Repository**: Must be public
- **GPG key**: Must be added to both GitHub and Terraform Registry

For more details, see:
- [Terraform Registry Publishing Guide](https://developer.hashicorp.com/terraform/registry/providers/publishing)
- [GPG Setup Guide](GPG_SETUP.md)

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

