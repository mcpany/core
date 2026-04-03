# Design Doc: Speculative Shard Prefetcher (SSP)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
Horizontal Agent Teams (e.g., Claude Code Agent Teams) often suffer from "Cognitive Stall" when parallel teammates attempt to access or mutate the same context shards on the Blackboard. Traditional locking mechanisms introduce significant latency (5s+) during conflict resolution.

The Speculative Shard Prefetcher (SSP) addresses this by predicting teammate trajectories and speculatively pre-loading/sandboxing relevant context shards. This allows teammates to continue reasoning against "Probabilistic Buffers" while the central arbiter resolves state conflicts in the background.

## 2. Goals & Non-Goals
* **Goals:**
    * Predict and pre-fetch context shards based on real-time agent reasoning intent.
    * Provide isolated "Speculative Sandboxes" for parallel teammate mutations.
    * Neutralize "Cognitive Stall" in high-density horizontal swarms.
    * Integrate with the Probabilistic Buffer Hardening (PBH) middleware for security.
* **Non-Goals:**
    * Automatically merging conflicting speculative states without mission-root approval.
    * Serving as a general-purpose cache for non-context data.
    * Managing low-level database transactions outside the Blackboard.

## 3. Critical User Journey (CUJ)
* **User Persona:** Horizontal Swarm Architect
* **Primary Goal:** Enable 10+ teammates to coordinate on a shared codebase without coordination locks.
* **The Happy Path (Tasks):**
    1. Teammate A begins reasoning about a file edit in Shard X.
    2. SSP analyzes Teammate A's internal monologue and predicts that Teammate B will soon require Shard Y (a dependency).
    3. SSP speculatively pre-fetches Shard Y and prepares an isolated sandbox for Teammate B.
    4. When Teammate B requests Shard Y, it is already available in the local speculative buffer.
    5. Teammate B performs "Optimistic Speculation" against the buffer while Shard Y is being verified for mission-root alignment.
    6. Once verified, the speculative state is committed to the primary mission-root Blackboard.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent Monologue] --> B[Intent Analyzer]
        B --> C[Trajectory Predictor]
        C --> D[Shard Prefetcher]
        D --> E[Speculative Sandbox]
        E --> F[Probabilistic Buffer]
        F --> G{Arbiter Validation}
        G -->|Success| H[Blackboard Commit]
        G -->|Failure| I[Buffer Purge]
    ```
* **APIs / Interfaces:**
    * `ssp.PrefetchShard(shardID, predictedIntent) -> SandboxID`: Pre-loads a shard into isolation.
    * `ssp.ExecuteSpeculative(sandboxID, mutations) -> BufferStatus`: Performs optimistic reasoning.
    * `ssp.ValidateAndCommit(bufferID) -> CommitResult`: Merges speculative state after mission-root attestation.
* **Data Storage/State:**
    * **Speculative State Store:** Ephemeral, hardware-locked storage for un-attested context mutations.
    * **Intent Heatmap:** Real-time tracking of shard access patterns to improve prediction accuracy.

## 5. Alternatives Considered
* **Global Sequential Locking:** Rejected due to prohibitive "Cognitive Stall" in parallel swarms.
* **Optimistic Concurrency Control (OCC) without Prefetching:** Rejected because it still results in reasoning restarts when conflicts are detected post-execution. Prefetching reduces the conflict window.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Probabilistic Buffer Hardening (PBH) ensures that speculative state cannot leak into the primary intent loop until verified.
* **Observability:** Integrated with the "Shard-Aware Performance Heatmap" in the UI to visualize prefetch hit rates and speculation latency.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
