# Design Doc: MAESTRO Orchestration Policy Engine

**Status:** Draft
**Created:** 2026-03-05

## 1. Context and Scope
With the rise of multi-agent systems like OpenClaw and the MAESTRO framework, agents are no longer just calling tools; they are spawning other agents (sub-agents) and communicating across agent boundaries. Without a centralized policy engine, these swarms can become unpredictable, expensive (token waste), and insecure (unauthorized sub-agent escalation). MCP Any must implement a Layer 7 Orchestration Policy Engine that governs agent-to-agent (A2A) interactions.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement an "Explicit Allowlist" for sub-agent spawning.
    *   Monitor and log all inter-agent (A2A) communication patterns.
    *   Enforce "Recursive Permission Scoping" where a sub-agent's capabilities are a strict subset of its parent's.
    *   Integrate with the "Policy Firewall" for unified tool and agent governance.
*   **Non-Goals:**
    *   Replacing the agent's internal reasoning or "Brain."
    *   Implementing a proprietary A2A messaging protocol (use industry standards like A2A).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Security Architect.
*   **Primary Goal:** Prevent a Research Agent from spawning an unauthorized "Execution Agent" that has access to production servers.
*   **The Happy Path (Tasks):**
    1.  Architect defines a MAESTRO policy: `allow: sub-agent spawn where parent.role == 'researcher' and child.role == 'summarizer'`.
    2.  The Research Agent attempts to spawn a 'summarizer' agent.
    3.  MCP Any validates the request against the policy engine and allows it.
    4.  The Research Agent attempts to spawn an 'executor' agent.
    5.  MCP Any blocks the request and triggers a security alert in the Dashboard.

## 4. Design & Architecture
*   **System Flow:**
    - **Interception**: All A2A messages (especially spawn requests) pass through the `MAESTROEngineMiddleware`.
    - **Policy Evaluation**: The request is evaluated using Rego/CEL (sharing the engine with the Policy Firewall).
    - **Lineage Tracking**: Every agent is assigned a "Lineage ID" to track parent-child relationships and enforce recursive scoping.
*   **APIs / Interfaces:**
    - `POST /v1/orchestration/policy`: API to update orchestration rules.
    - `POST /v1/agent/spawn`: A2A-compatible endpoint for agent lifecycle management.
*   **Data Storage/State:** Agent lineage and active session state are stored in the `Shared KV Store` (Blackboard).

## 5. Alternatives Considered
*   **Hardcoded Rules in Agents**: *Rejected* due to lack of centralized control and auditability.
*   **Standalone Orchestrator**: *Rejected* to avoid architectural complexity; MCP Any already sits in the data path.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Essential for preventing "Agent Escalation" attacks.
*   **Observability:** Visualized in the "Agent Chain Tracer" (A2A) in the UI.

## 7. Evolutionary Changelog
*   **2026-03-05:** Initial Document Creation.
