# Server Health History

MCP Any now includes a visual timeline of service health directly in the dashboard.

## Overview

The System Health widget has been enhanced to show a historical timeline of each service's status.

![Server Health History](../screenshots/dashboard_overview.png)

## Features

-   **Visual Timeline**: A heatmap-style visualization shows the status of each service over time.
    -   <span style="color: green">Green</span>: Healthy
    -   <span style="color: orange">Amber</span>: Degraded
    -   <span style="color: red">Red</span>: Unhealthy/Error
    -   <span style="color: gray">Gray</span>: Inactive
-   **Server-Side Persistence**: Health history is stored in the server, ensuring that all users see the same consistent history timeline, regardless of when they opened the dashboard.
-   **Real-time Updates**: The timeline updates in real-time as the dashboard polls for service status.
-   **Tooltip Details**: Hover over any point in the timeline to see the exact status and timestamp.

## Usage

Navigate to the **Dashboard** to view the System Health widget. The history is automatically collected and displayed.

## Roadmap & Limitations

> **Note:** Currently, health history is stored **in-memory** on the server.

The following enhancements are planned (see [Project Roadmap](../../roadmap.md)):
-   **Database Persistence**: Persisting health history to the backend database (SQLite/Postgres) to retain history across server restarts.
-   **Uptime Reporting**: Generating availability reports based on long-term history.
