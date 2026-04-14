# Design Doc: Reflection Storm Control (Consensus Deadlock Resolver)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the introduction of Teammate Reflection Quorums (TRQ) in horizontal swarms (e.g., Claude Code Agent Teams), agents are now required to reach a consensus before committing changes to the shared task list or executing high-stakes tools. However, in high-density teams with specialists holding conflicting mission-root constraints, these reflection cycles often enter "Consensus Deadlocks" or "Reflection Storms"—infinite loops of refinement that exhaust token budgets without reaching the required quorum.

MCP Any needs to solve this by acting as the authoritative "Consensus Arbiter." By monitoring the progress of teammate reflection quorums and applying mission-aligned tie-breaking policies, MCP Any ensures swarm stability and prevents cognitive stalls.

## 2. Goals & Non-Goals
* **Goals:**
    * Detect circular dependencies and non-convergent reasoning loops in reflection quorums.
    * Provide automated "Tie-Breaking" using priority-weighted mission-root rules.
    * Facilitate atomic state rollbacks for the Blackboard when consensus fails to reach a threshold within a timeout.
    * Enforce mission-root budget limits on reflection cycles.
* **Non-Goals:**
    * Automatically resolving low-risk semantic disagreements that don't impact mission-root stability.
    * Replacing the internal model's reasoning logic; only governing the *flow* and *finality* of the consensus.

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Swarm Orchestrator
* **Primary Goal:** Prevent a 10-agent security audit swarm from stalling when 2 agents disagree on the risk level of a specific dependency.
* **The Happy Path (Tasks):**
    1. The Lead Agent initiates a Reflection Quorum for a high-risk PR commit.
    2. Teammates submit reflection monologues; MCP Any monitors the "Consensus Strength."
    3. The quorum reaches a 60/40 split and enters a refinement loop.
    4. The CDR (Consensus Deadlock Resolver) detects that reasoning entropy is increasing while consensus is flat for 3 turns.
    5. MCP Any triggers the Tie-Breaker, injecting the "Mission-Root Priority" signal (e.g., "Safety Over Speed").
    6. Specialist agents adjust their stance based on the hardware-attested priority signal.
    7. Consensus is reached at 80%, and the task is committed.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Teammates] -->|Reflection Monologues| B(Quorum Monitor)
        B -->|Entropy/Consensus Scores| C{Deadlock Detector}
        C -->|Deadlock Detected| D[CDR Service]
        D -->|Inject Priority Signal| A
        C -->|Consensus Reached| E[Blackboard Commit]
        D -->|Timeout Exhausted| F[Atomic Rollback]
    ```
* **APIs / Interfaces:**
    * `POST /v1/quorum/initiate`: Start a reflection session with a specific threshold.
    * `POST /v1/quorum/reflect`: Submit a fragment-level reflection for scoring.
    * `GET /v1/quorum/status`: Retrieve real-time consensus metrics and deadlock risk.
* **Data Storage/State:**
    * Quorum states are held in an ephemeral, memory-mapped buffer linked to the mission-root session.
    * CDR policies (Tie-breaking rules) are stored in the hardware-locked mission manifest.

## 5. Alternatives Considered
* **Static Timeouts**: Rejected because simple timeouts lead to mission failure without attempting reconciliation, causing significant re-planning overhead.
* **Human-in-the-Loop (HITL) for every deadlock**: Rejected for high-density swarms as it causes "Approval Fatigue" and negates the speed benefits of autonomous Agent Teams.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** CDR priority signals must be hardware-attested to ensure a compromised subagent cannot "Tie-Break" in its own favor.
* **Observability:** Quorum transitions and CDR interdictions are logged to the **Swarm Execution Monitor** with reasoning-trace IDs.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
