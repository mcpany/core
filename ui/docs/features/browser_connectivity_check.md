# Browser-Side HTTP Connectivity Check

The **Browser-Side HTTP Connectivity Check** is a diagnostic tool integrated into the Connection Diagnostics dialog. It helps users distinguish between server-side networking issues and client-side (browser) connectivity issues.

## How it Works

When running diagnostics for an HTTP service:

1.  The tool attempts to fetch the service URL directly from the user's browser using `fetch` with `mode: 'no-cors'`.
2.  If the browser can reach the server (DNS resolution, TCP handshake, TLS handshake), the fetch succeeds (even if the response is opaque due to CORS).
3.  If the fetch fails, it indicates a network issue from the client's perspective (e.g., DNS failure, Firewall blocking, or invalid SSL certificate).

## Benefit

This feature is particularly useful in environments where:
- The MCP Any server is running in a container (Docker/K8s) and might have different network access than the user's browser.
- The user is trying to connect to a service on `localhost` relative to the server vs relative to the browser.
- There are corporate firewalls or VPNs involved.

## Screenshot

![Browser Connectivity Check](../../../.audit/ui/2026-01-22/browser_connectivity_check.png)
