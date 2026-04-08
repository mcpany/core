# Design Doc: Trust Inheritance Middleware
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
In large, elastic swarms, agents frequently join and leave the mesh. Requiring a full hardware-attested handshake for every node-to-node discovery event leads to "Handshake Storms," causing significant coordination latency.

Trust Inheritance allows a new agent to join an existing mission by inheriting the trust posture of a verified "Mission Anchor" (usually the parent agent). This maintains security while drastically reducing discovery-phase overhead.

## 2. Goals & Non-Goals
* **Goals:**
    * Facilitate sub-millisecond trust propagation for new teammates.
    * Maintain a verifiable lineage back to the hardware-attested mission root.
    * Automatically revoke inherited trust if the mission anchor is compromised.
* **Non-Goals:**
    * Bypassing initial hardware attestation for the mission anchor.
    * Providing trust between disparate missions (trust is mission-bound).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Node Agent Mesh Orchestrator
* **Primary Goal:** Scale a mission from 5 to 50 agents in under 1 second.
* **The Happy Path (Tasks):**
    1. Mission Anchor (Parent) completes a full hardware handshake with MCP Any.
    2. Parent spawns 45 specialist subagents.
    3. Instead of 45 handshakes, each subagent presents a "Trust Inheritance Ticket" signed by the Parent.
    4. MCP Any verifies the ticket against the active Mission Anchor state.
    5. Subagents are granted discovery capabilities immediately.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Mission Root] -->|Full Handshake| B(MCP Any)
        A -->|Sign Ticket| C[New Teammate]
        C -->|Present Ticket| B
        B -->|Verify Lineage| D{Valid?}
        D -->|Yes| E[Grant Access]
        D -->|No| F[Demand Full Handshake]
    ```
* **APIs / Interfaces:**
    * `IssueInheritanceTicket(targetNodeID string) -> ticket`
    * `RedeemInheritanceTicket(ticket string) -> missionToken`
* **Data Storage/State:**
    * Inheritance lineages are tracked in the ephemeral mesh state.

## 5. Alternatives Considered
* **Persistent Session Keys**: Rejected because they don't handle the dynamic entry/exit of new nodes as cleanly as inheritance tickets.
* **Reduced Strength Handshakes**: Rejected because they degrade the overall security posture of the mesh.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Inheritance is limited to a single depth level to prevent "Trust Dilution."
* **Observability:** Inheritance events are visualized in the Trust Inheritance Map.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
