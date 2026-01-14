# Testing Guide

## Quick Start

### Option 1: Using the test script (easiest)

```bash
cd examples/test
./test.sh
```

### Option 2: Manual testing

1. **Navigate to the test directory:**
   ```bash
   cd examples/test
   ```

2. **Set up Terraform to use the local provider:**
   ```bash
   export TF_CLI_CONFIG_FILE="$(pwd)/.terraformrc"
   ```

3. **Initialize Terraform:**
   ```bash
   terraform init
   ```

4. **Plan the changes:**
   ```bash
   terraform plan
   ```

5. **Apply the configuration:**
   ```bash
   terraform apply
   ```

6. **View outputs:**
   ```bash
   terraform output
   ```

7. **Clean up (optional):**
   ```bash
   terraform destroy
   ```

## What the tests do

The test configuration (`examples/test/main.tf`) includes three test cases:

1. **GET request** - Tests basic GET with query parameters
2. **POST request** - Tests POST with JSON body encoding
3. **Basic Auth** - Tests Basic Authentication

All tests use [httpbin.org](https://httpbin.org), a free HTTP testing service.

## Troubleshooting

### Provider not found

If Terraform can't find the provider:

1. Make sure the provider binary exists:
   ```bash
   ls -la ../../terraform-provider-httpx
   ```

2. If it doesn't exist, build it:
   ```bash
   cd ../..
   go build -o terraform-provider-httpx .
   ```

3. Verify the `.terraformrc` path is correct (should point to the directory containing the binary)

### Network errors

- Ensure you have internet connectivity
- httpbin.org should be accessible
- Check firewall/proxy settings if needed

### Build errors

If you get Go build errors:

```bash
cd ../..
go mod tidy
go build -o terraform-provider-httpx .
```

## Expected Results

After successful `terraform apply`, you should see:

- All three resources created
- Status codes: 200 for all requests
- Response bodies containing JSON from httpbin.org
- Outputs showing status codes and response data

## Testing Custom Scenarios

You can modify `examples/test/main.tf` to test other scenarios:

- Different HTTP methods (PUT, PATCH, DELETE)
- Custom headers
- File-based body (`body_file`)
- Different endpoints
- Error cases (4xx, 5xx status codes)

