# UI Features Documentation

## Overview

This document details the features implemented in the new "MCP Any" management console. The UI has been overhauled to follow a premium, enterprise-grade design philosophy (Unifi/Apple-esque), featuring clean lines, generous whitespace, and comprehensive observability.

## Features

### 1. Dashboard & Observability

The dashboard provides a high-level overview of the platform's health and performance.

*   **Key Metrics:** Real-time cards displaying Active Services, Total Tools, Request Rate, and Latency.
*   **Service Status:** A quick-glance table showing the health status and latency of all connected upstream services.
*   **Recent Activity:** A feed of recent operations and errors for immediate troubleshooting.

![Dashboard](.audits/ui/dashboard.png)

### 2. Service Management

Manage upstream MCP services with ease.

*   **List View:** View all configured services with their types (gRPC, HTTP, CMD) and current status.
*   **One-Click Toggle:** Enable or disable services instantly using the status switch.
*   **Add Service:** (Mocked) Interface to add new service connections.

![Services](.audits/ui/services.png)

### 3. Tool Management

Browse and search tools across the entire platform.

*   **Unified List:** Tools from all services are aggregated into a single searchable view.
*   **Search:** Filter tools by name, description, or service.

![Tools](.audits/ui/tools.png)

### 4. Resource & Prompt Management

Dedicated views for managing static resources and AI prompts.

*   **Resources:** List of data resources exposed by services.
*   **Prompts:** List of configured prompt templates.

![Resources](.audits/ui/resources.png)
![Prompts](.audits/ui/prompts.png)

### 5. Advanced Configuration

#### Profiles
Manage execution profiles for different environments (Development, Staging, Production). Configure debug modes and environment variables per profile.

![Profiles](.audits/ui/profiles.png)

#### Webhooks
Configure outgoing webhooks for system events. Manage target URLs and trigger conditions.

![Webhooks](.audits/ui/webhooks.png)

#### Middleware
Visual pipeline management for request processing. Reorder and toggle middleware components like Authentication, Rate Limiting, and Transformation.

![Middleware](.audits/ui/middleware.png)

## Technical Details

*   **Framework:** Next.js 15
*   **UI Library:** shadcn/ui + Tailwind CSS
*   **Icons:** Lucide React
*   **Testing:** Playwright E2E tests for verification.
