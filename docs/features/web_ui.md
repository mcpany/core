# Dynamic Web UI (Beta)

MCP Any includes a modern, dynamic web interface for monitoring and managing the server.

## Features

*   **Dashboard**: View real-time server health and metrics.
*   **Service Explorer**: Browse registered upstream services and their tools.
*   **Configuration Management**: View current configuration.
*   **Interactive Playground**: Test tools directly from the browser (Coming Soon).

## Accessing the UI

By default, the UI is available at `http://localhost:3000` when running the full stack, or served alongside the API if configured.

## Development

The UI is built with Next.js and located in the `ui/` directory.

To run the UI locally:

```bash
cd ui
npm install
npm run dev
```
