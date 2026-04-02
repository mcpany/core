# Design Doc: Recursive Resource Reclamation (RRR) v2 Manager
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agent swarms scale horizontally, the "Resource Squatting" problem has become a primary bottleneck for enterprise stability. Specialist agents often maintain hold over significant token and reasoning budgets long after their specific sub-task is completed, preventing the mission root from re-allocating those resources to more urgent branches.

The **Recursive Resource Reclamation (RRR) v2 Manager** is an authoritative economic security service for MCP Any. It provides high-frequency, autonomous reclamation of unused budgets from dormant or non-convergent sub-missions, ensuring optimal resource utilization across the entire mission-root lineage.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement sub-millisecond, hardware-attested budget reclamation.
    * Automate the termination of dormant sub-agent sessions based on inactivity heartbeats.
    * Provide a centralized "Economic Ledger" for mission-root resource consumption.
    * Integrate with the Reasoning-Budget Firewall (RBF) to enforce hierarchical quotas.
* **Non-Goals:**
    * Directly managing LLM token pricing (handled by providers).
    * Restricting mission-root budgets (governed by user policy).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Swarm Operator
* **Primary Goal:** Reclaim 50% of an overallocated token budget from a "blocked" researcher sub-agent and re-allocate it to an active "debugger" sub-agent.
* **The Happy Path (Tasks):**
    1. Researcher agent spawns with a 1M token budget.
    2. Researcher agent hits a "Cognitive Stall" (inactivity for 60s).
    3. RRR v2 Manager detects the stall via the Subagent Heartbeat Provider.
    4. RRR Manager issues a "Lease Revocation" signal to the Researcher agent.
    5. Unused tokens are returned to the mission-root pool.
    6. Debugger agent requests an expansion, and the RRR Manager authorizes it using the reclaimed budget.
    7. Mission continues without hitting the global budget cap.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Subagent Session] -->|Heartbeat| B[RRR v2 Manager]
        B -->|Status Check| C[Activity Monitor]
        C -->|Dormant Detected| D[Budget Arbiter]
        D -->|Revoke Lease| E[Reasoning-Budget Firewall]
        E -->|Update Quota| F[Mission-Root Pool]
        F -->|Re-allocate| G[Active Specialist]
    ```
* **APIs / Interfaces:**
    * `rrr.ReclaimBudget(sessionID, missionToken) -> ReclaimedAmount`: Forcefully reclaims unused resources.
    * `rrr.GetMissionLedger(missionRootID) -> LedgerData`: Returns detailed resource attribution.
* **Data Storage/State:**
    * **Resource Registry:** A high-speed, in-memory KV store (backed by SQLite) tracking active leases and parentage.

## 5. Alternatives Considered
* **Manual Revocation**: Rejected due to the speed requirements of autonomous swarms.
* **Static Timeouts**: Rejected as they don't account for the variable reasoning intensity of different tasks.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All reclamation signals must be signed by the hardware-attested RRR authority.
* **Observability:** Integrated with the "Mission Budget Dashboard" for real-time visualization of reclamation events.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation. Evolving from RRR Manager (v1) to support aggressive, autonomous reclamation.
