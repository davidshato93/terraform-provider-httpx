# Testing the httpx Terraform Provider

This directory contains test configurations for the httpx provider.

## Setup

1. Build the provider (if not already built):
   ```bash
   cd ../..
   go build -o terraform-provider-httpx .
   ```

2. Set up Terraform to use the local provider:
   ```bash
   export TF_CLI_CONFIG_FILE="$(pwd)/.terraformrc"
   ```

   Or copy `.terraformrc` to your home directory as `~/.terraformrc`

3. **Skip `terraform init`** - When using dev_overrides, Terraform will try to query the registry and fail. This is expected. Skip init and go straight to plan:
   ```bash
   terraform plan
   ```

4. Apply the configuration:
   ```bash
   terraform apply
   ```

## Test Cases

The configuration includes three test cases:

1. **test_get**: Simple GET request to httpbin.org with query parameters
2. **test_post**: POST request with JSON body
3. **test_auth**: GET request with Basic Authentication

All tests use httpbin.org, a free HTTP testing service.

## Expected Output

After running `terraform apply`, you should see:
- All three resources created successfully
- Status codes of 200 for all requests
- Response bodies containing the request/response data from httpbin.org

## Troubleshooting

If you encounter issues:

1. **Provider not found**: Make sure the provider binary is in the parent directory and `.terraformrc` is configured correctly
2. **Network errors**: Ensure you have internet connectivity to reach httpbin.org
3. **Build errors**: Make sure Go is installed and dependencies are downloaded (`go mod tidy`)

