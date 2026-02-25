# Design Doc: JIT Permission Broker

**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
As AI agents become more autonomous (e.g., OpenClaw swarms), they frequently encounter "Permission Deadlocks" where a required tool is outside their current capability scope. Currently, this requires a synchronous human intervention which stalls the agent. The JIT Permission Broker aims to provide a path for autonomous or asynchronous permission elevation within a Zero-Trust framework.

## 2. Goals & Non-Goals
* **Goals:**
    * Allow agents to request temporary, scoped permission elevation.
    * Use context-aware risk scoring to decide whether to grant elevation automatically or escalate to a human.
    * Maintain a detailed audit log of all JIT elevations.
    * Integrate with the "Policy Firewall" for enforcement.
* **Non-Goals:**
    * Permanent permission modification (all JIT grants are TTL-bound).
    * General-purpose identity management.

## 3. Critical User Journey (CUJ)
* **User Persona:** Autonomous Agent Swarm (e.g., a "Security Patching" swarm)
* **Primary Goal:** Access a restricted network tool to verify a fix without stalling for 8 hours for a human approval.
* **The Happy Path (Tasks):**
    1. Agent attempts a tool call and receives a `403 Forbidden` with a "JIT-Escalation-Available" hint.
    2. Agent submits a `RequestElevation` call to the JIT Broker, providing the context (current task, reason for need).
    3. JIT Broker evaluates the request against the "Risk Engine" and the "Global Policy."
    4. JIT Broker issues a temporary `JIT-Token` valid for 30 minutes.
    5. Agent retries the tool call with the `JIT-Token` and succeeds.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Agent->>Policy Firewall: Call Tool (Restricted)
        Policy Firewall-->>Agent: 403 + Elevation Hint
        Agent->>JIT Broker: RequestElevation(Scope, Context)
        JIT Broker->>Risk Engine: Evaluate(AgentID, Scope, Context)
        Risk Engine-->>JIT Broker: RiskScore(0.2)
        JIT Broker->>Policy Firewall: Register Temporary Grant
        JIT Broker-->>Agent: JIT-Token (TTL: 30m)
        Agent->>Policy Firewall: Call Tool + JIT-Token
        Policy Firewall->>Tool: Execute
    ```
* **APIs / Interfaces:**
    * `mcp.jit.request_elevation(scope: string, justification: string, ttl: duration)`
    * `mcp.jit.get_status(request_id: string)`
* **Data Storage/State:**
    * Temporary grants are stored in a high-performance in-memory TTL cache (e.g., Redis or internal Go map) and mirrored to the SQLite audit log.

## 5. Alternatives Considered
* **Static Over-Provisioning**: Granting agents broad permissions upfront. Rejected due to security risks (Violates Principle of Least Privilege).
* **Always-Human-In-The-Loop**: Requiring human approval for every 403. Rejected due to latency and "autonomy-killer" effect.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The Risk Engine must detect "Escalation Loops" where an agent requests multiple small permissions to achieve a high-privilege goal (Salami Attack).
* **Observability:** Every JIT request and grant must be visible in the "Supply Chain Attestation Viewer."

## 7. Evolutionary Changelog
* **2026-02-27:** Initial Document Creation.
