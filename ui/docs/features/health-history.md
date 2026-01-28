# Service Health History

## Overview

The Service Health History feature provides a visual timeline of the health status (UP/DOWN) and latency of your upstream services over time. This helps you diagnose stability issues and identify patterns of failure.

## Features

- **Visual Timeline**: Displays a bar chart of the last 100 health checks.
- **Color Coded**: Green bars indicate "UP" status, Red bars indicate "DOWN" status.
- **Latency Visualization**: The height of the green bars represents the latency (capped at 100ms for full height).
- **Detailed Tooltips**: Hover over any bar to see the exact timestamp, status, latency, and error message (if any).
- **Real-time Updates**: The timeline updates automatically every 5 seconds.

## Usage

1.  Navigate to the **Services** page.
2.  Click on any service to view its details.
3.  The **Stats** card on the Overview tab displays the "Health History" timeline.
