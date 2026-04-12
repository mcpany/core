# Design Doc: Hardware-Attested Resource Rebalancer (HARR)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agent swarms evolve into distributed meshes spanning multiple physical devices (OpenClaw SNT), static resource allocation becomes a primary bottleneck. Recursive Mission Exhaustion (RME) occurs when subagents in deep swarms initiate circular tool calls that drain parent token budgets without reaching state convergence.

HARR is designed to act as the authoritative "Budget Arbiter" for MCP Any. It enables the dynamic re-allocation of token and reasoning budgets across parallel missions by monitoring hardware-attested "Urgency Signals" and reasoning entropy. This ensures that high-priority missions have the resources they need to complete while preventing low-priority or stalled missions from exhausting the swarm's global capacity.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a dynamic resource re-allocation engine for distributed swarms.
    * Support hardware-attested "Urgency Signals" for priority-based rebalancing.
    * Neutralize Recursive Mission Exhaustion (RME) via real-time budget throttling.
    * Provide a unified interface for mission-root resource re-alignment.
* **Non-Goals:**
    * HARR will NOT manage the physical provisioning of LLM compute (e.g., GPU scheduling).
    * HARR will NOT perform semantic corrections on the agent's internal reasoning traces (handled by ARI).

## 3. Critical User Journey (CUJ)
* **User Persona:** Distributed Swarm Architect
* **Primary Goal:** Prevent a low-priority sub-mission from stalling a critical root-level task due to token exhaustion.
* **The Happy Path (Tasks):**
    1. The Mission Root defines a global token budget and initial allocation for 3 parallel teammates.
    2. Teammate A enters a recursive loop (RME) while attempting to solve a complex coding task.
    3. The HARR RME Monitor detects the entropy spike and budget drain.
    4. Teammate B sends a hardware-attested "Urgency Signal" for a high-priority user-facing task.
    5. HARR validates the attestation and forcefully reclaims 40% of Teammate A's remaining budget.
    6. HARR shifts the reclaimed budget to Teammate B, allowing the critical task to finish.

## 4. Design & Architecture
* **System Flow:**
```mermaid
graph TD
    A[Mission Root] -->|Initial Budget| B(HARR Arbiter)
    B -->|Throttled Budget| C[Teammate A (RME Detected)]
    B -->|Boosted Budget| D[Teammate B (High Urgency)]
    C -->|Entropy Signals| E(RME Monitor)
    E -->|Rebalance Trigger| B
    D -->|Attested Urgency| B
```
* **APIs / Interfaces:**
    * `POST /v1/resources/rebalance`: Accepts hardware-attested urgency signals and mission IDs.
    * `GET /v1/resources/status`: Provides real-time budget distribution across the mesh.
* **Data Storage/State:**
    * State is managed in the hardware-locked `Mission Budget Registry`, which uses monotonic counters to prevent replay attacks on budget grants.

## 5. Alternatives Considered
* **Static Quotas**: Rejected because they lead to mission stall when specialists encounter unforeseen complexity, even if sibling missions have surplus tokens.
* **User-in-the-Loop Rebalancing**: Rejected for machine-speed swarms; the MTTC (Mean Time to Coordinate) for human approval is too high for sub-second re-allocations.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Urgency signals MUST be hardware-attested (TPM/SEP) to prevent subagents from "Resource Squatting" via spoofed priority claims.
* **Observability:** Rebalancing events are logged to the `Reasoning-Effort Attribution Middleware` for post-mission economic auditing.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
