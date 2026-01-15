# Repository Structure & Organization

This document describes the clean, professional structure of the terraform-provider-httpx repository.

## Root Level Files

Essential files at the repository root:

| File | Purpose |
|------|---------|
| `main.go` | Provider entry point and setup |
| `go.mod`, `go.sum` | Go module dependencies |
| `README.md` | Main documentation and quick start |
| `TESTING.md` | Local testing procedures |
| `CHANGELOG.md` | Version history |
| `.gitignore` | Git ignore patterns |
| `.golangci.yml` | Linter configuration |
| `tfplugindocs.yaml` | Documentation generation config |
| `terraform-registry-manifest.json` | Terraform Registry metadata |

## Directory Structure

```
terraform-provider-httpx/
├── internal/              # Provider implementation
│   ├── provider/          # Resource & data source definitions
│   ├── client/            # HTTP client factory
│   ├── config/            # Shared configuration
│   └── utils/             # Utility functions
├── docs/                  # Documentation
│   ├── index.md           # Generated provider docs
│   ├── resources/         # Generated resource docs
│   ├── data-sources/      # Generated data source docs
│   ├── ON_DESTROY_IMPLEMENTATION.md
│   ├── RELEASE_SUMMARY.md
│   ├── RELEASE.md
│   ├── GPG_SETUP.md
│   ├── GOTCHAS.md
│   └── TERRAFORM_REGISTRY.md
├── examples/              # Example configurations
│   └── test/              # Test examples and guides
│       ├── README.md      # Quick start guide
│       ├── VERIFY_DESTROY.md
│       ├── ON_DESTROY_EXAMPLES.md
│       ├── *.tf           # Example Terraform files
│       └── .terraformrc   # Provider dev override config
├── scripts/               # Build and release scripts
│   ├── setup-gpg.sh
│   └── Makefile (optional)
├── tools/                 # Go tool dependencies
│   └── tools.go
├── .github/               # GitHub configuration
│   └── workflows/         # GitHub Actions CI/CD
│       ├── ci.yml         # Linting, testing, building
│       └── release.yml    # Release automation
└── templates/             # Terraform plugin docs templates

```

## Documentation Organization

### User Documentation

Located in `README.md` with links to:
- Provider configuration examples
- Resource and data source schemas (auto-generated in `docs/`)
- Examples and use cases
- Best practices and gotchas

### Developer Documentation

Located in `docs/`:
- **ON_DESTROY_IMPLEMENTATION.md** - Implementation details of the on_destroy feature
- **RELEASE_SUMMARY.md** - Summary of completed development phases
- **RELEASE.md** - Release process and versioning
- **GPG_SETUP.md** - GPG key generation for signing releases
- **GOTCHAS.md** - Common issues and solutions
- **TERRAFORM_REGISTRY.md** - Publishing to Terraform Registry
- **LINTING.md** - Code quality standards

### Testing and Examples

Located in `examples/test/`:
- **README.md** - Comprehensive testing guide
- **QUICK_START.md** - Quick setup for local testing
- **VERIFY_DESTROY.md** - How to verify on_destroy feature
- **ON_DESTROY_EXAMPLES.md** - Detailed on_destroy examples
- `*.tf` - Example Terraform configurations
- `test.sh`, `verify_destroy.sh` - Automated test scripts

## Key Design Principles

### 1. Clean Repository

✅ No temporary files in repo:
- terraform.tfstate files ignored
- terraform-debug.log files ignored
- Build artifacts ignored

✅ Organized documentation:
- Implementation docs in `docs/`
- Examples in `examples/test/`
- Only essential files at root

### 2. Clear Navigation

✅ README.md guides users:
- Feature overview at top
- "For Users" section with links
- "For Developers" section with links
- Development procedures clearly documented

✅ Each section self-contained:
- Local testing: See TESTING.md
- Verifying on_destroy: See examples/test/VERIFY_DESTROY.md
- Releases: See docs/RELEASE.md

### 3. Professional Structure

✅ Follows Terraform Provider conventions:
- `internal/` for implementation
- `docs/` for generated and custom documentation
- `examples/` for usage examples

✅ CI/CD integrated:
- GitHub Actions workflows in `.github/workflows/`
- Automated testing on push
- Automated releases on tags

## File Categories

### Source Code
- `main.go` - Entry point
- `internal/**/*.go` - Implementation
- `tools/tools.go` - Tool imports

### Configuration
- `go.mod`, `go.sum` - Dependencies
- `.gitignore` - Git settings
- `.golangci.yml` - Linter config
- `tfplugindocs.yaml` - Doc generation config

### Documentation
- `README.md` - Main documentation
- `TESTING.md` - Testing guide
- `CHANGELOG.md` - Version history
- `docs/*.md` - Implementation and guide docs

### Examples
- `examples/test/*.tf` - Terraform configs
- `examples/test/*.md` - Example guides
- `examples/test/*.sh` - Test scripts

### CI/CD
- `.github/workflows/*.yml` - GitHub Actions
- `.github/workflows/` - Build and release configs

## Maintenance Guidelines

### Adding New Documentation

1. **User Guide** → Add to `examples/test/` or link from README
2. **Implementation Detail** → Add to `docs/` directory
3. **Example** → Add `.tf` file to `examples/test/`

### Updating Examples

1. Ensure `.tf` files follow Terraform best practices
2. Update `examples/test/README.md` if adding new examples
3. Test locally with `.terraformrc` before committing

### Release Process

1. Update `CHANGELOG.md`
2. Create git tag: `git tag v1.2.3`
3. Push: `git push origin main --tags`
4. GitHub Actions runs tests and creates release
5. Release artifacts uploaded with GPG signatures

## Best Practices

✅ **Keep root level clean** - Only essential files
✅ **Document in right place** - Implementation docs in `docs/`, examples in `examples/`
✅ **Use consistent naming** - `on_destroy_example.tf`, `VERIFY_DESTROY.md`
✅ **Link between docs** - Make navigation easy for users
✅ **Ignore test artifacts** - `.gitignore` prevents state/log commits
✅ **Version in CHANGELOG** - Track all changes
✅ **Test before commit** - Ensure examples work

## Quick Reference

| Task | Where | File |
|------|-------|------|
| Understand provider | User | README.md |
| Test locally | User | TESTING.md or examples/test/README.md |
| Verify on_destroy | User | examples/test/VERIFY_DESTROY.md |
| View examples | User | examples/test/*.tf |
| Understand on_destroy | Dev | docs/ON_DESTROY_IMPLEMENTATION.md |
| Release | Dev | docs/RELEASE.md |
| Set up GPG | Dev | docs/GPG_SETUP.md |
| Lint code | Dev | .golangci.yml |
| CI/CD pipeline | Dev | .github/workflows/ |

