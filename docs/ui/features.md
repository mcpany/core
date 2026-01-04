# MCP Any UI Overhaul - Feature Documentation

## Overview

This update introduces a comprehensive overhaul of the MCP Any management console, focusing on a modern, "Premium" aesthetic and enhanced functionality.

## Features

### 1. Dashboard & Observability

**Objective:** Visualize key performance indicators and system health.

*   **Real-time Metrics:** Displays Total Requests, Average Latency, Active Services, and Error Rates.
*   **Service Health:** Live status (Healthy, Degraded, Unhealthy) for all connected upstream services.
*   **Request Volume:** Interactive area chart showing traffic over the last 24 hours.

### 2. Service Management (CRUD)

**Objective:** Full lifecycle management of upstream services.

*   **List View:** See all registered services with their type (HTTP, gRPC, MCP, etc.) and status.
*   **Toggle:** One-click enable/disable.
*   **Registration:** Wizard-like interface for registering new services (HTTP, gRPC, Command Line, MCP).
*   **Edit/Delete:** Modify configurations or remove services.

### 3. Tool Management

**Objective:** Inspect and control tools exposed by services.

*   **Discovery:** Automatically lists tools from registered services.
*   **Inspector:** View input schemas (JSON) and tool descriptions.
*   **Control:** Enable or disable specific tools globally.

### 4. Resource & Prompt Management

**Objective:** Manage static resources and prompt templates.

*   **Resources:** View available resources (files, logs, data) and their MIME types.
*   **Prompts:** Manage prompt templates and their arguments.

### 5. Execution Profiles

**Objective:** Manage environment configurations.

*   **Profiles:** Create distinct profiles (e.g., Dev, Prod, Staging).
*   **Environment Variables:** JSON-based editor for setting environment variables per profile.

### 6. Advanced Features

*   **Middleware Pipeline:** Drag-and-drop interface to reorder request processing middleware (Auth, Rate Limiting, etc.).
*   **Webhooks:** Configure and test external webhooks for system events.

## Audit & Verification

All features have been verified via E2E tests.

### Screenshots

*(Screenshots are stored in `.audit/ui/` and `docs/ui/screenshots/`)*

*   **Dashboard:** `docs/ui/screenshots/dashboard.png`
*   **Services:** `docs/ui/screenshots/services.png`
*   **Tools:** `docs/ui/screenshots/tools.png`
*   **Middleware:** `docs/ui/screenshots/middleware.png`

## Testing Strategy

*   **Unit Tests:** Component-level tests.
*   **E2E Tests:** Playwright tests covering all critical flows (CRUD, Toggles, Navigation).
*   **Visual Audit:** Automated screenshot generation for design verification.
