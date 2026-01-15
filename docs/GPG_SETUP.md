# GPG Key Setup for Release Signing

This guide will help you set up GPG signing for your Terraform provider releases, including publishing to the **Terraform Registry**.

## Prerequisites

- macOS (with Homebrew) or Linux
- GitHub account
- Terraform Registry account (linked to your GitHub account)

## Step 1: Install GPG

### macOS (using Homebrew)
```bash
brew install gnupg
```

### Linux (Ubuntu/Debian)
```bash
sudo apt-get update
sudo apt-get install gnupg
```

### Linux (RHEL/CentOS)
```bash
sudo yum install gnupg
```

## Step 2: Generate GPG Key

Run the following command to generate a new GPG key:

```bash
gpg --full-generate-key
```

**Recommended settings:**
1. **Key type**: `(1) RSA and RSA` (default)
2. **Key size**: `4096` (recommended for security)
3. **Expiration**: `0` (no expiration) or set a date (e.g., `2y` for 2 years)
4. **Real name**: Your name (e.g., `David Shato`)
5. **Email address**: Your GitHub email address (must match your GitHub account email)
6. **Comment**: Optional (e.g., `Terraform Provider Releases`)
7. **Passphrase**: Choose a strong passphrase (you'll need this for signing)

**Example interaction:**
```
gpg (GnuPG) 2.4.0; Copyright (C) 2021 Free Software Foundation, Inc.
Please select what kind of key you want:
   (1) RSA and RSA (default)
   (2) DSA and Elgamal
   (3) DSA (sign only)
   (4) RSA (sign only)
  (14) Existing key from card
Your selection? 1
RSA keys may be between 1024 and 4096 bits long.
What keysize do you want? (3072) 4096
Requested keysize is 4096 bits
Please specify how long the key should be valid.
         0 = key does not expire
      <n>  = key expires in n days
      <n>w = key expires in n weeks
      <n>m = key expires in n months
      <n>y = key expires in n years
Key is valid for? (0) 0
Key does not expire at all
Is this correct? (y/N) y

GnuPG needs to construct a user ID to identify your key.

Real name: David Shato
Email address: your-email@example.com
Comment: Terraform Provider Releases
You selected this USER-ID:
    "David Shato (Terraform Provider Releases) <your-email@example.com>"

Change (N)ame, (C)omment, (E)mail or (O)kay/(Q)uit? O
```

## Step 3: List Your GPG Keys

After generating the key, list it to get the key ID:

```bash
gpg --list-secret-keys --keyid-format LONG
```

You'll see output like:
```
sec   rsa4096/ABC123DEF4567890 2024-01-15 [SC]
      ABC123DEF4567890ABCDEF1234567890ABCDEF12
uid                 [ultimate] David Shato (Terraform Provider Releases) <your-email@example.com>
ssb   rsa4096/XYZ789ABC1234567 2024-01-15 [E]
```

The key ID is `ABC123DEF4567890` (the part after `rsa4096/`).

## Step 4: Export Your Public Key

Export your public key in ASCII format:

```bash
gpg --armor --export YOUR_KEY_ID
```

Replace `YOUR_KEY_ID` with your actual key ID from Step 3.

**Save this output** - you'll need it for GitHub.

## Step 5: Add GPG Key to GitHub

**This step is required for Terraform Registry!** The Terraform Registry uses GitHub's GPG key API to verify signatures.

1. Go to GitHub Settings: https://github.com/settings/keys
2. Click **"New GPG key"**
3. Paste your public key (from Step 4)
4. Click **"Add GPG key"**
5. Confirm with your GitHub password

**Important:** Your GPG key must be added to GitHub for Terraform Registry to verify your releases.

## Step 6: Configure Git to Use Your GPG Key

Tell Git to use your GPG key for signing commits:

```bash
git config --global user.signingkey YOUR_KEY_ID
git config --global commit.gpgsign true
```

## Step 7: Export GPG Key for GitHub Actions

For GitHub Actions to sign releases, you need to:

1. **Export your private key** (keep this secure!):
   ```bash
   gpg --armor --export-secret-keys YOUR_KEY_ID > private-key.asc
   ```

2. **Add it as a GitHub Secret**:
   - Go to your repository: https://github.com/davidshato93/terraform-provider-httpx/settings/secrets/actions
   - Click **"New repository secret"**
   - Name: `GPG_PRIVATE_KEY`
   - Value: Paste the contents of `private-key.asc`
   - Click **"Add secret"**

3. **Add your passphrase** (if you set one):
   - Click **"New repository secret"**
   - Name: `GPG_PASSPHRASE`
   - Value: Your GPG key passphrase
   - Click **"Add secret"**

4. **Add your key ID**:
   - Click **"New repository secret"**
   - Name: `GPG_KEY_ID`
   - Value: Your key ID (e.g., `ABC123DEF4567890`)
   - Click **"Add secret"**

5. **Clean up** (important!):
   ```bash
   rm private-key.asc
   ```

## Step 8: Test Your GPG Key

Test that your GPG key works:

```bash
echo "test" | gpg --clearsign
```

You should see signed output. Press `Ctrl+C` to exit.

## Step 9: Add GPG Key to Terraform Registry

**Required for publishing to Terraform Registry:**

1. Sign in to [Terraform Registry](https://registry.terraform.io/) using your GitHub account
2. Go to [User Settings > Signing Keys](https://registry.terraform.io/settings/gpg-keys)
3. Click **"Add a GPG Key"**
4. Paste your public key (from Step 4)
5. Click **"Add GPG Key"**

The Terraform Registry will use this key to verify signatures on your provider releases.

## Step 10: Verify GitHub Integration

1. Make a test commit:
   ```bash
   git commit --allow-empty -m "Test GPG signing"
   git push
   ```

2. Check on GitHub - the commit should show "Verified" badge.

## Step 11: Verify Terraform Registry Integration

After creating your first release:

1. Go to your provider page on Terraform Registry
2. Check that releases show as verified/signed
3. Users can verify signatures when downloading:
   ```bash
   gpg --verify terraform-provider-httpx_<platform>.<ext>.asc terraform-provider-httpx_<platform>.<ext>
   ```

## Troubleshooting

### GPG not found
- Make sure GPG is installed and in your PATH
- On macOS, you may need to add `export PATH="/opt/homebrew/bin:$PATH"` to your `~/.zshrc`

### Wrong email address
- Your GPG key email must match your GitHub email
- Check your GitHub email: https://github.com/settings/emails
- Update your GPG key: `gpg --edit-key YOUR_KEY_ID` then `adduid`

### Passphrase issues
- If you forgot your passphrase, you'll need to create a new key
- Consider using a password manager to store your passphrase securely

### GitHub Actions signing fails
- Verify all three secrets are set correctly (`GPG_PRIVATE_KEY`, `GPG_PASSPHRASE`, `GPG_KEY_ID`)
- Check that the private key export includes the full key (not just the public part)

## Security Best Practices

1. **Never commit your private key** to the repository
2. **Use a strong passphrase** for your GPG key
3. **Backup your private key** securely (encrypted storage, password manager)
4. **Set key expiration** if appropriate for your use case
5. **Revoke old keys** if compromised or no longer needed

## Terraform Registry Requirements

For your provider to be published on the Terraform Registry:

1. ✅ **GPG key added to GitHub** - Required for Terraform Registry to verify signatures
2. ✅ **GPG key added to Terraform Registry** - Required for publishing
3. ✅ **Releases signed** - Our GitHub Actions workflow automatically signs all binaries
4. ✅ **Signature files included** - `.asc` files are attached to each release

The Terraform Registry automatically:
- Fetches your GPG public key from GitHub
- Verifies signatures on release artifacts
- Shows verification status on your provider page

## Publishing to Terraform Registry

Once your GPG key is set up:

1. **Create a release** by pushing a tag:
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

2. **GitHub Actions will automatically**:
   - Build binaries for all platforms
   - Sign all binaries with your GPG key
   - Create a GitHub Release with signed artifacts

3. **Terraform Registry will automatically**:
   - Detect your GitHub release
   - Verify GPG signatures
   - Publish your provider

4. **Verify your provider**:
   - Go to: `https://registry.terraform.io/providers/davidshato93/httpx`
   - Check that releases show as verified

## Additional Resources

- [GitHub: Managing commit signature verification](https://docs.github.com/en/authentication/managing-commit-signature-verification)
- [Terraform Registry: Publishing Providers](https://developer.hashicorp.com/terraform/registry/providers/publishing)
- [Terraform Registry: GPG Keys Settings](https://registry.terraform.io/settings/gpg-keys)
- [GPG Best Practices](https://www.gnupg.org/documentation/manuals/gnupg/GPG-Configuration.html)

