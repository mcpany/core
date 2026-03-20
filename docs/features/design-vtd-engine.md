# Design Doc: VTD Autonomous Delegation Engine
**Status:** Draft
**Created:** 2026-04-16

## 1. Context and Scope
The "Approval Fatigue" bottleneck remains the primary inhibitor to scaling autonomous agent swarms. While the Delegation Attestation Layer (DAL) provides the "Safety Proofs," the VTD Autonomous Delegation Engine (VADE) is the execution layer that actually makes the decision to bypass manual approval for low-risk A2A (Agent-to-Agent) handoffs. It acts as the "Autonomous Controller" for the A2A Messaging Hub.

## 2. Goals & Non-Goals
* **Goals:**
    * Automatically execute A2A task delegations when the DAL provides a high-confidence "Safety Proof."
    * Manage "Delegation Budgets" to prevent autonomous swarms from over-consuming resources.
    * Provide an "Emergency Halt" mechanism that reverts all autonomous delegations if a mission-level anomaly is detected.
    * Integrate with the "Resident Integrity Monitor" to ensure the subagent's environment is still attested before delegation.
* **Non-Goals:**
    * Performing the safety evaluation itself (this is delegated to the DAL).
    * Bypassing security policies for high-risk operations.

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Swarm Administrator
* **Primary Goal:** Scale a code-review swarm to handle 100+ PRs without requiring a human to approve every sub-task delegation between the "Linter Agent" and the "Security Scanner Agent."
* **The Happy Path (Tasks):**
    1. Linter Agent proposes a "Scan for Secrets" task to the Security Scanner Agent.
    2. A2A Messaging Hub routes the proposal to the DAL.
    3. DAL returns a "Safety Proof" with a 0.99 confidence score and "Low Risk" classification.
    4. VADE checks the "Delegation Budget" for the current mission.
    5. VADE confirms the Security Scanner's sandbox is still hardware-attested via the Resident Integrity Monitor (RIM).
    6. VADE automatically signs the delegation and triggers the handoff.
    7. The task is executed autonomously.

## 4. Design & Architecture
* **System Flow:**
    `[DAL Safety Proof] -> [VADE] -> (Budget + RIM Check) -> [A2A Messaging Hub] -> (Auto-Approval) -> [Subagent]`
* **APIs / Interfaces:**
    * `DelegationController`: `AuthorizeAutonomousHandoff(proof SafetyProof) (ApprovalToken, error)`
    * `BudgetService`: `CheckAndConsume(missionID string, resourceCost float64) (bool)`
* **Data Storage/State:**
    * Persistent storage of active delegation tokens and their associated safety proofs.
    * Real-time tracking of mission budgets in the Blackboard (Shared KV Store).

## 5. Alternatives Considered
* **Time-Based Approvals:** Rejected because "Implicit Approval" after X minutes is insecure.
* **Rule-Based Routing:** Rejected as too brittle for the non-deterministic nature of LLM agent task bidding.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** VADE must never authorize a delegation if the "Root Mission Intent" has been modified or if the hardware attestation for the subagent's sandbox is stale.
* **Observability:** All autonomous decisions are logged with "Auto-Approval" tags in the A2A Messaging Hub Dashboard.

## 7. Evolutionary Changelog
* **2026-04-16:** Initial Document Creation.
