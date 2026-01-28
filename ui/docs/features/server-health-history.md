# Server Health History

MCP Any now includes a visual timeline of service health directly in the dashboard.

## Overview

The System Health widget has been enhanced to show a historical timeline of each service's status over the last 10 minutes (configurable).

![Server Health History](../screenshots/dashboard_overview.png)

## Features

-   **Visual Timeline**: A heatmap-style visualization shows the status of each service over time.
    -   <span style="color: green">Green</span>: Healthy
    -   <span style="color: orange">Amber</span>: Degraded
    -   <span style="color: red">Red</span>: Unhealthy/Error
    -   <span style="color: gray">Gray</span>: Inactive
-   **Client-Side Persistence**: Health history is persisted in your browser's local storage, allowing you to see trends even after refreshing the page.
-   **Real-time Updates**: The timeline updates in real-time as the dashboard polls for service status.
-   **Tooltip Details**: Hover over any point in the timeline to see the exact status and timestamp.

## Usage

Navigate to the **Dashboard** to view the System Health widget. The history is automatically collected and displayed.
