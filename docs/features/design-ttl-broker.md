# Design Doc: Thinking-Time Lease (TTL) Broker
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
The introduction of "Thinking Tools" and Reasoning-as-a-Service (RaaS) has led to "Cognitive Exhaustion" attacks, where agents enter infinite internal reasoning loops that consume parent mission tokens without progress. Following the CBS v1.0 standard, MCP Any needs a TTL Broker to manage granular, time-bound reasoning leases for all connected agents.

## 2. Goals & Non-Goals
* **Goals:**
    * Mediate "Thinking-Time" requests between agents and model providers.
    * Enforce strict temporal and token boundaries on internal reasoning cycles.
    * Automatically revoke capabilities if a reasoning loop exceeds its leased budget.
* **Non-Goals:**
    * Modify the internal weights of models.
    * Enforce prompt-length limits (handled by the Context Engine).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Performance Engineer
* **Primary Goal:** Prevent specialist agents from "squatting" on reasoning tokens.
* **The Happy Path (Tasks):**
    1. A subagent identifies a need for "Thinking Time" and sends a request to the TTL Broker.
    2. The TTL Broker checks the agent's current mission-root allocation.
    3. The Broker issues a `TTL-Token` with a 500ms / 1k token budget.
    4. The agent passes this token to the RaaS provider via MCP Any.
    5. MCP Any monitors the reasoning trace; if the budget is exceeded, the reasoning session is forcefully terminated.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        Agent[Subagent] -->|Request Lease| TTLB[TTL Broker]
        TTLB -->|Validate Budget| MRB[Mission Root Budget]
        TTLB -->|Issue Token| Agent
        Agent -->|Execute Thinking| MCPG[MCP Gateway]
        MCPG -->|Monitor Delta| TTLB
        TTLB -->|Expiry/Violation| MCPG
        MCPG -->|Interdict| Agent
    ```
* **APIs / Interfaces:**
    * `POST /v1/lease/thinking`: Requests a reasoning lease.
    * `X-CBS-TTL-Token`: Standardized header for CBS v1.0 compliance.
* **Data Storage/State:**
    * Lease states are held in high-speed, in-memory buffers within MCP Any.

## 5. Alternatives Considered
* **Global Token Limits:** Rejected as they don't prevent high-frequency "Low-Token" reasoning stalls.
* **Reactive Throttling:** Rejected as it occurs after the resource exhaustion has already begun.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The TTL token is mission-bound and cannot be reused across disparate mission roots.
* **Observability:** Thinking metrics (Efficiency, Stall Rate) are exported to the Swarm Dashboard.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
