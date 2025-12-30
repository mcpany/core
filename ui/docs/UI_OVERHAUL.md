# MCP Any UI Overhaul - Feature Documentation

## Overview

This document outlines the new "Apple/Ubiquiti" inspired UI for the MCP Any management console. It covers the dashboard, core management features, and advanced configurations.

## 1. Dashboard & Observability

The dashboard provides a high-level overview of the system's health and performance.

*   **Real-time Metrics:** Displays Request/Sec, Active Services, Connected Tools, and Latency.
*   **Visual Aesthetics:** Uses glassmorphism (backdrop blur), clean typography, and subtle shadows.

![Dashboard Audit](../.audit/ui/2025-12-30/dashboard.png)

## 2. Core Management

### Services
Manage upstream services (HTTP, gRPC, MCP, CMD).
*   **List View:** See all registered services with their status and version.
*   **Toggle:** One-click enable/disable.
*   **Edit:** Configure service details via a side sheet.

![Services Audit](../.audit/ui/2025-12-30/services.png)

### Tools
View and manage tools exposed by upstream services.
*   **List View:** Filterable list of tools.
*   **Status:** Enable/Disable individual tools.

![Tools Audit](../.audit/ui/2025-12-30/tools.png)

### Resources
Manage static and dynamic resources.
*   **Control:** Toggle access to specific resources.

![Resources Audit](../.audit/ui/2025-12-30/resources.png)

### Prompts
Manage prompt templates.
*   **Templates:** View available prompts and their arguments.

![Prompts Audit](../.audit/ui/2025-12-30/prompts.png)

## 3. Profiles
Manage execution profiles for different environments (Dev, Prod, Debug).
*   **Card Layout:** Quick overview of profile settings.
*   **Management:** Create, Edit, Delete profiles.

![Profiles Audit](../.audit/ui/2025-12-30/profiles.png)

## 4. Advanced Features

### Middleware
Visual pipeline editor for request processing.
*   **Drag & Drop:** Reorder middleware easily.
*   **Visualization:** See the flow of a request from ingress to service.

![Middleware Audit](../.audit/ui/2025-12-30/middleware.png)

### Webhooks
Configure outbound webhooks for system events.
*   **Management:** Add new webhooks, test delivery, and view status.

![Webhooks Audit](../.audit/ui/2025-12-30/webhooks.png)
