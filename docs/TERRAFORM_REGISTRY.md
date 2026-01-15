# Publishing to Terraform Registry

This guide covers the complete process of publishing your `terraform-provider-httpx` to the Terraform Registry.

## Prerequisites Checklist

Before publishing, ensure you have:

- [ ] **GPG key generated** and added to GitHub
- [ ] **GPG key added** to Terraform Registry
- [ ] **GitHub Secrets configured** (`GPG_PRIVATE_KEY`, `GPG_PASSPHRASE`, `GPG_KEY_ID`)
- [ ] **Repository is public** (required for Terraform Registry)
- [ ] **Repository follows naming convention**: `terraform-provider-httpx`

## Step-by-Step Publishing Guide

### 1. Set Up GPG Key

Follow the complete guide in [GPG_SETUP.md](GPG_SETUP.md) to:

1. Generate a GPG key
2. Add it to GitHub: https://github.com/settings/keys
3. Add it to Terraform Registry: https://registry.terraform.io/settings/gpg-keys

**Important:** Both steps are required! Terraform Registry uses GitHub's GPG key API to verify signatures.

### 2. Configure GitHub Secrets

For automated signing in GitHub Actions:

1. Go to: https://github.com/davidshato93/terraform-provider-httpx/settings/secrets/actions
2. Add these secrets:
   - `GPG_PRIVATE_KEY`: Your private key (export with `gpg --armor --export-secret-keys KEY_ID`)
   - `GPG_PASSPHRASE`: Your GPG key passphrase (if set)
   - `GPG_KEY_ID`: Your GPG key ID (e.g., `ABC123DEF4567890`)

### 3. Prepare Your Release

Before creating a release:

- [ ] All tests pass (`go test ./...`)
- [ ] Linting passes (`golangci-lint run`)
- [ ] Documentation is up to date
- [ ] CHANGELOG.md is updated
- [ ] Examples are tested

### 4. Create a Release

Create and push a version tag:

```bash
# Update CHANGELOG.md with release notes
# Commit your changes
git add .
git commit -m "Prepare release v1.0.0"
git push

# Create and push tag
git tag v1.0.0
git push origin v1.0.0
```

### 5. GitHub Actions Workflow

The release workflow will automatically:

1. ✅ Build binaries for all platforms:
   - Linux (amd64)
   - Windows (amd64)
   - macOS Intel (amd64)
   - macOS Apple Silicon (arm64)

2. ✅ Sign all binaries with your GPG key:
   - Creates `.asc` signature files for each binary
   - Uses your GPG key from GitHub Secrets

3. ✅ Create GitHub Release:
   - Attaches all binaries
   - Attaches all signature files
   - Includes release notes

### 6. Terraform Registry Detection

The Terraform Registry will:

1. **Automatically detect** your GitHub release (usually within 5-10 minutes)
2. **Verify GPG signatures** using your GitHub GPG key
3. **Publish your provider** if signatures are valid

### 7. Verify Publication

1. **Check Terraform Registry**:
   - Go to: `https://registry.terraform.io/providers/davidshato93/httpx`
   - Verify your provider appears
   - Check that releases show as verified

2. **Test provider installation**:
   ```hcl
   terraform {
     required_providers {
       httpx = {
         source  = "davidshato93/httpx"
         version = "~> 1.0"
       }
     }
   }
   ```

3. **Verify signatures manually**:
   ```bash
   # Download a binary and its signature
   wget https://github.com/davidshato93/terraform-provider-httpx/releases/download/v1.0.0/terraform-provider-httpx_linux_amd64
   wget https://github.com/davidshato93/terraform-provider-httpx/releases/download/v1.0.0/terraform-provider-httpx_linux_amd64.asc
   
   # Verify signature
   gpg --verify terraform-provider-httpx_linux_amd64.asc terraform-provider-httpx_linux_amd64
   ```

## Repository Requirements

Your repository must meet these requirements:

### Naming Convention

- ✅ Repository name: `terraform-provider-httpx`
- ✅ Provider name in code: `httpx`
- ✅ Provider source: `davidshato93/httpx`

### Repository Structure

```
terraform-provider-httpx/
├── main.go                 # Provider entry point
├── go.mod                  # Go module definition
├── README.md               # Provider documentation
├── CHANGELOG.md            # Version history
├── .github/
│   └── workflows/
│       └── release.yml     # Release automation
└── internal/
    └── provider/           # Provider implementation
```

### Required Files

- `README.md` - Provider documentation (displayed on Terraform Registry)
- `CHANGELOG.md` - Version history
- `.github/workflows/release.yml` - Automated releases

## Versioning

Follow [Semantic Versioning](https://semver.org/):

- **MAJOR** (1.0.0): Breaking changes
- **MINOR** (0.1.0): New features, backward compatible
- **PATCH** (0.0.1): Bug fixes

Tag format: `v1.0.0` (with `v` prefix)

## Troubleshooting

### Provider Not Appearing on Registry

- **Check repository is public**: Private repos cannot be published
- **Verify GPG key**: Must be added to both GitHub and Terraform Registry
- **Check release**: Must have signed binaries attached
- **Wait time**: Registry detection can take 5-10 minutes

### Signature Verification Fails

- **Verify GPG key**: Check it's added to both GitHub and Terraform Registry
- **Check GitHub Secrets**: Ensure `GPG_PRIVATE_KEY`, `GPG_PASSPHRASE`, `GPG_KEY_ID` are correct
- **Review workflow logs**: Check GitHub Actions logs for signing errors

### Release Not Detected

- **Check tag format**: Must be `v1.0.0` (semantic versioning)
- **Verify release exists**: Check GitHub Releases page
- **Check repository name**: Must match `terraform-provider-<name>`

## Additional Resources

- [Terraform Registry Publishing Guide](https://developer.hashicorp.com/terraform/registry/providers/publishing)
- [Terraform Registry Provider Requirements](https://developer.hashicorp.com/terraform/registry/providers/docs)
- [GPG Setup Guide](GPG_SETUP.md)
- [Release Process](RELEASE.md)

