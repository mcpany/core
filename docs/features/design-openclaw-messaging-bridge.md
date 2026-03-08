# Design Doc: OpenClaw Messaging Bridge
**Status:** Draft
**Created:** 2026-03-05

## 1. Context and Scope
OpenClaw (formerly Clawdbot) has emerged as a dominant autonomous agent framework, primarily interacting with users through messaging platforms like WhatsApp, Telegram, and Slack. Currently, OpenClaw instances are often isolated and require complex setup to interact with wider MCP tool ecosystems. MCP Any needs to provide a native "Messaging Bridge" that allows OpenClaw-style triggers to be mapped directly to MCP tool calls, and allows MCP-native agents to communicate with OpenClaw agents via messaging webhooks.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide native adapters for common messaging webhooks (WhatsApp, Telegram, Slack).
    * Map incoming messages to specific MCP tool executions.
    * Support OpenClaw's "heartbeat" pattern for autonomous, scheduled tool calls.
    * Implement a "Message-Bound Buffer" to handle intermittent connectivity of local OpenClaw instances.
* **Non-Goals:**
    * Replacing the OpenClaw agent logic itself.
    * Building a full-featured chat application (MCP Any remains a protocol gateway).

## 3. Critical User Journey (CUJ)
* **User Persona:** Mobile AI Power User.
* **Primary Goal:** Trigger a local coding task on their home machine via Telegram while commuting.
* **The Happy Path (Tasks):**
    1. User sends "/run-tests" to their OpenClaw bot on Telegram.
    2. Telegram sends a webhook to MCP Any's `Messaging Bridge`.
    3. MCP Any authenticates the request and maps it to the `local_shell_execute(cmd="make test")` tool.
    4. The tool executes on the local machine; the result is captured by MCP Any.
    5. MCP Any sends the test results back to the user via the Telegram API.

## 4. Design & Architecture
* **System Flow:**
    `Messaging Platform (Telegram/Slack) -> MCP Any Messaging Bridge -> Auth/Policy Filter -> MCP Tool Registry -> Local/Remote MCP Server`
* **APIs / Interfaces:**
    * `POST /webhooks/telegram/:token`: Endpoint for Telegram bot updates.
    * `POST /webhooks/slack/:id`: Endpoint for Slack Event API.
    * Internal Service: `MessagingAdapterService` for normalizing different platform payloads.
* **Data Storage/State:**
    * Message history and pending responses are stored in the `A2A Stateful Residency` (Stateful Buffer).

## 5. Alternatives Considered
* **Direct OpenClaw-to-MCP Integration**: Requiring OpenClaw to implement full MCP client logic. *Rejected* as it adds complexity to the agent and lacks centralized security/observability provided by MCP Any.
* **Generic Webhook Tool**: Using a generic "Webhook-to-Tool" mapping. *Rejected* because messaging platforms have specific auth and message formatting requirements that benefit from native handling.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All incoming webhooks must be validated using platform-specific signatures (e.g., Slack's `X-Slack-Signature`). Access is restricted via the Policy Firewall based on the sender's verified identity.
* **Observability:** Message-to-Tool mappings are logged in the Trace ID timeline, allowing users to see exactly which message triggered which tool call.

## 7. Evolutionary Changelog
* **2026-03-05:** Initial Document Creation.
