# Secrets Management

**Status:** Mock / Preview

## Goal
Securely store and manage sensitive information. The Secrets Vault enables you to inject API keys, passwords, and tokens into services without exposing them in plain text configuration files.

> **Note:** The current implementation in the UI is a **Mock** for demonstration purposes. Secrets added here are stored in-memory within the UI server and are **not** persisted or accessible by the backend MCP server. For production use, please configure secrets using environment variables as described in the Configuration Guide. Integration with real secret managers (Vault/AWS) is on the roadmap.

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

When configuring a service, you can reference the secret in environment variable fields.

**YAML Configuration:**
Use the `${secrets.KEY_NAME}` syntax in your `config.yaml` or `config.json` files.

**UI Configuration:**
The UI currently treats manual input as plain text. To use secrets in the UI, ensure they are loaded from the configuration file, or use the "Secrets" management interface to view existing secret references.
