# Design Doc: Quorum-Bound Summarization (QBS)
**Status:** Draft
**Created:** 2026-07-06

## 1. Context and Scope
As AI agent swarms move toward horizontal teammate coordination with context windows exceeding 1M tokens, "Attention Drift" and aggressive context compaction have become primary stability risks. Modern summarizer agents often over-optimize for token savings, leading to "Mission-Root Erasure"—the loss of core mission constraints during long-running sessions.

MCP Any needs to act as the authoritative governor for context lifecycle. QBS introduces a consensus-driven compaction model where state fragments can only be summarized after a multi-agent quorum validates that the summary remains anchored to the user's primary objectives.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a multi-agent quorum protocol for context compaction events.
    * Prevent "Mission-Root Erasure" by mandating Auditor attestation for summaries.
    * Support "Optimistic Summarization Commits" to minimize coordination latency.
    * Maintain a verifiable lineage of context state transitions.
* **Non-Goals:**
    * Replacing the underlying LLM's internal attention mechanism.
    * Providing a general-purpose agent orchestration framework.

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Swarm Security Architect
* **Primary Goal:** Ensure that a deep agent chain (10+ hops) does not "forget" data privacy constraints during autonomous context compaction.
* **The Happy Path (Tasks):**
    1. The primary agent triggers a context compaction event due to 80% window utilization.
    2. MCP Any intercepts the request and identifies a "High-Trust" mission root.
    3. The compacting agent generates a proposed summary shard.
    4. MCP Any routes the proposal to an independent "Security Auditor" subagent.
    5. The Auditor agent compares the summary against the hardware-attested Mission Manifest.
    6. Both Mission-Root and Auditor provide TPM-signed approval tokens.
    7. MCP Any commits the summarized shard to the teammate bus and allows reasoning to continue.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Agent->>QBS Hub: Request Compaction (Provisional Summary)
        QBS Hub->>Auditor: Validate Summary against Mission Root
        QBS Hub-->>Agent: Optimistic Acknowledgement (Speculative Reasoning)
        Auditor-->>QBS Hub: Signed Attestation Token
        MissionRoot-->>QBS Hub: Signed Attestation Token
        QBS Hub->>Teammate Bus: Commit Summarized Shard
    ```
* **APIs / Interfaces:**
    * `POST /v1/context/summarize`: Submits a summary proposal.
    * `GET /v1/quorum/status/{proposal_id}`: Polls for attestation status.
    * `X-MCP-Quorum-Token`: Header for hardware-attested consensus signatures.
* **Data Storage/State:**
    * Summaries are stored in the Shared KV Store (Blackboard) using "Speculative" vs "Committed" tags.
    * Quorum metadata is pinned to hardware Inodes via EMA (Enclave-local Metadata Attestation).

## 5. Alternatives Considered
* **Static Token-Bound Rules:** Rejected because they lack the semantic nuance to distinguish between "noise" and "critical constraints."
* **Single-Agent Gating:** Rejected due to the risk of "Summarization Ghosting" seen in un-audited OpenClaw sessions.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Quorum tokens are hardware-attested (TPM/SEP). Any mismatch in signatures triggers an immediate "Mission Failure" signal and state rollback.
* **Observability:** Real-time visualization via the **Summarization Quorum Hub** dashboard in the UI.

## 7. Evolutionary Changelog
* **2026-07-06:** Initial Document Creation.
