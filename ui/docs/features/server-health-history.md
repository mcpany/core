# Traffic & Health History

MCP Any provides a real-time visual timeline of server traffic and availability directly in the dashboard.

## Overview

The System Uptime widget has been upgraded to a **Traffic & Health Monitor**, displaying the availability and request volume over the last hour.

![Traffic & Health](../screenshots/dashboard_overview.png)

## Features

-   **Real-time Bar Chart**: Visualizes traffic volume and health status minute-by-minute for the last hour.
-   **Availability Tracking**: Bars are color-coded based on the success rate of requests:
    -   <span style="color: green">Green</span>: Healthy (>99% Success)
    -   <span style="color: orange">Amber</span>: Degraded (>90% Success)
    -   <span style="color: red">Red</span>: Critical (<90% Success)
    -   <span style="color: gray">Gray</span>: No Traffic / Offline
-   **Real-time Polling**: The chart updates automatically every 30 seconds to show the latest data.
-   **Detailed Tooltips**: Hover over any bar to see:
    -   Time bucket
    -   Availability Percentage
    -   Total Request Count
    -   Error Count
-   **Backend Powered**: Data is sourced from the server's traffic history, ensuring accuracy across sessions.

## Usage

Navigate to the **Dashboard** to view the **Traffic & Health (Last Hour)** widget. It provides an immediate "at-a-glance" view of your system's recent performance and reliability.
