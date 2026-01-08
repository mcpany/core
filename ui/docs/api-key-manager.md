# API Key Manager

The API Key Manager is a secure vault for managing credentials used by upstream services and tools within MCP Any.

## Features

- **Secure Storage**: Secrets are stored securely (simulated in this version via in-memory mock storage on the server, to be replaced by backend vault).
- **Visibility Control**: Secrets are masked by default. Toggle visibility to view or copy.
- **Provider Metadata**: Tag secrets with their provider (OpenAI, AWS, etc.) for better organization.
- **Clipboard Integration**: One-click copy functionality.

## Usage

1.  Navigate to **Settings > Secrets & Keys**.
2.  Click **Add Secret**.
3.  Select a provider, enter a friendly name, the environment variable key, and the secret value.
4.  Click **Save Secret**.

![API Key Manager](api_key_manager.png)
