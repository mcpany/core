# Design Doc: Attestation Compression Broker (ACB)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agent meshes scale across multiple hardware nodes, the overhead of verifying individual hardware-attested signatures for every inter-node coordination fragment has become a critical performance bottleneck. In deep, multi-node swarms, the "Attestation Tax" can account for up to 40% of coordination latency.

The Attestation Compression Broker (ACB) implements the Recursive Attestation Compression (RAC) standard to merge multiple hardware signatures into a single, high-performance proof. This ensures that trust remains non-repudiable while drastically reducing the handshake payload and verification overhead.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement the RAC standard to compress multi-node hardware signatures.
    * Reduce inter-node coordination latency by at least 50% in deep swarms.
    * Provide a unified interface for verifying compressed attestation chains.
    * Integrate with the AMT Broker for high-performance mesh tunneling.
* **Non-Goals:**
    * Replacing the underlying hardware attestation (TPM/SEP) primitives.
    * Managing non-agentic network traffic.
    * Serving as a general-purpose compression utility.

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Node Swarm Developer
* **Primary Goal:** Maintain sub-100ms coordination latency in a 5-node distributed agent swarm.
* **The Happy Path (Tasks):**
    1. A subagent on Node A delegates a task to a specialist on Node E, crossing three intermediate nodes (B, C, D).
    2. Each intermediate node adds its hardware-attested signature to the intent lineage.
    3. The ACB on Node D intercepts the multi-signature lineage and applies RAC to compress it into a single proof fragment.
    4. Node E receives the compressed proof and uses its local ACB to verify the entire "Chain of Command" in a single operation.
    5. Coordination proceeds with minimal latency compared to linear signature verification.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Intent Lineage] --> B[Node A Signature]
        B --> C[Node B Signature]
        C --> D[Node C Signature]
        D --> E{ACB (RAC)}
        E -->|Compressed Proof| F[Target Node]
        F --> G[Unified Verification]
    ```
* **APIs / Interfaces:**
    * `acb.CompressLineage(lineageTokens) -> CompressedToken`: Merges multi-node signatures.
    * `acb.VerifyCompressed(compressedToken) -> LineageStatus`: Validates the entire chain.
    * `acb.RefreshProof(compressedToken, newNodeSignature) -> CompressedToken`: Appends a new node to an existing compressed proof.
* **Data Storage/State:**
    * **RAC Scheme Registry:** Metadata for supported compression algorithms and security levels.
    * **Lineage Cache:** Short-term storage for recently verified compressed proofs to accelerate repeated handshakes.

## 5. Alternatives Considered
* **Linear Signature Chains:** Rejected due to prohibitive latency and payload size in deep swarms (8+ nodes).
* **Centralized Attestation Service:** Rejected because it introduces a single point of failure and violates the decentralized mesh sovereignty principle.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Compression must be mathematically sound (e.g., using aggregate signatures or zero-knowledge proofs) to prevent signature spoofing or "Lineage Splicing."
* **Observability:** Integrated with the "Attestation Bridge Monitor" in the UI to visualize compression ratios and latency gains.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
