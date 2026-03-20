# Design Doc: Asynchronous Intent Consistency (AIC) Broker
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
In high-density horizontal meshes, synchronous state locking (Mailbox Lock) has become the primary bottleneck for performance. Agents spend up to 40% of their reasoning cycle waiting for state synchronization. To achieve machine-speed coordination, MCP Any needs to move from a "Lock-First" to a "Consistency-First" model. The AIC Broker utilizes CRDTs to allow teammates to work on local intent fragments that merge asynchronously into the mission-root.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide non-blocking intent synchronization for sharded meshes.
    * Utilize CRDTs to ensure deterministic convergence of intent fragments.
    * Maintain hardware-attested lineage for every merged fragment.
    * Support sub-10ms local intent updates.
* **Non-Goals:**
    * Resolving semantic conflicts (handled by the AIA Broker).
    * Managing non-intent state (e.g., large files).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator (10+ parallel teammates)
* **Primary Goal:** Allow 10 agents to simultaneously update their "Shared Task Progress" without causing coordination deadlocks.
* **The Happy Path (Tasks):**
    1. The mission-root initializes an AIC shard for "Task Progress."
    2. Parallel teammates mount the shard as a local CRDT.
    3. Teammates perform local updates to their task status (e.g., "Step 1 Complete").
    4. The AIC Broker asynchronously synchronizes fragments over the isolated named-pipe transport.
    5. The mission-root observes the deterministically merged progress state without ever issuing a lock.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        T1[Teammate 1] -->|Local Update| L1[Local AIC Fragment]
        T2[Teammate 2] -->|Local Update| L2[Local AIC Fragment]
        L1 -->|Async Sync| AIC[AIC Broker]
        L2 -->|Async Sync| AIC
        AIC -->|CRDT Merge| MR[Mission Root State]
        MR -->|Converged State| T1
        MR -->|Converged State| T2
    ```
* **APIs / Interfaces:**
    * `POST /v1/aic/shard/init`: Initialize a CRDT-backed intent shard.
    * `POST /v1/aic/shard/update`: Submit a local intent fragment for merging.
    * `GET /v1/aic/shard/view`: Retrieve the current converged state.
* **Data Storage/State:**
    * CRDT fragments are stored in a distributed memory-mapped buffer.

## 5. Alternatives Considered
* **Optimistic Locking:** Rejected as it causes excessive rollbacks in high-contention meshes.
* **Centralized Redis/SQL:** Rejected due to transport latency and lack of hardware-attested fragment lineage.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Every AIC fragment must carry a RLV-compliant lineage signature. The AIC Broker rejects fragments with invalid mission-root ancestry.
* **Observability:** Visualization of "Convergence Latency" and "Fragment Churn" in the LFMA Mesh State Debugger.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
