# Dynamic UI Hosting

The MCP Any server includes built-in support for hosting the "Dynamic UI" (the web-based management dashboard).

## Overview

The "Dynamic UI" is a React-based single-page application (SPA) located in the [`ui/`](../../../ui/) directory. It provides a real-time interface for:
- Managing upstream services.
- Inspecting tools and resources.
- Viewing live logs and traces.
- Configuring global settings.

## Server Integration

The server serves the compiled UI assets statically.

### Discovery Logic

The server looks for the UI build artifacts in the following order:
1. `./ui/out` (Next.js static export)
2. `./ui/dist` (Generic build output)
3. `./ui` (Source directory - *Blocked if package.json exists to prevent source leakage*)

### Routes

- **`/`**: Serves `index.html` (the SPA entry point).
- **`/ui/*`**: Serves static assets (JS, CSS, images).
- **Fallbacks**: Any unknown route that accepts `text/html` falls back to `index.html` to support client-side routing.

## Development

To run the UI in development mode, refer to the [UI README](../../../ui/README.md).
In production, the `Dockerfile` builds the UI and places it in `/app/ui/out`.
