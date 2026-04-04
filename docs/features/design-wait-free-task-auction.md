# Design Doc: Wait-Free Task Auction (WFTA) Broker
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agent swarms move from hierarchical sessions to horizontal "teammate meshes" (e.g., Claude Code Agent Teams), the primary coordination bottleneck has shifted from transport latency to **State Lock Contention**. Synchronous locks on the shared task list create "Cognitive Stall" where teammates wait seconds for access, while "Task Card Shadowing" vulnerabilities allow compromised agents to suppress legitimate bids.

MCP Any needs a **Wait-Free Task Auction (WFTA) Broker** to enable optimistic task-claiming and hardware-attested bidding. This ensures that coordination is non-blocking while maintaining the absolute integrity of the decentralized task allocation process.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement optimistic, lock-free task-claiming for parallel teammates.
    * Provide hardware-attested (TPM/SEP) verification for all task bids and claims.
    * Neutralize "Bid-Shadowing" and "Auction Hijacking" via monotonic auction counters.
    * Reduce inter-teammate coordination latency by 80% in high-density swarms.
* **Non-Goals:**
    * Replacing the underlying Shared KV Store (Blackboard).
    * Managing the internal reasoning logic of specialist agents.
    * Implementing a global cross-cloud auction (restricted to local mesh scope).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Securely allocate 50+ concurrent tasks to a mesh of 20 specialists without blocking reasoning loops.
* **The Happy Path (Tasks):**
    1. The Mission-Root posts a batch of task cards to the WFTA mailbox.
    2. Specialist agents perform optimistic local "claims" on available tasks based on their capability cards.
    3. Specialists broadcast their hardware-attested bids to the WFTA Broker.
    4. The WFTA Broker asynchronously resolves conflicts (e.g., multiple agents claiming the same task) using priority-weighted mission-root rules.
    5. The "Winner" is issued a TPM-signed mission-bound lease; the "Losers" automatically revert their local state and re-bid.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Mission Root] -- POST Task Cards --> B[WFTA Broker]
        B -- Broadcast --> C[Specialist A]
        B -- Broadcast --> D[Specialist B]
        C -- Optimistic Claim --> E[Local State A]
        D -- Optimistic Claim --> F[Local State B]
        C -- Signed Bid --> B
        D -- Signed Bid --> B
        B -- Conflict Resolution --> G{Winner?}
        G -- Yes --> H[Issue TPM Lease]
        G -- No --> I[Trigger State Revert]
    ```
* **APIs / Interfaces:**
    * `POST /v1/auction/tasks`: Submit a batch of tasks.
    * `POST /v1/auction/bid`: Submit a hardware-attested task claim.
    * `GET /v1/auction/status`: Monitor auction resolution and lease issuance.
* **Data Storage/State:**
    * Uses **Conflict-Free Replicated Data Types (CRDTs)** for the shared task list to allow concurrent mutations.
    * Auction logs are stored in a hardware-locked monotonic append-only buffer.

## 5. Alternatives Considered
* **Global Mutex Locks**: Rejected due to prohibitive latency and "Cognitive Stall" at scale (20+ agents).
* **Centralized Scheduler**: Rejected to maintain framework-neutrality and support decentralized "Agent Teams" coordination.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):**
    * Every bid must be accompanied by a hardware-attested **Mission-Root Token**.
    * **HAAI (Hardware-Attested Auction Integrity)** prevents bid-shadowing by using monotonic auction counters that cannot be re-played.
* **Observability:**
    * Real-time visualization via the **Lock-Free Coordination Monitor**.
    * Detailed logging of "Bid Collisions" and resolution win/loss rates.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
