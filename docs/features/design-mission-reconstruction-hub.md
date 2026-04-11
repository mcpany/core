# Design Doc: Mission Reconstruction Hub (MRH)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agent swarms scale to hundreds of specialist subagents, the overhead of transferring complete mission context across peer-to-peer (P2P) tunnels becomes a performance bottleneck. The "Tunneling Overhead" often leads to "Cognitive Stall," where agents wait for large context objects to synchronize before they can begin reasoning.

The Mission Reconstruction Hub (MRH) implements the "Fragmented Intent Reconstruction" (FIR) protocol. It allows agents to synchronize only minimal, hardware-attested metadata (Intent Shards) and reconstruct the complete, semantically equivalent mission context locally. This drastically reduces network I/O while maintaining the cryptographic continuity of the mission root.

## 2. Goals & Non-Goals
* **Goals:**
    * Reduce inter-node context synchronization latency by 80%+.
    * Provide hardware-attested "Reconstruction Receipts" to ensure reconstructed context matches the mission root.
    * Facilitate lock-free state synchronization in horizontal teammate meshes.
    * Neutralize "Token Storms" caused by redundant JSON state transfers.
* **Non-Goals:**
    * Replacing the Shared KV Store (Blackboard) for persistent state.
    * Handling non-attested or low-trust context fragments.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Perform a complex multi-device research task without incurring the 5s+ latency penalty of full context synchronization.
* **The Happy Path (Tasks):**
    1. Primary agent generates an "Intent Shard" containing mission metadata and a reconstruction key.
    2. The Intent Shard is sent across an authenticated P2P tunnel to a remote specialist subagent.
    3. The remote subagent's Mission Reconstruction Hub receives the shard.
    4. MRH queries the local "Episodic Cache" and uses the reconstruction key to inflate the shard into the full mission context.
    5. MRH verifies the inflated context against the hardware-attested fingerprint provided in the shard.
    6. The specialist subagent begins reasoning immediately using the reconstructed context.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        A[Primary Node] -->|Intent Shard| B[P2P Tunnel]
        B -->|Intent Shard| C[Remote Node MRH]
        C --> D{Episodic Cache}
        D -->|Base Fragments| C
        C -->|FIR Reconstruction| E[Inflated Context]
        E --> F[Reasoning Engine]
        C -->|Verify| G[Hardware Root of Trust]
    ```
* **APIs / Interfaces:**
    * `POST /v1/reconstruct`: Reconstruct context from a fragmented intent shard.
    * `GET /v1/shards/metadata`: Retrieve metadata for localepisodic fragments.
* **Data Storage/State:**
    * Episodic Cache stores frequently used context fragments, indexed by content-addressable hashes.

## 5. Alternatives Considered
* **Binary State Handoff (BSH) with Zlib Compression:** Rejected because it still requires transferring the full (albeit compressed) data, which is insufficient for large (1M+ token) contexts.
* **Centralized State Hub:** Rejected because it introduces a single point of failure and violates the "Sovereign Node" architecture.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All reconstructed contexts must pass hardware-attested fingerprint verification. Unauthorized reconstruction attempts trigger immediate tunnel revocation.
* **Observability:** Metrics on "Reconstruction Hit Rate" and "Sync Latency Reduction" are reported to the performance dashboard.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
