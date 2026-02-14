# Webhooks

**Status:** Prototype / In Progress

## Goal

Set up external notifications for system events. Webhooks allow you to integrate MCP Any with third-party systems like Slack, PagerDuty, or custom pipelines.

> **Note:** The UI currently demonstrates the configuration flow. Backend integration for dynamic webhook management is pending. Currently, webhooks are configured statically via the `config.yaml` (Global Settings).

## Usage Guide

### 1. Webhook List

Navigate to `/webhooks`. This page lists configured outbound webhooks (currently mock data).

![Webhooks List](screenshots/webhooks_list.png)

### 2. Register New Webhook

1. Click **"Add Webhook"**.
2. Enter the **Target URL**.
3. Select the **Events** to subscribe to (e.g., `service.down`, `job.failed`).

*(Changes in this UI are currently local-only and not persisted to the server)*

![Create Webhook](screenshots/webhook_create_modal.png)

### 3. Test Delivery

Click the **"Test"** button on a webhook row to simulate a payload delivery.

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
