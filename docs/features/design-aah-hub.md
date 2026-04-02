# Design Doc: Attestation Aggregator Hub (AAH)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agent swarms become deeper and more horizontal, the coordination overhead is increasingly dominated by "Attestation Fatigue." Every inter-agent delegation currently requires a full hardware-bound (TPM/SEP) signature and verification cycle, often adding 100ms+ latency per hop. In a 10-hop subagent chain, this results in a full second of coordination tax.

The Attestation Aggregator Hub (AAH) aims to collapse these multiple subagent attestation tokens into a single, O(1) verifiable compressed proof, significantly reducing latency while maintaining Zero-Trust security.

## 2. Goals & Non-Goals
* **Goals:**
    * Aggregate multiple hardware-attested signatures into a single proof.
    * Reduce inter-agent coordination latency by at least 70% in deep swarms.
    * Maintain non-repudiable lineage for every sub-instruction.
    * Support "Proof Pruning" for stale mission branches.
* **Non-Goals:**
    * Replacing hardware enclaves (TPM/SEP); AAH relies on their base signatures.
    * Providing general-purpose data compression for non-security metadata.

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Frequency Swarm Orchestrator
* **Primary Goal:** Execute a 15-step subagent delegation chain with sub-200ms coordination overhead.
* **The Happy Path (Tasks):**
    1. Parent agent initiates a multi-hop delegation mission.
    2. Subagents A, B, and C generate their individual hardware-attested intent tokens.
    3. The AAH intercepts these tokens and performs an aggregation operation (e.g., using BLS signatures or Snarks).
    4. The resulting compressed proof is sent to the target specialist agent.
    5. The specialist agent performs a single O(1) verification on the compressed proof.
    6. Specialist execution begins immediately, bypassing the individual verification of all 15 parent signatures.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent 1 Signature] --> H[AAH Aggregator]
        B[Agent 2 Signature] --> H
        C[Agent 3 Signature] --> H
        H --> P[Compressed Proof]
        P --> V[One-Click Verifier]
        V --> S[Specialist Execution]
    ```
* **APIs / Interfaces:**
    * `aah.Aggregate(tokens[]) -> CompressedProof`: Merges a set of hardware tokens.
    * `aah.Verify(compressedProof, missionRoot) -> Boolean`: Validates the entire chain.
    * `aah.Prune(compressedProof, branchID) -> NewProof`: Removes a specific sub-branch from the proof.
* **Data Storage/State:**
    * **Aggregation Buffer:** Ephemeral storage for signatures pending compression.
    * **Root Identity Cache:** Local cache of verified mission-root public keys.

## 5. Alternatives Considered
* **Trust Propagation (TTL-based):** Rejected because it introduces "Identity Decay" where a compromised agent could squat on an un-attested lease.
* **Centralized Verification Server:** Rejected as it creates a single point of failure and violates the "Local-First" sovereignty principle.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The aggregation logic must be "Zero-Information Loss" regarding the lineage of authority. Any tampering with the aggregate must invalidate the entire proof.
* **Observability:** Integrated with the "Fast-Path Attestation Visualizer" in the UI to show latency savings.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
