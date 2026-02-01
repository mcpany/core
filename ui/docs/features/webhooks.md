# Webhooks

**Status:** Implemented

## Goal

Set up external notifications for system events. Webhooks allow you to integrate MCP Any with third-party systems like Slack, PagerDuty, or custom pipelines.

## Usage Guide

### 1. Webhook List

Navigate to `/webhooks`. This page lists all configured outbound webhooks.

![Webhooks List](screenshots/webhooks_list.png)

### 2. Register New Webhook

1. Click **"Add Webhook"**.
2. Enter the **Target URL**.
3. Select the **Events** to subscribe to (e.g., `service.down`, `job.failed`).

![Create Webhook](screenshots/webhook_create_modal.png)

### 3. Test Delivery

Click the **"Test"** button on a webhook row to send a sample payload to the configured URL. The response status code (e.g., 200 OK) will be displayed ensuring connectivity.

## Technical Details

### Event Types

- `service.health_change`: Triggered when a service goes Up or Down.
- `tool.error`: Triggered when a tool execution fails.
- `audit.log`: Triggered for every audit event (high volume).

### Sample Payload

```json
{
  "event_id": "evt_12345",
  "type": "service.health_change",
  "timestamp": "2026-01-10T12:00:00Z",
  "data": {
    "service_id": "postgres-primary",
    "previous_status": "healthy",
    "new_status": "unhealthy",
    "reason": "Connection refused"
  }
}
```
