# Design Doc: Cascading Failure Circuit Breaker (CFCB)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
In complex agent swarms and service meshes, a single failing specialist agent or tool can trigger a system-wide collapse. This "Cascading Failure" (OWASP ASI08:2026) occurs when error states, latency spikes, or poisoned outputs propagate through inter-agent dependencies and tool chains. Traditional per-service circuit breakers are insufficient because they lack awareness of the "Agentic Mission" and the semantic relationship between teammates.

The Cascading Failure Circuit Breaker (CFCB) provides an authoritative "Safety Fuse" for the mesh. It monitors the health of inter-agent coordination fragments and tool-call sequences in real-time, automatically isolating failing nodes to preserve the stability of the primary mission root.

## 2. Goals & Non-Goals
* **Goals:**
    * Detect error propagation patterns across multi-agent tool chains.
    * Automatically revoke discovery and communication capabilities for failing specialist agents.
    * Provide "Mission-Bound Isolation" to prevent a failure in one task branch from affecting others.
    * Support hardware-attested failure signals to prevent malicious "Ghost Failures."
* **Non-Goals:**
    * Automatically fixing the underlying bug in the agent (handled by Self-Correction/RAM).
    * Replacing existing transport-level retries.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Maintain mission stability when a specialized "Database Agent" starts returning persistent timeouts or corrupted schemas.
* **The Happy Path (Tasks):**
    1. The mission begins with 4 specialized agents coordinated via the T2T Bridge.
    2. The Database Agent encounters a series of upstream timeouts.
    3. CFCB detects the rising error rate and the propagation of "Wait cycles" to the "UI Agent."
    4. CFCB "trips" the circuit for the Database Agent, revoking its capability to write to the shared mailbox.
    5. CFCB triggers an "Urgent Interrupt" to the Supervisor, suggesting a failover or task re-assignment.
    6. The rest of the swarm (Linter, Coder) continues operating within restricted scopes, preventing system-wide stall.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Inter-Agent Coordination] --> B[CFCB Monitor]
        B --> C{Failure Threshold?}
        C -->|Yes| D[Isolate Agent]
        C -->|No| E[Continue Monitoring]
        D --> F[Revoke Mailbox Access]
        D --> G[Notify Mission Root]
        H[Upstream Tool] --> B
    ```
* **APIs / Interfaces:**
    * `GET /v1/mesh/health`: Returns real-time health scores for all active agent tool-chains.
    * `POST /v1/mesh/circuit/reset`: Manually resets a tripped circuit after user attestation.
* **Data Storage/State:**
    * Failure metrics are stored in ephemeral, high-speed memory buffers. Trip states are persisted in the Shared KV Store (Blackboard) under mission-root isolation.

## 5. Alternatives Considered
* **Global Rate Limiting**: Rejected because it penalizes healthy agents and does not stop error propagation.
* **Manual Supervisor Oversight**: Rejected due to high latency in machine-speed swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** CFCB uses hardware-attested identity fragments (SMI) to ensure isolation signals cannot be spoofed by rogue subagents.
* **Observability:** Integrated with the "Service Mesh Topology Monitor" to visualize "tripped" paths.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
