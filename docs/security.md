# 🔒 Securing Your MCP Any Server

This guide explains how to secure your MCP Any server to prevent unauthorized access.

## API Key Authentication

The primary method for securing the MCP Any server is by using a server-wide API key. When an API key is configured, all incoming requests must include a valid `X-API-Key` header to be processed.

### Configuration

To enable API key authentication, add the `api_key` field to the `global_settings` section of your `config.yaml` file:

```yaml
global_settings:
  api_key: "your-super-secret-and-long-api-key"
```

**Important Security Considerations:**

*   **Key Length:** The API key **must** be at least 16 characters long. The server will fail to start if the key is too short.
*   **Key Strength:** Use a strong, randomly generated key. Avoid using common phrases or easily guessable strings.
*   **Secret Management:** For production environments, it is highly recommended to source your API key from a secure location, such as an environment variable or a secret management system, rather than hardcoding it in the configuration file.

### Client Requests

When the server is secured, clients must include the API key in the `X-API-Key` header of every request.

Here is an example of how to make a `tools/list` request using `curl`:

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-super-secret-and-long-api-key" \
  -d '{"jsonrpc": "2.0", "method": "tools/list", "id": 1}' \
  http://localhost:50050
```

If the `X-API-Key` header is missing, or if the key is incorrect, the server will reject the request with a `401 Unauthorized` status code.
