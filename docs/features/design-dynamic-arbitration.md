# Design Doc: Dynamic Policy Arbitration Engine
**Status:** Draft
**Created:** 2026-03-11

## 1. Context and Scope
With the introduction of "Dynamic Capability Negotiation" (DCN) in frameworks like OpenClaw, subagents can now request elevated permissions at runtime. This creates a risk where a compromised parent agent could over-authorize its subagents. The Dynamic Policy Arbitration Engine acts as the "Final Arbiter," validating runtime escalation requests against immutable, human-defined security policies before they are granted.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept and validate all runtime permission escalation (DCN) requests.
    * Use Rego (OPA) or CEL to evaluate requests against "Immutable Safety Bounds."
    * Support "Time-Bound" and "Task-Bound" temporary elevations.
    * Provide a cryptographic proof of authorization for granted escalations.
* **Non-Goals:**
    * Defining the agent's internal logic for *when* to request permissions.
    * Managing static permissions (handled by the core Policy Firewall).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator.
* **Primary Goal:** Ensure a subagent cannot gain host-level `root` access even if requested by a jailbroken parent agent.
* **The Happy Path (Tasks):**
    1. Subagent A requests `fs:write:/etc/hosts` via a DCN call to the Parent Agent.
    2. Parent Agent forwards the request to MCP Any.
    3. Arbitration Engine checks `global_safety.rego`.
    4. Policy defines `/etc/` as a "Protected Zone" for all subagents.
    5. Arbitration Engine rejects the request and logs a "High Severity" security event.

## 4. Design & Architecture
* **System Flow:**
    `Agent (Requester)` -> `Parent Agent (Proxy)` -> `MCP Any (Arbitrator)` -> `Policy Store`
    1. **Evaluation**: The engine receives the `requested_capability`, `requesting_agent_id`, and `context`.
    2. **Bounded Check**: It checks if the request falls within the "Maximum Permissible Envelope" for that specific task.
    3. **Approval Flow**: If within bounds but sensitive, it triggers a `HITL Middleware` prompt.
* **APIs / Interfaces:**
    * `POST /v1/arbitrate/escalate`: Endpoint for DCN validation.
    * `Policy.evaluate(request)` -> `Decision (Allow/Deny/Suspend)`
* **Data Storage/State:**
    * `policies/dynamic/`: Directory of Rego files defining the arbitration logic.

## 5. Alternatives Considered
* **Trust the Parent Agent**: Rejected due to the increasing frequency of agent jailbreaks and prompt injection.
* **Static Permissions Only**: Too restrictive for complex, autonomous agent workflows.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Implements the "Dual Authorization" principle for sensitive actions.
* **Observability**: Real-time "Escalation Dashboard" showing active temporary permissions and rejected requests.

## 7. Evolutionary Changelog
* **2026-03-11:** Initial Document Creation.
