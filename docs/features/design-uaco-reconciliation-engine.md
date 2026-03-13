# Design Doc: UACO v2.1 Reconciliation Engine
**Status:** Draft
**Created:** 2026-03-30

## 1. Context and Scope
As agent swarms grow in complexity and depth, "State Drift"—where an agent's internal monologue or sub-task execution diverges from the global swarm state (Blackboard)—becomes inevitable. Current systems are reactive, often failing only after a catastrophic divergence. The UACO v2.1 Reconciliation Engine introduces an "Immuno-Governance" layer that proactively identifies and heals these drifts using self-correction primitives.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement native support for UACO v2.1 "Reconciliation Bids."
    * Provide a mechanism for agents to "Self-Report" state divergence to the coordination hub.
    * Orchestrate "Healing Loops" that can roll back or re-align specific branches of an intent tree without freezing the entire swarm.
    * Maintain an immutable audit trail of reconciliation events for forensic analysis.
* **Non-Goals:**
    * Automatically resolving logical reasoning errors (this requires LLM intervention).
    * Replacing the base UACO delegation logic.

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Security Architect
* **Primary Goal:** Automatically heal a swarm where a subagent has been "tinted" by a Shadow Intent attack.
* **The Happy Path (Tasks):**
    1. A subagent detects that its current task output conflicts with the parent's signed intent.
    2. The subagent issues a "Reconciliation Bid" to the UACO v2.1 Engine.
    3. The Engine pauses the affected sub-branch and snapshots the current Blackboard state.
    4. The Engine triggers a "Validation Monitor" agent to review the divergence.
    5. Upon validation, the Engine executes an "Atomic State Rollback" to the last known good checkpoint for that specific intent branch.
    6. The subagent is re-initialized with the corrected state and resumes the task.

## 4. Design & Architecture
* **System Flow:**
    `Divergence Detected` -> `Reconciliation Bid` -> `Branch Isolation` -> `Healing Loop (Rollback/Re-align)` -> `Mission Resumption`
* **APIs / Interfaces:**
    * `ReconciliationEngine` Interface: `ProposeReconciliation(bid *UACOReconBid) (*ReconResult, error)`
    * `HealBranch(intent_id string, checkpoint_id string)`: Targetted rollback of a specific intent tree branch.
* **Data Storage/State:**
    * Leverages the "Atomic State Rollback Middleware" and "Blackboard Integrity Validator."

## 5. Alternatives Considered
* **Global Swarm Rollback**: Too disruptive and expensive in terms of token consumption and latency.
* **Passive Logging**: Provides no protection against active "Shadow Intent" or "Ghost Fragment" attacks.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Reconciliation bids must be multi-signed to prevent "Denial of Progress" attacks by compromised agents.
* **Observability:** A "State Alignment Monitor" UI will provide real-time visibility into healing events.

## 7. Evolutionary Changelog
* **2026-03-30:** Initial Document Creation.
