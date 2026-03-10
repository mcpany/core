# Design Doc: Context Isolation Zones (CIZ)
**Status:** Draft
**Created:** 2026-03-10

## 1. Context and Scope
As AI agents become more autonomous and utilize complex subagent swarms (e.g., OpenClaw), the risk of "Context Bleeding"—where sensitive state or history from one task accidentally leaks into another—increases significantly. CIZ provides a cryptographically enforced boundary for every tool call and subagent session.

## 2. Goals & Non-Goals
* **Goals:**
    * Create ephemeral, isolated execution contexts for every tool call.
    * Enforce state-clearing between sequential subagent handoffs.
    * Provide cryptographic proof of isolation for audit logs.
* **Non-Goals:**
    * Replacing the LLM's own context window management.
    * Providing full OS-level virtualization (focus is on the MCP/Agent layer).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Orchestrator (e.g., CrewAI)
* **Primary Goal:** Ensure that a "Financial Analysis" subagent cannot see the context of a "Personal Email" subagent, even if they share the same MCP gateway.
* **The Happy Path (Tasks):**
    1. Orchestrator initiates a session with a unique `ZoneID`.
    2. MCP Any creates a temporary context buffer bound to that `ZoneID`.
    3. All tool calls within that session are restricted to the buffer.
    4. Upon session completion, MCP Any securely wipes the buffer and records an isolation attestation.

## 4. Design & Architecture
* **System Flow:**
    - **Zone Manager**: Tracks active `ZoneIDs` and their associated state buffers.
    - **Isolation Middleware**: Intercepts tool calls and injects zone-specific headers and filters.
    - **Ephemeral State Storage**: Uses encrypted, in-memory storage for zone-specific data.
* **APIs / Interfaces:**
    - `POST /v1/zones/create` -> Returns `zone_token`.
    - `mcp_zone_id` header required for all tool calls when CIZ is enabled.
* **Data Storage/State:** Encrypted RAM-backed KV store (cleared on TTL or explicit delete).

## 5. Alternatives Considered
* **Namespace-based Isolation**: Using simple string prefixes for state keys. *Rejected* as it is too easy to bypass with prompt injection.
* **Full Containerization**: Running a separate MCP Any instance per agent. *Rejected* due to excessive resource overhead.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** CIZ is the implementation of "Least Privilege" for agent context.
* **Observability:** The UI will feature a "Zone Map" showing active isolation boundaries and data flows.

## 7. Evolutionary Changelog
* **2026-03-10:** Initial Document Creation.
