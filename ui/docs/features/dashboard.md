# System Dashboard

**Status:** Implemented

## Goal
Assess the overall health and performance of the MCP ecosystem. The Dashboard serves as the landing page, providing immediate visibility into key metrics and system status.

## Usage Guide

### 1. Overview
Navigate to `/` (Home).
The dashboard is composed of three main sections:
- **KPI Cards**: Top-level metrics.
- **Charts**: Historical trends.
- **Service Status**: Breakdown of connected components.

![Dashboard Overview](screenshots/dashboard_overview.png)

### 2. Live Metrics
The dashboard updates in real-time.
- **Total Requests**: Aggregate count of all operations processed.
- **Active Services**: Number of services currently in "Healthy" state.
- **Error Rate**: Percentage of requests resulting in failures.

### 3. Drill Down
Clicking on any KPI card or Chart point will navigate to the **Live Logs** or **Traces** view, filtered to the relevant context (e.g., clicking on a spike in errors takes you to logs showing those errors).
