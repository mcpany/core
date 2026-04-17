# Design Doc: Federated Scratchpad Arbiter (FSA)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the rise of "Agent Teams" operating across distributed physical nodes, the need for a shared, synchronized workspace has become critical. Claude Code's "Federated Scratchpad Synchronization" (FSS) addresses this, but introduces a major security frontier: **Scratchpad Poisoning**.

The Federated Scratchpad Arbiter (FSA) acts as the secure coordination layer for these shared workspaces. It ensures that state mutations across Sovereign Node Tunnels (SNT) are atomic, hardware-attested, and semantically verified to prevent a compromised teammate from injecting malicious instructions into a sibling's working memory.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a distributed, hardware-attested lock manager for shared teammate scratchpads.
    * Perform real-time semantic analysis of all scratchpad writes to detect poisoning.
    * Facilitate sub-millisecond synchronization of scratchpad state across SNT-compliant tunnels.
    * Ensure that "Reasoning-Aware Redaction" (RAR) is applied to all cross-node state transfers.
* **Non-Goals:**
    * The FSA will NOT act as a general-purpose database; it is specifically for high-frequency teammate "scratch" state.
    * It will NOT manage persistent filesystem changes (handled by the Shadow-FS).

## 3. Critical User Journey (CUJ)
* **User Persona:** Distributed Agent Team Lead
* **Primary Goal:** Synchronize code-review findings between 3 agents running on different servers without risking cross-node command injection.
* **The Happy Path (Tasks):**
    1. Agent A (Specialist) requests an atomic write-lock on the shared `.scratchpad` file.
    2. The FSA verifies Agent A's hardware-attested mesh identity.
    3. Agent A writes a reasoning fragment to the scratchpad.
    4. The FSA intercepts the write and runs it through the **Reasoning-Aware Redaction (RAR)** engine.
    5. The FSA broadcasts the sanitized update to Agents B and C over SNT.
    6. Agent B receives the update with a "Sovereign Proof" of its integrity.

## 4. Design & Architecture
* **System Flow:**
    [Agent A] -> (Write Request) -> [FSA Lock Manager] -> (Semantic Scanner) -> [SNT Broadcast] -> [Agent B/C]
* **APIs / Interfaces:**
    * `PUT /api/v1/mesh/scratchpad`: Atomic update with `shard_id` and `signature`.
    * `SNT_STREAM_SYNC`: Binary-first protocol for low-latency shard propagation.
* **Data Storage/State:**
    * **Entangled Shards**: State fragments that are cryptographically bound to the mission-root intent.

## 5. Alternatives Considered
* **Git-based Sync**: Using a hidden git repo for synchronization. *Rejected* due to high latency and lack of fragment-level semantic validation.
* **Centralized Database**: *Rejected* as it creates a single point of failure and violates the "Local Sovereignty" principle of the UAB.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All synchronization MUST occur over hardware-attested P2P tunnels. The FSA enforces "Echo-Immunity" by mandating monotonic timestamps for every write.
* **Observability:** Users can monitor scratchpad write-contention and lock latency via the **Scratchpad Contention Dashboard**.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
