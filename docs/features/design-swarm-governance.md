# Design Doc: Swarm Governance Middleware (Critic Agent)
**Status:** Draft
**Created:** 2026-03-08

## 1. Context and Scope
The rise of "Agentic Swarms" (e.g., Gemini's Generalist Agent and Claude's Agent Teams) has introduced a new class of risk: delegated decision-making without oversight. When a parent agent delegates a task to 100+ subagents, it becomes impossible for a human—or even the parent agent—to verify that every sub-task aligns with the original intent. MCP Any needs a standardized "Governance Layer" to intercept, audit, and validate inter-agent communication.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept all inter-agent (A2A) messages within the MCP Any mesh.
    * Provide an "Auditor" (Critic Agent) that compares subagent actions against a signed "Intent Manifest" from the parent.
    * Support "Hardened Rejection" of messages that violate project-level policies.
    * Implement "Steering Hint" propagation to ensure specialized agents remain aligned with high-level goals.
* **Non-Goals:**
    * Replacing the primary reasoning of the agents.
    * Adding excessive latency to simple tool calls (auditing should be asynchronous where possible, or extremely lightweight).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise AI Safety Officer
* **Primary Goal:** Ensure that a "Marketing Swarm" does not accidentally bypass budget constraints or brand guidelines when delegating tasks to specialized "Copywriter" and "Designer" subagents.
* **The Happy Path (Tasks):**
    1. Parent agent (The Architect) generates an "Intent Manifest" signed with its cryptographic identity.
    2. The Architect delegates a task to a "Copywriter" subagent via the A2A Bridge.
    3. The Swarm Governance Middleware intercepts the message.
    4. The Critic Agent verifies the message against the Intent Manifest and the global "Brand Policy."
    5. If valid, the message is delivered. If invalid (e.g., requesting a tool outside of its scope), the message is blocked and an "Audit Failure" is logged.

## 4. Design & Architecture
* **System Flow:**
    - **Intent Manifest**: A structured object containing the parent agent's signed goal and constraints.
    - **Critic Engine**: A middleware layer that uses a specialized, high-speed LLM or Rego/CEL policies to evaluate messages.
    - **Lineage Tracking**: Every message is tagged with a `_swarm_lineage` token that links back to the original Architect's Intent.
* **APIs / Interfaces:**
    - `POST /v1/swarm/audit`: Endpoint for manual or automated auditing of a message chain.
    - Metadata extension: `_mcp_intent_signature`: Cryptographic signature of the parent's goal.
* **Data Storage/State:** Intent manifests are stored in the `Shared KV Store` (Blackboard) for the duration of the swarm session.

## 5. Alternatives Considered
* **Centralized Orchestrator**: Forcing all communication through a single "Master" agent. *Rejected* because it creates a single point of failure and bottleneck for large swarms.
* **Manual HITL for every sub-task**: Requiring human approval for every delegation. *Rejected* as it defeats the purpose of autonomous swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The Critic Agent uses Zero-Trust principles to verify that every subagent is explicitly authorized to perform its requested action.
* **Observability:** Audit logs are fed directly into the "A2A Interaction Observability Dashboard."

## 7. Evolutionary Changelog
* **2026-03-08:** Initial Document Creation.
