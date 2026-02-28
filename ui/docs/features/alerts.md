# Alerts & Incidents Feature

**Date:** 2026-01-20
**Status:** Active

## Overview

The **Alerts & Incidents** feature provides a centralized console for monitoring system health, viewing active alerts, and managing incident response. It is designed to elevate the "Premium Enterprise" feel of the MCP Any platform.

## Key Capabilities

1.  **Dashboard Stats:** Real-time KPI cards showing Active Critical alerts, Warning counts, MTTR (Mean Time To Resolution), and Total Incident volume.
2.  **Alert Feed:** A sortable, filterable list of all alerts with color-coded severity badges (Critical, Warning, Info) and status indicators (Active, Acknowledged, Resolved).
3.  **Filtering:** Users can filter alerts by Severity, Status, or free-text search (Title, Message, Service).
4.  **Rule Management:** A "Create Alert Rule" dialog allows users to define new monitoring conditions. Backend storage (In-Memory) is fully implemented.
5.  **Health Webhooks:** The system supports configuring a global webhook URL to receive notifications when upstream service health status changes (e.g. Healthy -> Degraded).

## Implementation Details

-   **Code Location:**
    -   Backend: `server/pkg/alerts/manager.go` (In-memory storage and logic), `server/pkg/app/api_alerts.go` (HTTP Handlers).
    -   Frontend: `ui/src/app/alerts/page.tsx`, `ui/src/components/alerts/alert-list.tsx`.
-   **Route:** `/alerts`
-   **API Endpoints:**
    -   `GET /api/v1/alerts`: List all alerts.
    -   `POST /api/v1/alerts`: Create a new alert.
    -   `GET /api/v1/alerts/{id}`: Get alert details.
    -   `PATCH /api/v1/alerts/{id}`: Update alert status.
    -   `GET /api/v1/alerts/rules`: List all alert rules.
    -   `POST /api/v1/alerts/rules`: Create a new alert rule.
    -   `GET /api/v1/alerts/rules/{id}`: Get rule details.
    -   `PUT /api/v1/alerts/rules/{id}`: Update rule.
    -   `DELETE /api/v1/alerts/rules/{id}`: Delete rule.
-   **Components:**
    -   `AlertsPage`: Main container layout.
    -   `AlertList`: The data table component with filtering logic, connected to the backend API.
    -   `AlertStats`: Top-level metrics.
    -   `CreateRuleDialog`: Configuration form.

## Verification

The feature is integrated with the backend `AlertsManager` for Alert and Rule CRUD operations using in-memory storage.
