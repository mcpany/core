# Service Health History

## Overview
The Service Health History feature provides a visual timeline of the health status and latency of your upstream services. This allows operators to identify flapping services, latency spikes, and intermittent failures that instantaneous health checks might miss.

## Features
- **Latency Visualization**: An area chart showing the response time (latency) of health checks over time.
- **Error Overlay**: Periods where the service returned an error are highlighted in red.
- **Detailed Tooltips**: Hovering over the chart reveals the exact timestamp, latency, status, and any error message.
- **In-Memory Storage**: History is stored in an efficient in-memory ring buffer (last 100 checks), ensuring low overhead.

## Usage
1. Navigate to the **Services** page.
2. Click on a service to view its details.
3. Switch to the **Metrics** tab.
4. The **Health History** chart is displayed below the usage metrics.

## API
- `GET /api/v1/services/{name}/health-history`: Returns a JSON array of health check records.
