#!/bin/bash

# GPG Key Setup Script for Terraform Provider Releases
# This script helps you generate a GPG key for signing releases

set -e

echo "=========================================="
echo "GPG Key Setup for Terraform Provider"
echo "=========================================="
echo ""

# Check if GPG is installed
if ! command -v gpg &> /dev/null; then
    echo "Error: GPG is not installed."
    echo "Install it with: brew install gnupg"
    exit 1
fi

echo "This script will help you create a GPG key for signing releases."
echo ""
echo "You'll need:"
echo "  - Your name"
echo "  - Your GitHub email address (must match your GitHub account)"
echo "  - A strong passphrase"
echo ""

read -p "Press Enter to continue or Ctrl+C to cancel..."
echo ""

# Get user information
read -p "Enter your name (e.g., David Shato): " USER_NAME
read -p "Enter your GitHub email address: " USER_EMAIL
read -p "Enter a comment (optional, e.g., 'Terraform Provider Releases'): " USER_COMMENT

if [ -z "$USER_NAME" ] || [ -z "$USER_EMAIL" ]; then
    echo "Error: Name and email are required."
    exit 1
fi

echo ""
echo "Generating GPG key..."
echo "Key type: RSA and RSA"
echo "Key size: 4096 bits"
echo "Expiration: No expiration (you can change this later)"
echo ""

# Create batch file for GPG key generation
BATCH_FILE=$(mktemp)
cat > "$BATCH_FILE" <<EOF
Key-Type: RSA
Key-Length: 4096
Subkey-Type: RSA
Subkey-Length: 4096
Name-Real: $USER_NAME
Name-Email: $USER_EMAIL
Name-Comment: ${USER_COMMENT:-Terraform Provider Releases}
Expire-Date: 0
%no-protection
EOF

# Generate the key
gpg --batch --generate-key "$BATCH_FILE"

# Clean up batch file
rm "$BATCH_FILE"

echo ""
echo "=========================================="
echo "GPG Key Generated Successfully!"
echo "=========================================="
echo ""

# Get the key ID
KEY_ID=$(gpg --list-secret-keys --keyid-format LONG | grep -E "^(sec|ssb)" | head -1 | awk '{print $2}' | cut -d'/' -f2)

echo "Your GPG Key ID: $KEY_ID"
echo ""

# Export public key
PUBLIC_KEY_FILE="gpg-public-key-${KEY_ID}.asc"
gpg --armor --export "$KEY_ID" > "$PUBLIC_KEY_FILE"

echo "Public key exported to: $PUBLIC_KEY_FILE"
echo ""
echo "=========================================="
echo "Next Steps:"
echo "=========================================="
echo ""
echo "1. Add your public key to GitHub:"
echo "   - Go to: https://github.com/settings/keys"
echo "   - Click 'New GPG key'"
echo "   - Paste the contents of: $PUBLIC_KEY_FILE"
echo ""
echo "2. Add your public key to Terraform Registry (REQUIRED for publishing):"
echo "   - Go to: https://registry.terraform.io/settings/gpg-keys"
echo "   - Click 'Add a GPG Key'"
echo "   - Paste the contents of: $PUBLIC_KEY_FILE"
echo ""
echo "3. Export your private key for GitHub Actions:"
echo "   gpg --armor --export-secret-keys $KEY_ID > gpg-private-key.asc"
echo ""
echo "4. Add GitHub Secrets (go to: https://github.com/davidshato93/terraform-provider-httpx/settings/secrets/actions):"
echo "   - GPG_PRIVATE_KEY: Contents of gpg-private-key.asc"
echo "   - GPG_PASSPHRASE: Your GPG key passphrase (if you set one)"
echo "   - GPG_KEY_ID: $KEY_ID"
echo ""
echo "5. Configure Git to use your key:"
echo "   git config --global user.signingkey $KEY_ID"
echo "   git config --global commit.gpgsign true"
echo ""
echo "6. Test your setup:"
echo "   git commit --allow-empty -m 'Test GPG signing'"
echo "   git push"
echo ""
echo "7. Clean up private key file (IMPORTANT!):"
echo "   rm gpg-private-key.asc"
echo ""
echo "=========================================="
echo "Terraform Registry Publishing"
echo "=========================================="
echo ""
echo "Once your GPG key is set up:"
echo "1. Create a release: git tag v1.0.0 && git push origin v1.0.0"
echo "2. GitHub Actions will automatically sign all binaries"
echo "3. Terraform Registry will detect and verify your release"
echo ""
echo "For detailed instructions, see:"
echo "  - docs/GPG_SETUP.md (GPG setup)"
echo "  - docs/TERRAFORM_REGISTRY.md (Publishing guide)"
echo ""

