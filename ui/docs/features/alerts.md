# Alerts & Notifications

**Status:** Implemented

## Goal
Acknowledge and investigate system anomalies. The Alerts system centralizes notifications about service health, high error rates, and performance degradation.

## Usage Guide

### 1. View Incidents
Navigate to `/alerts`.
- **Incidents Tab**: Shows active and resolved alert instances.
- **Severity**: Critical (Red), Warning (Yellow), Info (Blue).

![Alerts List](screenshots/alerts_list.png)

### 2. Manage Rules
Navigate to the **Alert Rules** tab to configure monitoring conditions.
- **Create Rule**: Click "New Alert Rule" to define conditions (e.g., `cpu_usage > 90` for `5m`).
- **Manage**: View, enable/disable, or delete existing rules.

![Alert Rules](screenshots/alerts_rules.png)

### 3. Investigate
View alert details including severity, status, and service origin. Use the actions menu to acknowledge or resolve alerts.
