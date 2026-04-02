# Design Doc: Wait-Graph Lease Arbiter (WLA)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agent swarms transition from linear sessions to high-density horizontal meshes (e.g., Claude Code Agent Teams), the reliance on Conflict-Free Replicated Data Types (CRDTs) has proven insufficient for behavioral coordination. While CRDTs ensure state convergence, they do not manage the temporal logic of task acquisition. Recent benchmarks reveal that agents frequently enter recursive dependency loops—where Agent A waits for a resource held by Agent B, which is waiting for a task from Agent A—leading to "Cognitive Stalls" exceeding 5 seconds.

MCP Any needs to solve this by moving from passive mailbox sharding to active, kernel-level coordination. The Wait-Graph Lease Arbiter (WLA) will provide the infrastructure to detect these circular dependencies in real-time and forcefully re-allocate mission-root leases to maintain swarm momentum.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement real-time directed acyclic graph (DAG) monitoring for all active teammate task claims.
    * Provide automated "Lease Revocation" triggers for detected circular dependencies.
    * Integrate priority-weighted re-allocation based on the Mission-Root manifest.
    * Maintain sub-millisecond overhead for the wait-graph analysis.
* **Non-Goals:**
    * Replacing CRDTs for state synchronization (WLA manages the *lock*, not the *data*).
    * Resolving application-level logic errors within the agent's reasoning.

## 3. Critical User Journey (CUJ)
* **User Persona:** Distributed Swarm Architect
* **Primary Goal:** Resolve a 3-agent deadlock in a local-to-cloud hybrid mesh without manual supervisor intervention.
* **The Happy Path (Tasks):**
    1. Agent A (Local) claims `fs:write` but requires `db:query`.
    2. Agent B (Cloud) claims `db:query` but requires `fs:read` (held by Agent C).
    3. Agent C claims `fs:read` but requires `fs:write` (held by Agent A).
    4. WLA detects the circular dependency on the mission blackboard.
    5. WLA identifies Agent A as the "Mission Primary" based on hardware-attested lineage.
    6. WLA forcefully revokes the `db:query` lease from Agent B and assigns it to Agent A.
    7. The swarm resumes execution in <50ms.

## 4. Design & Architecture
* **System Flow:**
    * **Dependency Collector**: Ingests `Claim` and `Wait` signals from the T2T Encryption Bridge.
    * **Graph Engine**: Maintains a global Wait-Graph for the mission scope.
    * **Arbiter**: Executes cycle-detection algorithms (e.g., Tarjan's) on every graph mutation.
    * **Enforcer**: Communicates with the EPM (Ephemeral Privilege Manager) to rotate leases.
* **APIs / Interfaces:**
    * `POST /v1/coordination/claim`: Request a task/resource lease.
    * `POST /v1/coordination/wait`: Signal a dependency on an external lease.
    * `STREAM /v1/coordination/alerts`: Real-time notification of lease revocation.
* **Data Storage/State:**
    * In-memory Graph structure, mirrored to the versioned Blackboard (Shared KV Store) for recovery.

## 5. Alternatives Considered
* **Timeouts (Static)**: Rejected due to "Thundering Herd" problems and the difficulty of setting optimal thresholds for heterogeneous LLMs.
* **Centralized Locking**: Rejected due to high latency and single-point-of-failure risks in distributed tunnels.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** WLA decisions must be cryptographically signed by the Mission-Root key. Revocations must be hardware-attested to prevent "Lease Hijacking" by compromised specialists.
* **Observability:** Integrated into the **Wait-Graph Visualizer** in the UI, showing real-time edge weights and cycle-break events.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
