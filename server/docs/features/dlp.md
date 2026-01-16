# Data Loss Prevention (DLP) Middleware

The Data Loss Prevention (DLP) middleware scans and redacts sensitive information (PII) from both request arguments (inputs) and result content (outputs).

## Overview

DLP is critical for preventing sensitive data leaks when interacting with LLMs. This middleware sits in the request/response path and automatically sanitizes data based on configured rules.

## Features

- **Input Redaction**: Scans arguments in `CallToolRequest` for PII.
- **Output Redaction**: Scans text content in `CallToolResult` for PII.
- **Default Patterns**: Automatically redacts Credit Card numbers, SSNs, and Email addresses.
- **Custom Patterns**: Support for additional regex patterns.

## Configuration

To enable DLP, add the `dlp` section to your configuration:

```yaml
dlp:
  enabled: true
  # Custom regex patterns to redact in addition to defaults
  custom_patterns:
    - "SECRET-[A-Z0-9]+"
```

### Behavior

- **Default Patterns**: The following patterns are always enabled when DLP is on:
  - **Credit Cards**: Matches common credit card formats.
  - **SSN**: Matches Social Security Numbers.
  - **Email**: Matches email addresses.
- **Replacement**: All matched sensitive data is replaced with `***REDACTED***`.

## Implementation

The middleware is implemented in `server/pkg/middleware/dlp.go` and `server/pkg/middleware/redactor.go`. It uses regex-based replacement to sanitize data before it reaches the tool (for inputs) or before it returns to the client (for outputs).
