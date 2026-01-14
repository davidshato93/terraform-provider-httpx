# SSH Setup for davidshato93 GitHub Account

## Current Issue
Your SSH key is authenticated for `david-shato-sisense` account, but the repository belongs to `davidshato93`.

## Solution: Add SSH Key to davidshato93 Account

### Step 1: Get Your SSH Public Key
Run this command to see your public key:
```bash
cat ~/.ssh/id_ed25519.pub
# OR
cat ~/.ssh/id_rsa.pub
```

### Step 2: Add Key to GitHub
1. Go to: https://github.com/settings/keys
2. Click "New SSH key"
3. Title: `MacBook` (or any descriptive name)
4. Key: Paste your public key from Step 1
5. Click "Add SSH key"

### Step 3: Test Connection
```bash
ssh -T git@github.com
```
You should see: `Hi davidshato93! You've successfully authenticated...`

### Step 4: Push Code
```bash
cd /Users/davidshato/Documents/projects/self-projects/devops-tools/terraform-provider-httpx
git push -u origin main
```

## Alternative: Use HTTPS with Personal Access Token

If you prefer not to add SSH keys:
```bash
git remote set-url origin https://github.com/davidshato93/terraform-provider-httpx.git
git push -u origin main
# Username: davidshato93
# Password: <your Personal Access Token>
```

## Current Remote Configuration
- Remote URL: `git@github.com:davidshato93/terraform-provider-httpx.git`
- Branch: `main`
- Status: Ready to push (waiting for SSH authentication)

