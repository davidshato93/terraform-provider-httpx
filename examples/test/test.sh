#!/bin/bash

# Test script for httpx provider

set -e

echo "=== Testing httpx Terraform Provider ==="
echo ""

# Get the directory of this script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROVIDER_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo "Provider directory: $PROVIDER_DIR"
echo "Test directory: $SCRIPT_DIR"
echo ""

# Check if provider binary exists
if [ ! -f "$PROVIDER_DIR/terraform-provider-httpx" ]; then
    echo "Error: Provider binary not found at $PROVIDER_DIR/terraform-provider-httpx"
    echo "Building provider..."
    cd "$PROVIDER_DIR"
    go build -o terraform-provider-httpx .
    echo "Provider built successfully!"
fi

# Set Terraform config
export TF_CLI_CONFIG_FILE="$SCRIPT_DIR/.terraformrc"

# Change to test directory
cd "$SCRIPT_DIR"

echo "=== Running Terraform Plan ==="
echo "Note: Skipping 'terraform init' as it's not needed with dev_overrides"
terraform plan

echo ""
echo "=== Running Terraform Apply ==="
terraform apply -auto-approve

echo ""
echo "=== Showing Outputs ==="
terraform output

echo ""
echo "=== Test Complete! ==="

