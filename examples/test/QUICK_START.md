# Quick Start Testing Guide

## The Issue with `terraform init`

When using `dev_overrides` for local provider development, Terraform will try to query the registry even though you're using a local provider. **This is expected behavior** - you can safely ignore the error and skip `terraform init`.

## How to Test (Skip `terraform init`)

1. **Navigate to test directory:**
   ```bash
   cd examples/test
   ```

2. **Set Terraform config to use local provider:**
   ```bash
   export TF_CLI_CONFIG_FILE="$(pwd)/.terraformrc"
   ```

3. **Skip `terraform init` and go straight to plan:**
   ```bash
   terraform plan
   ```

4. **Apply the configuration:**
   ```bash
   terraform apply
   ```

5. **View outputs:**
   ```bash
   terraform output
   ```

## Why Skip Init?

The warning message from Terraform says:
> "Skip terraform init when using provider development overrides. It is not necessary and may error unexpectedly."

When using `dev_overrides`, Terraform will:
- Still try to query the registry (which will fail)
- But then use the local provider from the override path
- This causes the error you see, but it's harmless

## Alternative: Use the Test Script

The `test.sh` script handles this automatically:

```bash
./test.sh
```

This script will:
- Build the provider if needed
- Set the TF_CLI_CONFIG_FILE
- Run `terraform plan` and `terraform apply`
- Show outputs

## Expected Results

After `terraform apply`, you should see:
- ✅ All 3 resources created successfully
- ✅ Status code: 200 for all requests  
- ✅ Response bodies with JSON from httpbin.org
- ✅ Outputs showing status codes

