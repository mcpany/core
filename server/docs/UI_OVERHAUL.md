# MCP Any UI Overhaul

**Version:** 2.1
**Date:** 2025-12-31
**Status:** Released

---

## Overview

The MCP Any management console has been overhauled to provide a modern, enterprise-grade experience. This update focuses on observability, ease of management, and advanced configuration capabilities, all wrapped in a premium, responsive design.

## Features

### 1. Dashboard & Observability

The new dashboard provides immediate visibility into the health and performance of your MCP infrastructure.

*   **Real-time Metrics:** key performance indicators including Request Volume, Active Services, Latency, and Active Users.
*   **Service Health:** Live status monitoring for all connected upstream services.

![Dashboard Preview](../.audit/ui/2025-12-31/dashboard.png)

### 2. Core Service Management

Manage your upstream services with precision.

*   **Service Listing:** View all registered services with their type, version, and status.
*   **One-click Toggles:** Enable or disable services instantly.
*   **Quick Editing:** Modify service configurations via a slide-over panel without leaving the context.

![Services Management](../.audit/ui/2025-12-31/services.png)

### 3. Tool Management & Inspection

Explore and manage tools exposed by your services.

*   **Tool Inspector:** View detailed schema and configuration for any tool.
*   **Status Control:** Enable or disable specific tools.
*   **Search & Filter:** Quickly find the tools you need.

![Tool Inspector](../.audit/ui/2025-12-31/tools.png)

### 4. Resources & Prompts

Manage static resources and prompt templates.

*   **Resource Inventory:** List available resources and their MIME types.
*   **Prompt Templates:** Manage standard prompts available to agents.

![Resources](../.audit/ui/2025-12-31/resources.png)
![Prompts](../.audit/ui/2025-12-31/prompts.png)

### 5. Profiles

Manage execution profiles for different environments (Dev, Prod, Debug).

![Profiles](../.audit/ui/2025-12-31/profiles.png)

### 6. Middleware Pipeline

Visual management of the request processing pipeline.

*   **Drag-and-Drop Ordering:** Easily reorder middleware components.
*   **Status Indicators:** Quickly see which middleware components are active.
*   **Visualization:** Visual representation of the request flow.

![Middleware Pipeline](../.audit/ui/2025-12-31/middleware.png)

### 7. Webhooks Configuration

Configure and test outbound webhooks for system events.

*   **Event Subscription:** Subscribe to specific events like `service.down` or `error.critical`.
*   **Status Management:** Enable or disable webhooks as needed.
*   **Testing:** Trigger test events to verify connectivity.

![Webhooks Configuration](../.audit/ui/2025-12-31/webhooks.png)

---

## Technical Details

*   **Framework:** Next.js with TypeScript
*   **Styling:** Tailwind CSS with `shadcn/ui` components
*   **Testing:** Comprehensive E2E testing with Playwright
