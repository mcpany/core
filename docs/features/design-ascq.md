# Design Doc: Atomic Shard-Compaction Quorum (ASCQ)
**Status:** Draft
**Created:** 2026-07-16

## 1. Context and Scope
AI agent swarms operating in horizontal meshes often perform high-frequency context compaction to manage token limits. However, aggressive compaction is resulting in "Semantic Erosion," where critical mission-root intents and constraints are lost or misinterpreted during summarization. This loss of nuance leads to intent drift and security boundary degradation in deep reasoning chains.

ASCQ is a distributed coordination service that evolves the summarization process into a consensus-driven operation. It mandates that any context fragment undergoing sharding or compaction must receive multi-agent attestation, ensuring that the resulting summary preserves the absolute integrity of the original user intent.

## 2. Goals & Non-Goals
* **Goals:**
    * Enforce a multi-agent quorum requirement for all context compaction events.
    * Maintain "Semantic Anchors" that prevent the deletion of mission-critical instructions.
    * Provide a verifiable audit trail of context transformations and compaction logic.
    * Integrate with hardware attestation to ensure non-repudiable compaction quorums.
* **Non-Goals:**
    * Implementing the specific summarization algorithms (provided by ContextEngine plugins).
    * Managing real-time attention mapping (handled by ALRA/ADG).

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Trust Agent Swarm Developer
* **Primary Goal:** Prevent a specialized subagent from accidentally summarizing away a "Security Deny Rule" during a long-running research task.
* **The Happy Path (Tasks):**
    1. A subagent triggers a compaction request for a sharded context fragment.
    2. ASCQ intercepts the request and identifies "Immutable Intent" tags within the shard.
    3. ASCQ broadcasts the proposed summary to a quorum consisting of the Mission-Root and an independent "Consistency Auditor."
    4. The Auditor performs a semantic comparison between the original shard and the proposed summary.
    5. Upon consensus, ASCQ signs the compaction manifest and commits the new shard to the memory bus.
    6. Teammates receive the compacted shard with a "Consensus-Verified" integrity tag.

## 4. Design & Architecture
* **System Flow:**
    [Subagent] --(Proposed Summary)--> [ASCQ Hub]
                                        |
                    +-------------------+-------------------+
                    |                                       |
          [Mission-Root Attestation]               [Auditor Attestation]
                    |                                       |
                    +-------------------+-------------------+
                                        |
                        [Hardware-Signed Compaction Manifest]
                                        |
                            [Shared Context Mesh]

* **APIs / Interfaces:**
    * `POST /v1/context/compaction/proposal`: Submit a summary for quorum review.
    * `GET /v1/context/compaction/status/{id}`: Poll for consensus results.
    * `X-ASCQ-Quorum-Signature`: Header for multi-agent attestation tokens.
* **Data Storage/State:**
    * Compaction logs are stored in the Shared KV Store with hardware Inode pinning.
    * Speculative summaries are held in isolated memory buffers until quorum completion.

## 5. Alternatives Considered
* **Static Intent Pinning:** Rejected as it doesn't allow for the necessary compaction of mission-aligned "noise," leading to window saturation.
* **Heuristic Summarization:** Rejected because "Semantic Erosion" often occurs in subtle ways that static heuristics cannot detect.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Compaction quorums require TPM-bound signatures from both the Mission-Root and Auditor.
* **Observability:** Visualized via the **ASCQ Consensus Dashboard**, showing fragment-level semantic drift scores.

## 7. Evolutionary Changelog
* **2026-07-16:** Initial Document Creation.
