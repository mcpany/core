# Middleware Visualization

The Middleware Visualization UI provides an interactive way to view and manage the request processing pipeline.

## Overview

MCP Any uses a chain of middleware to handle cross-cutting concerns like authentication, rate limiting, and logging. This UI allows you to visualize the order of these middleware and dynamically enable/disable them.

## Features

- **Visual Pipeline**: See the flow of requests from top to bottom through the middleware chain.
- **Drag & Drop Reordering**: Change the order of middleware execution by dragging cards.
- **Toggle Control**: Enable or disable specific middleware components with a switch.
- **Configuration Access**: Quick access to settings for each middleware.

## Usage

Navigate to the **Middleware** section in the dashboard.
1. **Reorder**: Drag the handle on the left of a middleware card to move it up or down.
2. **Toggle**: Use the switch on the right to enable or disable the middleware.

## Implementation

The page is implemented in `ui/src/app/middleware/page.tsx`. It uses `@hello-pangea/dnd` for the drag-and-drop functionality.

**Note:** Currently, this feature is a client-side simulation. Changes made in the UI (reordering, toggling) update the local state for visualization purposes but do not yet persist to the server configuration.
