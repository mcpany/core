# Alerts & Notifications

**Status:** Implemented

## Goal
Acknowledge and investigate system anomalies. The Alerts system centralizes notifications about service health, high error rates, and performance degradation.

## Usage Guide

### 1. View Alerts
Navigate to `/alerts`.
- **Active**: Current unresolved issues.
- **History**: Past resolved alerts.
- **Severity**: Critical (Red), Warning (Yellow), Info (Blue).

![Alerts List](screenshots/alerts_list.png)

### 2. Investigate
Click on an alert to view context. The system attempts to link you to the relevant **Log** or **Trace** time window.
