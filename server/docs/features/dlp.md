# Data Loss Prevention (DLP) Middleware

The Data Loss Prevention (DLP) middleware scans and redacts sensitive information (PII) from both request arguments (inputs) and result content (outputs).

## Overview

DLP is critical for preventing sensitive data leaks when interacting with LLMs. This middleware sits in the request/response path and automatically sanitizes data based on configured rules.

## Features

- **Input Redaction**: Scans arguments in `CallToolRequest` for PII.
- **Output Redaction**: Scans text content in `CallToolResult` for PII.
- **Configurable Rules**: Define custom regex patterns to look for.
- **Default Protection**: Automatically attempts to detect and redact Email addresses, Credit Card numbers, and Social Security Numbers.

## Configuration

To enable DLP, add the `dlp` section to your configuration:

```yaml
dlp:
  enabled: true
  # List of additional regex patterns to redact.
  # Matches will be replaced with "***REDACTED***".
  custom_patterns:
    - "SECRET-[A-Z0-9]+"
    - "API_KEY_[a-zA-Z0-9]+"
```

## Implementation

The middleware is implemented in `server/pkg/middleware/dlp.go`. It uses regex-based replacement to sanitize data before it reaches the tool (for inputs) or before it returns to the client (for outputs).
