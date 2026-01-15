# Alerts & Incidents Feature

**Date:** 2026-01-09
**Status:** Prototype / Mock

## Overview

The **Alerts & Incidents** feature provides a centralized console for monitoring system health, viewing active alerts, and managing incident response. It is designed to elevate the "Premium Enterprise" feel of the MCP Any platform.

**Note:** This feature is currently in a **Prototype** state. The UI is fully functional but it operates on mock data and is not yet connected to a backend alerting system.

## Key Capabilities

1.  **Dashboard Stats:** Real-time KPI cards showing Active Critical alerts, Warning counts, MTTR (Mean Time To Resolution), and Total Incident volume.
2.  **Alert Feed:** A sortable, filterable list of all alerts with color-coded severity badges (Critical, Warning, Info) and status indicators (Active, Acknowledged, Resolved).
3.  **Filtering:** Users can filter alerts by Severity, Status, or free-text search (Title, Message, Service).
4.  **Rule Management:** A "Create Alert Rule" dialog allows users to define new monitoring conditions using PromQL-like syntax.

## Implementation Details

-   **Route:** `/alerts`
-   **Components:**
    -   `AlertsPage`: Main container layout.
    -   `AlertList`: The data table component with filtering logic.
    -   `AlertStats`: Top-level metrics.
    -   `CreateRuleDialog`: Configuration form.
-   **Mock Data:** The current implementation uses static mock data for demonstration purposes.

## Verification

The feature has been verified with E2E tests using Playwright against the mock data.
