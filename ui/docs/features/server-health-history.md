# Server Health History & System Status

MCP Any provides comprehensive visibility into the availability and health of your upstream services.

## Overview

The System Health monitoring capabilities have been upgraded to include:
-   **Dashboard Widget**: A 10-minute "live" view of service status.
-   **System Status Page**: A dedicated page showing **24-hour history** for all services.
-   **Uptime Calculation**: Real-time observed uptime percentage.

![System Status Page](../screenshots/system_status_page.png)

## Dashboard Widget

The widget on the main dashboard provides a quick glance at the current health of all services.

-   **Visual Timeline**: Shows the last 10 minutes of health checks.
-   **Status Indicators**: Green (Healthy), Amber (Degraded), Red (Unhealthy).
-   **Real-time Updates**: Updates every 10 seconds.

## System Status Page

Access the **System Status** page from the sidebar to view detailed historical data.

### Features

-   **24-Hour Timeline**: A heat-map style visualization aggregating health checks into 15-minute blocks over the last 24 hours.
    -   Blocks are color-coded based on the *worst* status observed in that period.
    -   Hover over any block to see the time range and number of checks recorded.
-   **Observed Uptime**: Displays the percentage of successful health checks relative to total checks observed by your client.
-   **Overall System Status**: A banner summarizing the global state of the system (e.g., "All Systems Operational", "Critical Outage").

### Data Persistence

Health history is persisted in your browser's local storage (`localStorage`). This allows:
-   **Trend Analysis**: You can close the tab and return later; as long as the local storage isn't cleared, the history remains.
-   **Privacy**: Health data stays in your browser and is not sent to any third-party telemetry service.

## Usage

1.  Navigate to **System Status** in the sidebar.
2.  Review the overall system health banner.
3.  Inspect individual service cards for detailed timelines.
4.  Hover over timeline segments to investigate past incidents.
