# Design Doc: Semantic Subsidiarity Broker (SSB)
**Status:** Draft
**Created:** 2026-07-24

## 1. Context and Scope
As agent swarms scale to 100+ concurrent sub-agents (e.g., Cursor Swarm), the traditional "Parent-Supervised" model becomes a performance bottleneck. **Semantic Subsidiarity** is a new governance pattern where absolute execution authority is delegated to specialist sub-agents, but the "Mission Root" (the user's primary agent) retains a recursive, hardware-locked veto power.

MCP Any needs to solve this by providing the infrastructure to persist and enforce these "Veto Contracts" across heterogeneous framework boundaries, ensuring that a sub-agent's specialized output never commits to the host if it violates the semantic baseline of the mission.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a standardized "Veto Contract" header for UACO/A2A messages.
    * Implement a hardware-attested (TPM) signal for "Veto Execution."
    * Facilitate real-time comparison between sub-agent reasoning outputs and "Mission Root" semantic baselines.
* **Non-Goals:**
    * Automatically resolving reasoning conflicts (handled by the AIR Hub).
    * Providing the LLM logic for the veto decision (handled by the Mission Root agent).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Orchestrator (100+ Agents)
* **Primary Goal:** Delegate a complex codebase migration to a swarm while ensuring no sub-agent deletes core authentication logic.
* **The Happy Path (Tasks):**
    1. The Mission Root defines a "Semantic Baseline" (e.g., "Auth logic must remain OIDC compliant").
    2. Mission Root spawns 10 specialist agents via SSB, attaching a "Veto Contract."
    3. Specialist Agent A proposes a change that simplifies auth but breaks OIDC.
    4. SSB detects the proposed commit and pauses the handoff.
    5. Mission Root performs a "Sovereign Context Comparison" and issues a hardware-locked Veto.
    6. SSB forcefully rolls back Agent A's state and prevents the commit.

## 4. Design & Architecture
* **System Flow:**
    `Mission Root` -> `SSB (Veto Contract)` -> `Sub-Agent` -> `Proposed Commit` -> `SSB (Pre-Commit Buffer)` -> `Mission Root (Veto/Approve)` -> `Host`
* **APIs / Interfaces:**
    * `POST /v1/subsidiarity/contracts`: Create a mission-bound veto contract.
    * `POST /v1/subsidiarity/veto`: Issue a hardware-signed veto signal.
* **Data Storage/State:**
    Veto contracts are stored in the `Blackboard` with `Mission-Root` ownership locks.

## 5. Alternatives Considered
* **Strict Parent Supervision:** Rejected due to 500ms+ latency tax per tool call in large swarms. Subsidiarity allows agents to run at full speed until the *point of commit*.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Veto signals must be hardware-attested to prevent sub-agent "Veto Spoofing" where a specialist tries to veto its own supervisor.
* **Observability:** Logs will include "Subsidiarity Divergence" scores, showing how often sub-agents trigger veto conditions.

## 7. Evolutionary Changelog
* **2026-07-24:** Initial Document Creation.
