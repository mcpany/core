# Deep Connection Validation

MCP Any now supports "Deep Connection Validation" during the service registration process. This feature goes beyond static configuration checks and actively attempts to connect to the upstream service to verify its availability and correctness.

## Overview

When registering or editing an upstream service, the "Test Connection" button now performs two levels of validation:

1.  **Static Schema Validation**: Checks if the configuration matches the required schema (e.g., required fields, valid JSON/YAML).
2.  **Deep Connection Check**: actively attempts to start the service (for Stdio/Command types) or reach the endpoint (for HTTP/SSE types).

## Addressing User Pain Points

This feature directly addresses common configuration issues such as:
- **Filesystem/Command Errors**: Providing a path that does not exist or a command that fails to start. Previously, the server would crash or the service would silently fail. Now, the validation step catches the process exit code and stderr output.
- **Network Reachability**: Verifying that an HTTP endpoint is actually reachable before saving the configuration.

## How it Works

The frontend calls the `/api/v1/services/validate` endpoint with a `check_connection=true` query parameter.

The backend then:
1.  Temporarily instantiates the upstream service using the provided configuration.
2.  Attempts to `Register` the service with a temporary, isolated environment.
3.  For **Stdio** services, this involves spawning the process and performing the initial handshake (ListTools). If the process exits immediately (e.g., due to a bad argument), the error is captured.
4.  The temporary service is then immediately shut down.

## UI Experience

- **Valid Configuration**: User sees a "Validation Successful" message.
- **Invalid Configuration**: User sees a "Validation Failed" message with detailed error output (e.g., "process exited with error: exit status 1. Stderr: path /bad/path does not exist").

![Connection Validation Screenshot](../screenshots/connection_validation.png)
