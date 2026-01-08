# SSO Integration

MCP Any supports Single Sign-On (SSO) integration to secure your MCP server.

## Configuration

```yaml
sso:
  enabled: true
  idp_url: "https://your-idp.example.com"
```

## Features

- **Identity Header Support**: Trusted proxy pattern via `X-MCP-Identity`.
- **Bearer Token Validation**: Validates `Authorization: Bearer <token>` headers.
- **Redirects**: Redirects unauthenticated users to the IDP login URL.
