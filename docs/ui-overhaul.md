# MCP Any UI Overhaul - Feature Documentation

## Overview

This document outlines the features and design of the new "MCP Any" Management Console. The UI has been completely overhauled to provide a modern, enterprise-grade experience inspired by Apple and Ubiquiti aesthetics.

## Features

### 1. Dashboard & Observability

The dashboard provides a real-time overview of the system's health and performance.

*   **Key Metrics:** View Total Requests, Active Services, Average Latency, and Error Rates at a glance.
*   **Performance Graph:** A time-series chart visualizes request volume and latency trends.
*   **Service Status:** A quick list of active upstream services and their health status.

![Dashboard](../../.audit/ui/2025-12-23/dashboard.png)

### 2. Service Management

Manage your upstream MCP services with ease.

*   **List View:** See all connected services, their types (MCP, HTTP, gRPC), versions, and status.
*   **Quick Actions:** Enable/Disable services or edit their configurations directly from the list.

![Services](../../.audit/ui/2025-12-23/services.png)

### 3. Tool Discovery

Explore the tools exposed by your connected services.

*   **Tool Browser:** A searchable list of all available tools.
*   **Details:** View tool descriptions and schema requirements.

![Tools](../../.audit/ui/2025-12-23/tools.png)

### 4. Middleware Pipeline

Configure the global request processing pipeline visually.

*   **Visual Editor:** Drag-and-drop interface (mocked) to order middleware components.
*   **Pipeline Stages:** Manage Security, Rate Limiting, Logging, and Caching layers.

![Middleware](../../.audit/ui/2025-12-23/middleware.png)

### 5. Settings & Webhooks

Centralized configuration for the MCP Any server.

*   **General Settings:** Server naming and maintenance mode.
*   **Webhooks:** Configure global webhooks to receive notifications for system events (Service Up/Down, Errors).

![Settings](../../.audit/ui/2025-12-23/settings.png)

## Technical Details

*   **Framework:** Next.js 15 (React 18)
*   **Styling:** Tailwind CSS + Shadcn/UI
*   **Icons:** Lucide React
*   **Charts:** Recharts
*   **Testing:** Playwright (E2E), Jest/Vitest (Unit)

## Verification

This release has been verified with a full suite of E2E tests covering dashboard rendering, navigation, and core feature visibility. Audit screenshots are generated automatically during the test run.
