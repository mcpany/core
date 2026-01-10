# Data Loss Prevention (DLP) Middleware

The Data Loss Prevention (DLP) middleware scans and redacts sensitive information (PII) from both request arguments (inputs) and result content (outputs).

## Overview

DLP is critical for preventing sensitive data leaks when interacting with LLMs. This middleware sits in the request/response path and automatically sanitizes data based on configured rules.

## Features

- **Input Redaction**: Scans arguments in `CallToolRequest` for PII.
- **Output Redaction**: Scans text content in `CallToolResult` for PII.
- **Configurable Rules**: Define what patterns to look for (e.g., Credit Card numbers, SSN, Email addresses).

## Configuration

To enable DLP, add the `dlp` section to your configuration:

```yaml
dlp:
  enabled: true
  rules:
    - name: "credit_card"
      pattern: "\\d{4}-\\d{4}-\\d{4}-\\d{4}"
      replacement: "[REDACTED_CC]"
    - name: "email"
      pattern: "[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}"
      replacement: "[REDACTED_EMAIL]"
```

## Implementation

The middleware is implemented in `server/pkg/middleware/dlp.go`. It uses regex-based replacement to sanitize data before it reaches the tool (for inputs) or before it returns to the client (for outputs).
