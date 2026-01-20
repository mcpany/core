# Config Validation Playground

## Overview

The Config Validation Playground is a dedicated UI tool that allows developers and operators to validate their `config.yaml` or JSON configuration files against the server's schema without needing to reload the server or run CLI commands.

## Features

- **Real-time Validation**: Paste your configuration and get immediate feedback.
- **Schema Compliance**: Validates against the exact JSON schema generated from the server's internal Protocol Buffers definitions.
- **Visual Feedback**: Clear success/error indicators with detailed error messages.

## Usage

1.  Navigate to **Config Validator** in the sidebar.
2.  Paste your configuration content (YAML or JSON) into the editor on the left.
3.  Click **Validate Configuration**.
4.  View the results in the right pane.

## API Endpoint

The feature exposes a REST API endpoint:

-   `POST /api/v1/config/validate`
-   **Body**: `{"content": "..."}`
-   **Response**: `{"valid": boolean, "errors": ["..."]}`

![Config Validator UI](screenshot.png)
