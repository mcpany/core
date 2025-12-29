# API Key & Secrets Manager

## Overview

The API Key & Secrets Manager is a secure vault UI designed to manage credentials for upstream services within the MCP Any platform. It provides a centralized location to store, view, and manage sensitive information such as API keys, tokens, and other secrets.

## Features

-   **Secure Storage**: Secrets are stored securely (simulated in local storage for this demo).
-   **Visibility Control**: Secret values are masked by default and can be revealed with a click.
-   **Copy to Clipboard**: One-click copy functionality for easy usage.
-   **Provider Metadata**: Tag secrets with their associated provider (e.g., OpenAI, AWS).
-   **Search**: Quickly find secrets by name or key.

## Usage

1.  Navigate to **Settings** -> **Secrets & Keys**.
2.  Click **+ Add Secret**.
3.  Fill in the details:
    -   **Provider**: Select the service provider.
    -   **Friendly Name**: A descriptive name for the secret.
    -   **Key Name**: The environment variable name (e.g., `OPENAI_API_KEY`).
    -   **Secret Value**: The actual secret string.
4.  Click **Save Secret**.

## Implementation Details

The feature is implemented as a client-side component using React and shadcn/ui. It interacts with a mock backend service via the `apiClient` to simulate CRUD operations.

### Component Structure

-   `SecretsManager`: The main container component handling state and interaction.
-   `SecretItem`: Individual secret row with visibility toggle and actions.

### Testing

-   **Unit Tests**: Verified with Vitest in `ui/src/tests/components/secrets-manager.test.tsx`.
-   **E2E Tests**: User flow verified with Playwright in `ui/tests/secrets.spec.ts`.

## Screenshot

![API Key Manager](../../.audit/ui/2025-12-29/api_key_manager.png)
