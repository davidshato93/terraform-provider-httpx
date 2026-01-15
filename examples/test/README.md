# Examples: httpx Terraform Provider

This directory contains test configurations and examples for the httpx provider.

## Quick Start

```bash
# Build the provider
cd ../..
go build -o terraform-provider-httpx .
cd examples/test

# Set up provider override
export TF_CLI_CONFIG_FILE="$(pwd)/.terraformrc"

# Plan and apply
terraform plan
terraform apply

# Clean up
terraform destroy
```

## Examples Included

### Basic Examples
- **`main.tf`** - Simple GET and POST requests with basic authentication
- **`retry_example.tf`** - Demonstrating retry configurations (exponential, linear, fixed backoff)
- **`conditional_retry_example.tf`** - Polling until conditions are met (retry_until)
- **`extraction_example.tf`** - Extracting values from responses using JSON paths and headers

### Advanced Examples
- **`datasource_example.tf`** - Using httpx_request as a read-only data source
- **`extract_vs_jsondecode.tf`** - Comparing extract blocks vs direct jsondecode()
- **`test_on_destroy.tf`** - Demonstrating the on_destroy feature for resource cleanup

### Verification Tools
- **`verify_destroy.tf`** - Simple test resource to verify on_destroy is working
- **`verify_destroy.sh`** - Automated script to test on_destroy with debug logging

## Documentation

- **[QUICK_START.md](./QUICK_START.md)** - Quick setup guide
- **[VERIFY_DESTROY.md](./VERIFY_DESTROY.md)** - How to verify on_destroy works
- **[ON_DESTROY_EXAMPLES.md](./ON_DESTROY_EXAMPLES.md)** - Detailed on_destroy examples

## Testing the Provider Locally

### Setup

1. Build the provider (from repo root):
   ```bash
   go build -o terraform-provider-httpx .
   ```

2. Enable provider override in Terraform:
   ```bash
   export TF_CLI_CONFIG_FILE="$(pwd)/examples/test/.terraformrc"
   cd examples/test
   ```

3. **Important**: Skip `terraform init` - When using dev_overrides, Terraform will try to query the registry and fail. This is expected. Just run:
   ```bash
   terraform plan
   terraform apply
   ```

### Running Tests

```bash
# Plan to see what will be created
terraform plan

# Apply to run the HTTP requests
terraform apply

# Destroy to clean up (triggers on_destroy for applicable resources)
terraform destroy
```

### With Debug Logging

```bash
# Enable Terraform debug logging
export TF_LOG=DEBUG
export TF_LOG_PATH=terraform-debug.log

# Run your test
terraform apply

# Check the logs
grep -i "delete\|destroy\|request" terraform-debug.log
```

## Expected Output

After `terraform apply`, you should see:
- ✅ Resources created successfully
- ✅ HTTP status codes (usually 200)
- ✅ Extracted values in outputs
- ✅ Response bodies or status codes logged

## Troubleshooting

| Issue | Solution |
|-------|----------|
| "Provider not found" | Build the provider with `go build -o terraform-provider-httpx` and verify `.terraformrc` path |
| "terraform init" fails | Don't run `terraform init` with dev_overrides - just run `terraform plan` directly |
| Network errors | Ensure internet connectivity to reach httpbin.org (or your test endpoint) |
| "Duplicate provider config" | Remove terraform/provider blocks from test files - they're defined in versions.tf |
| DEBUG logs not showing | Set `export TF_LOG=DEBUG` before running commands |

## Test Services

These examples use **httpbin.org**, a free HTTP testing service. If it's unavailable, update the URLs in the `.tf` files to your own test endpoint.

## Clean Up

To remove all test resources:

```bash
terraform destroy -auto-approve
```

This will:
1. Execute any configured `on_destroy` blocks (cleanup HTTP requests)
2. Remove resources from state
3. Remove any local state files

## Next Steps

- See [../../../docs/resources/request.md](../../../docs/resources/request.md) for full resource schema
- See [../../../docs/data-sources/request.md](../../../docs/data-sources/request.md) for data source schema
- See [../../../TESTING.md](../../../TESTING.md) for detailed testing procedures