#!/bin/bash

# Script to verify on_destroy feature with detailed logging

set -e

PROVIDER_DIR="/Users/davidshato/Documents/projects/self-projects/project-to-github/terraform-provider-httpx"
TEST_DIR="$PROVIDER_DIR/examples/test"

echo "========================================="
echo "on_destroy Feature Verification"
echo "========================================="
echo ""

# Enable debug logging
export TF_LOG=DEBUG
export TF_LOG_PATH="$TEST_DIR/terraform-debug.log"

cd "$TEST_DIR"

# Clean up any existing state
echo "1. Cleaning up old state..."
rm -f terraform.tfstate terraform.tfstate.backup .terraform.lock.hcl

# Export provider override
export TF_CLI_CONFIG_FILE=.terraformrc

echo ""
echo "2. Creating resources (this will execute Create and extract values)..."
terraform plan -target=httpx_request.verify_destroy -no-color > /dev/null
terraform apply -target=httpx_request.verify_destroy -auto-approve -no-color

echo ""
echo "3. Resource created successfully!"
echo "   Outputs:"
terraform output -json | jq '.resource_id, .extracted_test_id'

echo ""
echo "4. Now destroying resource (this will execute on_destroy)..."
echo "   Watch for DELETE request in the logs below:"
echo ""

# Run destroy with debug output visible
terraform destroy -target=httpx_request.verify_destroy -auto-approve -no-color 2>&1 | tail -20

echo ""
echo "========================================="
echo "Debug logs written to: terraform-debug.log"
echo "========================================="
echo ""
echo "To see all on_destroy activity, run:"
echo "  grep -i 'destroy\|delete\|interpolate' terraform-debug.log"
echo ""
echo "Key things to look for:"
echo "  - 'Delete method called' - confirms Delete() was invoked"
echo "  - 'executing on_destroy request' - confirms on_destroy block found"
echo "  - 'Interpolated' - shows template expansion"
echo "  - 'Destroy request' - shows HTTP request details"
echo "  - 'status code' - shows response status"
echo ""

