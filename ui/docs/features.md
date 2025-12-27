# MCP Any UI Features

This document describes the new features implemented in the MCP Any UI Overhaul.

## Design Philosophy

The new UI follows a "Premium" design aesthetic inspired by Apple and Ubiquiti/Unifi. It features:
- **Clean Lines & Whitespace:** Generous spacing to reduce cognitive load.
- **Glassmorphism:** Subtle translucency and blur effects (`backdrop-blur-xl`) on cards and sidebars.
- **High-Contrast Typography:** Using the Inter font family for readability and a modern look.
- **Interactive Elements:** Hover effects, smooth transitions, and immediate feedback (toasts).

## Features

### 1. Dashboard (`/`)
The dashboard provides a high-level overview of the MCP infrastructure.
- **Real-time Metrics:** Visualizes "Requests per Second" using an area chart.
- **Key Indicators:** Cards for Total Requests, Active Services, Latency, and Error Rates with trend indicators.
- **Service Status:** A live list of connected services with uptime and status badges.

![Dashboard Audit](../.audit/ui/2025-05-20/dashboard.png)

### 2. Services Management (`/services`)
Manage upstream MCP services.
- **List View:** Displays services with their type, version, and status.
- **Toggle:** One-click enable/disable switch for each service.
- **CRUD Operations:**
    - **Add Service:** Form to register new HTTP, gRPC, or Command Line services.
    - **Edit Service:** Modify existing configurations.
    - **Delete:** Remove a service (via dropdown menu).

![Services Audit](../.audit/ui/2025-05-20/services.png)

### 3. Resource Management
Dedicated views for different MCP primitives.

#### Tools (`/tools`)
- List all discoverable tools from connected services.
- Toggle availability of individual tools.
- Search/Filter by name.

![Tools Audit](../.audit/ui/2025-05-20/tools.png)

#### Resources (`/resources`)
- View exposed file paths, database tables, and other resources.
- Toggle access to specific resources.

![Resources Audit](../.audit/ui/2025-05-20/resources.png)

#### Prompts (`/prompts`)
- Manage reusable prompts and templates.
- View arguments and descriptions.

![Prompts Audit](../.audit/ui/2025-05-20/prompts.png)

### 4. Advanced Features

#### Middleware (`/middleware`)
- Visual pipeline editor for global request processing.
- Reorder middleware (visualized order).
- Toggle specific middleware components (Auth, Rate Limiting, Logging).

![Middleware Audit](../.audit/ui/2025-05-20/middleware.png)

#### Webhooks (`/webhooks`)
- Configure external event notifications.
- Add new webhook endpoints.
- List active subscriptions.

![Webhooks Audit](../.audit/ui/2025-05-20/webhooks.png)
