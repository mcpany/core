# Actionable Configuration Errors

MCP Any includes a robust configuration validation system designed to provide **Actionable Errors**. Instead of opaque error messages, the server analyzes configuration issues and provides specific suggestions to fix them.

## Overview

When the server starts, it validates the loaded configuration (`config.yaml`). If an error is detected, the system attempts to wrap it in an `ActionableError`, which includes:
1.  **Context**: What went wrong.
2.  **Suggestion**: A specific step the user can take to resolve the issue.

## Features

### 1. Environment Variable Validation
The server checks for required environment variables. If one is missing, it suggests:
*   Adding it to the `.env` file.
*   Exporting it in the shell.
*   **Fuzzy Matching**: If you have a typo (e.g., `OPENAI_API_KEYY` instead of `OPENAI_API_KEY`), the error message will suggest the correct variable name.

**Example Error:**
```text
Error: api key secret validation failed
Suggestion: Environment variable 'OPENAI_API_KEY' is missing.
Did you mean 'OPENAI_API_KEYY'?
Check your .env file or export the variable.
```

### 2. Path & Command Validation
For local command-line tools (`stdio`), the server verifies that the executable path exists and is executable.

**Example Error:**
```text
Error: command_line_service command validation failed
Suggestion: The command '/usr/bin/phython3' was not found.
Did you mean '/usr/bin/python3'?
```

### 3. URL Validation
For HTTP and WebSocket services, URLs are validated for correct scheme (`http://`, `https://`, `ws://`) and format.

**Example Error:**
```text
Error: http_service address validation failed
Suggestion: Address must start with 'http://' or 'https://'.
Found: 'api.weather.com'
```

### 4. Whitespace Detection
The validator detects invisible whitespace characters that often creep in when copying configurations from the web.

**Example Error:**
```text
Error: invalid configuration
Suggestion: The value for 'api_key' contains leading/trailing whitespace.
Please check your config file and remove hidden characters.
```

## Developer Guide

If you are extending MCP Any, you can use the `ActionableError` type in `pkg/config/errors.go` to provide better error messages in your own components.

```go
import "github.com/mcpany/core/server/pkg/config"

func validate() error {
    if somethingWrong {
        return &config.ActionableError{
            Err: fmt.Errorf("validation failed"),
            Suggestion: "Try setting X to Y.",
        }
    }
    return nil
}
```
