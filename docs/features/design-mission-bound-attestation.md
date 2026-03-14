# Design Doc: Mission-Bound Attestation Token (MBAT)
**Status:** Draft
**Created:** 2026-04-06

## 1. Context and Scope
With the rise of autonomous agent swarms and multi-agent systems (MAS), the primary threat vector has shifted from "External Injection" to "Internal Contagion" (A2A Contagion). When a parent agent delegates a task to a subagent, it currently passes context and authority without cryptographically binding them to a specific mission. If the parent or any intermediate agent is compromised, they can inject malicious intents into the subagent, leading to lateral movement within the infrastructure.

MCP Any needs to solve this by providing a "Mission-Bound Attestation Token" (MBAT) that cryptographically pins every tool call and state mutation to a verified mission statement.

## 2. Goals & Non-Goals
* **Goals:**
    * Cryptographically bind agent sessions to a signed "Mission Intent."
    * Block tool calls or state mutations that diverge from the mission-bound intent at the gateway level.
    * Provide a verifiable chain of custody (Lineage) for mission delegation.
    * Support "State Segmentation" by linking MBATs to specific Context Shards.
* **Non-Goals:**
    * Replacing existing authentication (API keys/mTLS). MBAT is an additional layer of behavioral attestation.
    * Real-time "thought monitoring." MBAT focuses on the *output* (tool calls/state changes) matching the *intent*.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Securely delegate a "Read-Only Code Audit" mission to a subagent without risking the subagent being coerced into "Writing" to the codebase.
* **The Happy Path (Tasks):**
    1. Parent Agent requests a subagent session with a signed "Mission Card" (e.g., "Mission: Audit /src for SQLi").
    2. MCP Any validates the Mission Card and issues an MBAT.
    3. Subagent receives the MBAT and includes it in all tool calls.
    4. MCP Any Policy Engine intercepts a `fs:read` call, verifies the MBAT matches the "Audit" mission, and allows it.
    5. Subagent (compromised) attempts a `fs:write` call.
    6. MCP Any detects the MBAT does not authorize "Write" actions for the "Audit" mission and blocks the call.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        ParentAgent->>Gateway: Request Sub-Mission (Signed Intent)
        Gateway->>Registry: Validate Lineage & Scopes
        Registry-->>Gateway: MBAT issued (bound to Mission Hash)
        Gateway-->>SubAgent: Session + MBAT
        SubAgent->>Gateway: Tool Call (MBAT included)
        Gateway->>PolicyEngine: Validate Call vs Mission Intent
        PolicyEngine-->>Gateway: Authorized / Blocked
    ```
* **APIs / Interfaces:**
    * `POST /v1/mission/attest`: Exchange a signed intent for an MBAT.
    * `Header: X-MCP-MBAT`: Required for all tool calls in a mission-bound session.
* **Data Storage/State:**
    * MBATs are stored in a high-speed, session-bound LRU cache.
    * Mission Lineage is persisted in the Shared Blackboard with "Mission-Bound" isolation.

## 5. Alternatives Considered
* **Capability-Only Tokens:** Rejected because they don't capture "Intent." An agent could have `fs:read` access but use it to exfiltrate data not related to the mission.
* **Continuous LLM Monitoring:** Rejected due to extreme latency and cost. Gateway-level attestation is more performant.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** MBATs are short-lived and cryptographically bound to the hardware/session origin.
* **Observability:** All MBAT violations are logged to the "Security Audit Dashboard" with full intent-divergence traces.

## 7. Evolutionary Changelog
* **2026-04-06:** Initial Document Creation.
