# Design Doc: Reflective Quorum (RQ) Hub
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
In parallel Agent Teams, teammates often act on local assumptions that conflict with the global mission root, leading to "State Splicing" and reasoning drift. Existing quorums validate tool outputs but not the underlying assumptions.

The RQ Hub orchestrates a "Synchronous Reflection" phase. Before state is committed, teammates must peer-review each other's reasoning assumptions against the hardware-attested mission manifest.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement synchronous reflection checkpoints for horizontal Agent Teams.
    * Mandate multi-agent consensus on "Mission Assumptions" before Blackboard writes.
    * Provide a hardware-attested audit trail of the reflection phase.
* **Non-Goals:**
    * Governing individual model reasoning (handled by SRM).
    * Resolving tool-level conflicts (handled by MRCR).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Prevent 3 parallel agents from overwriting the same mission parameter with conflicting assumptions.
* **The Happy Path (Tasks):**
    1. Agent A proposes a state change to the Blackboard.
    2. RQ Hub triggers a "Reflection Challenge" to Agent B and Agent C.
    3. Agents B and C review Agent A's reasoning traces for consistency with the Mission Root.
    4. Upon consensus, RQ Hub releases the commit lock and signs the transaction.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        AgentA->>RQHub: Propose State Change
        RQHub->>AgentB: Request Reflection Review
        RQHub->>AgentC: Request Reflection Review
        AgentB-->>RQHub: Assumption Verified
        AgentC-->>RQHub: Assumption Verified
        RQHub->>Blackboard: Commit Change
    ```
* **APIs / Interfaces:**
    * `POST /quorum/reflect`: Initiates a reflection cycle.
    * `GET /quorum/status/{cycle_id}`: Polls consensus strength.
* **Data Storage/State:** Reflection logs stored as "Ghost Shards" in the ZCMB until consensus is reached.

## 5. Alternatives Considered
* **Asynchronous Pruning**: Allowing conflicts and pruning them later. Rejected due to the risk of "Instruction Pollution" in long-running loops.
* **Central Supervisor Review**: Routing all assumptions through a parent agent. Rejected as a performance bottleneck in 10+ agent meshes.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Teammates must provide hardware-attested "Reflection Tokens" to prevent spoofing of the review process.
* **Observability:** Monitored via the "Reflective Drift Meter" in the UI.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
