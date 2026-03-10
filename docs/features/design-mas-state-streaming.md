# Design Doc: Real-time MAS State Streaming
**Status:** Draft
**Created:** 2026-03-10

## 1. Context and Scope
As agent workflows evolve from sequential turns to parallel swarms (e.g., Anthropic's parallel code review), the "Shared KV Store" (Blackboard) becomes a bottleneck. Agents currently "poll" the blackboard for updates, leading to latency and race conditions. Parallel agents need a way to receive immediate notifications when shared state changes so they can adjust their reasoning in real-time.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a Pub/Sub mechanism for the Shared KV Store.
    * Enable agents to subscribe to specific keys or "Intent Scopes."
    * Support "Streaming Context Updates" where partial tool outputs are broadcast to the swarm.
    * Maintain backwards compatibility with the existing KV CRUD API.
* **Non-Goals:**
    * Building a full message broker (e.g., Kafka). We will use a lightweight in-memory or WebSocket-based bus.
    * Persistent history of all state changes (only the "current" state is guaranteed).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Orchestrator.
* **Primary Goal:** 3 agents performing parallel tasks (Backend, Frontend, Docs) stay synced on a shared "API Contract."
* **The Happy Path (Tasks):**
    1. Agent "Backend" updates the key `api_spec` in the Blackboard.
    2. MCP Any broadcasts a `STATE_UPDATED` event to all subscribed agents.
    3. Agents "Frontend" and "Docs" receive the event via their open WebSocket/SSE connection.
    4. "Frontend" immediately updates its implementation plan without needing a new turn or manual polling.

## 4. Design & Architecture
* **System Flow:**
    * **Producer**: An agent calling `set_key(key, value)`.
    * **Hub**: MCP Any's `StateBus` middleware.
    * **Consumer**: Parallel agents with an active subscription.
* **APIs / Interfaces:**
    * `subscribe(pattern)`: Tool to register interest in key changes.
    * `X-MCP-Event-Stream`: SSE (Server-Sent Events) endpoint for receiving updates.
* **Data Storage/State:**
    * In-memory subscription map (Key -> List of SIDs).

## 5. Alternatives Considered
* **Database Triggers**: Rejected as too heavy for local developer environments.
* **Long Polling**: Rejected due to inefficiency and high latency.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Subscriptions are bound by the same "Agent-Aware Isolation" as the KV Store. An agent can only subscribe to keys it has permission to read.
* **Observability:** Track "Broadcast Latency" to ensure the swarm stays synchronized.

## 7. Evolutionary Changelog
* **2026-03-10:** Initial Document Creation.
