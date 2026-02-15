# Alerts & Notifications

**Status:** Implemented

## Goal
Acknowledge and investigate system anomalies. The Alerts system centralizes notifications about service health, high error rates, and performance degradation.

## Usage Guide

### 1. View Alerts
Navigate to `/alerts`. The default view shows all active alerts. You can filter the list using the dropdown menus:
- **Severity**: Filter by Critical, Warning, or Info.
- **Status**: Filter by Active, Acknowledged, or Resolved.

![Alerts List](screenshots/alerts_list.png)

### 2. Investigate
View alert details including severity, status, and service origin. Use the actions menu (three dots) on each row to:
- **Copy Alert ID**: For reference.
- **Acknowledge**: Mark an alert as under investigation.
- **Resolve**: Mark an alert as fixed.
