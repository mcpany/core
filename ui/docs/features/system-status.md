# System Status & Health History

MCP Any provides comprehensive visibility into the availability and health of your upstream services with a dedicated **System Status** page and 24-hour history retention.

## Overview

The System Status feature allows administrators to track service reliability over time, identifying patterns of instability or outages that might have occurred when they weren't watching the dashboard.

![System Status Page](../screenshots/system_status.png)

## Key Features

-   **24-Hour History**: The system now retains up to 24 hours of health check history for all services (persisted in browser storage).
-   **Visual Timeline**: A GitHub-contribution-style timeline visualizes the status of each service in 15-minute blocks.
    -   <span style="color: green">Green</span>: Fully Operational
    -   <span style="color: orange">Amber</span>: Degraded Performance (high latency)
    -   <span style="color: red">Red</span>: Outage / Error
-   **Uptime Calculation**: Automatically calculates and displays the uptime percentage for each service over the last 24 hours.
-   **Dedicated Status Page**: A full-page view accessible via the sidebar under "System Status".
-   **Overall System Health**: A top-level summary of the entire platform's health.

## Usage

1.  Click on **System Status** in the sidebar (under Platform).
2.  View the **Overall Status** banner to see if any critical issues exist.
3.  Review the list of services. Each card shows:
    -   Current Status
    -   24h Uptime Percentage
    -   Timeline Graph
4.  Hover over any block in the timeline to see detailed status, timestamp, and error counts for that period.

## Dashboard Integration

The dashboard **System Health** widget continues to show a brief 10-minute snapshot for quick monitoring, but now leverages the same 24-hour data store for consistency.
