# Design Doc: Governance Checkpoint Middleware
**Status:** Draft
**Created:** 2026-04-05

## 1. Context and Scope
Regulatory frameworks such as the **EU AI Act** and **FINRA 2026** guidelines mandate human oversight for high-risk autonomous actions. Currently, agents often operate in an "all or nothing" permission model. If an agent has access to a tool, it can use it without specific context-aware constraints.

The Governance Checkpoint Middleware provides the "Checkpoint Broker" functionality required to satisfy these mandates by enforcing human-in-the-loop (HITL) approval for specifically flagged high-risk actions.

## 2. Goals & Non-Goals
* **Goals:**
    * Identify high-risk tool calls based on a configurable "Risk Matrix."
    * Enforce mandatory HITL approval or hardware-attested auditor signatures before execution.
    * Provide a standardized "Suspension Protocol" that allows agents to wait safely for approval.
    * Generate hash-chained audit trails for all checkpoint decisions.
* **Non-Goals:**
    * Replacing the standard Policy Firewall (which handles binary allow/deny).
    * Providing the UI for approval (this is handled by the A2UI Gateway).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Admin
* **Primary Goal:** Ensure that no AI agent can modify production firewall rules without explicit approval from a human Network Engineer.
* **The Happy Path (Tasks):**
    1. Admin defines `modify_firewall` as a "P0 High-Risk" tool in the Risk Matrix.
    2. An agent attempts to call `modify_firewall` to block a malicious IP.
    3. The Governance Checkpoint Middleware intercepts the call and suspends the agent's execution state.
    4. A notification is sent to the human approval queue via the A2UI Gateway.
    5. The Network Engineer reviews the reasoning trace and IP logs.
    6. Upon human approval, the Middleware releases the suspension and allows the tool call to proceed.

## 4. Design & Architecture
* **System Flow:**
    * Call Interception -> Risk Assessment -> Execution Suspension -> Approval Request -> Signal Reception -> Execution Resume/Abort.
* **APIs / Interfaces:**
    * `checkpoint.RegisterRisk(toolPattern, riskLevel)`: Defines the threshold for interdiction.
    * `checkpoint.Suspend(callId, context)`: Holds the request in a pending state.
    * `checkpoint.Approve(callId, signature)`: Releases the held request.
* **Data Storage/State:**
    * **Pending Queue:** Persistent (SQLite-backed) queue of actions awaiting approval to survive server reboots.

## 5. Alternatives Considered
* **Static Deny Rules:** Rejected because high-risk actions are often necessary; they just require oversight.
* **Agent-Initiated Approvals:** Rejected because a compromised agent could choose *not* to ask for approval. The infrastructure must force the checkpoint.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All approval signals must be hardware-attested to prevent spoofing by other agents.
* **Observability:** "Mean Time to Approval" and queue depth are tracked as core metrics.

## 7. Evolutionary Changelog
* **2026-04-05:** Initial Document Creation following FINRA/EU AI Act requirements.
