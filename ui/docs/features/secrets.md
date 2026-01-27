# Secrets Management

**Status:** Implemented

## Goal
Securely store and manage sensitive information. The Secrets Vault enables you to inject API keys, passwords, and tokens into services without exposing them in plain text configuration files.

## Usage Guide

### 1. View Secrets
Navigate to `/secrets` (or **Settings > Secrets**).
- **List**: Shows all stored keys.
- **Value**: Displayed as `*****` for security.

![Secrets List](screenshots/secrets_list.png)

### 2. Add Secret
1. Click **"Add Secret"**.
2. Enter the **Key** (e.g., `OPENAI_API_KEY`).
3. Enter the **Value** (e.g., `sk-...`).
4. Click **"Save"**.

![Create Secret](screenshots/secret_create_modal.png)

### 3. Usage in Services
When configuring a service, reference the secret using the `${secrets.KEY_NAME}` syntax in any environment variable field.
