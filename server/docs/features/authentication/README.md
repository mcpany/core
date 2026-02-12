# Authentication

The `mcpany` server supports flexible authentication mechanisms.
 for both incoming requests (securing your MCP server) and outgoing requests (authenticating with upstream services). These are configured **per upstream service**.

## Configuration

Incoming authentication is configured under `authentication`. Outgoing authentication is configured under `upstream_auth`.

### Incoming Authentication

To secure access to a specific service exposed by MCP Any, you can use several authentication methods:

#### 1. API Key

```yaml
upstream_services:
  - name: "secure-service-apikey"
    authentication:
      api_key:
        param_name: "X-Mcp-Api-Key"
        in: "HEADER"
        verification_value: "my-secret-key"
```

#### 2. Basic Authentication

```yaml
upstream_services:
  - name: "secure-service-basic"
    authentication:
      basic_auth:
        username: "admin"
        password_hash: "$2a$12$..." # Bcrypt hash
```

#### 3. Trusted Header

Useful when running behind a proxy like Nginx or Cloudflare Access.

```yaml
upstream_services:
  - name: "secure-service-header"
    authentication:
      trusted_header:
        header_name: "X-Forwarded-User"
        # Optional: enforce a specific value
        # header_value: "expected-secret"
```

#### 4. OIDC / OAuth2

Validate JWT tokens from an identity provider (e.g., Google, Auth0).

```yaml
upstream_services:
  - name: "secure-service-oidc"
    authentication:
      oidc:
        issuer: "https://accounts.google.com"
        audience: ["my-client-id"]
```

### Outgoing Authentication

To authenticate with an upstream service:

```yaml
upstream_services:
  - name: "secure-upstream"
    upstream_auth:
      bearer_token:
        token:
          environment_variable: "UPSTREAM_TOKEN"
    http_service:
      address: "https://api.secure.com"
```

## Use Case

**Incoming**: You want to prevent unauthorized users from calling tool X.
**Outgoing**: Upstream API Y requires an API key or an OAuth token.

Clients calling `secure-service` must provide the configured authentication (e.g., adding `X-Mcp-Api-Key` header).

## Real World Example: IPInfo

This example demonstrates how to configure an upstream service (`ipinfo.io`) that requires an API key, load that key from an environment variable, and verify it using the Gemini CLI.

### 1. Prerequisite: Get an API Key

1.  Sign up at [ipinfo.io](https://ipinfo.io/signup).
2.  Copy your access token from the dashboard.

### 2. Configuration

Create a file named `config.yaml` with the following content. We use `${IPINFO_API_KEY}` to load the access token securely from the environment.

```yaml
upstream_services:
  - name: "ipinfo"
    http_service:
      address: "https://ipinfo.io"
      tools:
        - name: "get_ip_info"
          ignore_arguments: true
          http:
             endpoint_path: "/json"
             method: "GET"
    upstream_auth:
      bearer_token:
        token:
          environment_variable: "IPINFO_API_KEY"
```

### 3. Run the Server

Set the environment variable and start the server:

```bash
export IPINFO_API_KEY="your_actual_token_here"
# Assuming you have the mcp-any binary built
./bin/mcp-any --config config.yaml
```

### 4. Verification with Gemini CLI

We can use the `@google/gemini-cli` to verify that the tool works effectively.

1.  **Install Gemini CLI** (if not already installed):
    ```bash
    npm install -g @google/gemini-cli
    ```
    *Note: You can also use `npx -y @google/gemini-cli` to run it without installing globaly.*

2.  **Authenticate Gemini CLI**:
    You need a Gemini API Key from [AI Studio](https://aistudio.google.com/).
    ```bash
    export GEMINI_API_KEY="your_gemini_api_key"
    ```

3.  **Connect and Test**:
    Add the local MCP server to Gemini CLI and ask it to use the tool.

    ```bash
    # Add the local server (assuming default port 8080)
    npx -y @google/gemini-cli mcp add --transport http mcp-server http://localhost:8080/mcp/v1

    # Ask a question that triggers the tool
    npx -y @google/gemini-cli -p "What is my IP address info?"
    ```

    **Expected Output**:
    The CLI should show the tool call `ipinfo` -> `get_ip_info` and then print the JSON response containing your IP details.
