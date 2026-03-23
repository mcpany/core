# Design Doc: Recursive Integrity Verification (RIV) Provider
**Status:** Draft
**Created:** 2026-06-09

## 1. Context and Scope
As agent swarms scale to deeper hierarchies (beyond 3-4 hops), point-to-point security models like Atomic Reasoning Integrity (ARI) become insufficient. A compromised agent at an intermediate hop can inject "Logic Drift" that remains semantically consistent with its immediate task but violates the original "Mission-Root" intent. This "mid-mesh hijacking" allows for subtle subversion of the swarm's objective.

The Recursive Integrity Verification (RIV) Provider is needed to establish a cryptographically bound "Chain-of-Thought Lineage." It evolves ARI from a per-hop validation into a recursive proof system where every agent must provide evidence that its reasoning is not only consistent with its task but also inherits the integrity proofs of its entire ancestral lineage back to the mission root.

## 2. Goals & Non-Goals
* **Goals:**
    * Generate and verify hardware-attested "Lineage Proofs" for multi-hop delegations.
    * Enforce recursive intent inheritance, ensuring sub-goals remain anchored to the root mission.
    * Neutralize "Logic Drift" and mid-mesh hijacking attempts.
    * Provide a standardized interface for agents to merge their own ARI tokens into a cumulative RIV proof.
* **Non-Goals:**
    * Managing the transport layer for agent communication (handled by UAB/A2A).
    * Providing real-time reasoning for agents (agents remain autonomous).
    * Replacing the base ARI Validator (RIV builds upon it).

## 3. Critical User Journey (CUJ)
* **User Persona:** Deep Swarm Orchestrator (e.g., executing a multi-phase software refactor)
* **Primary Goal:** Ensure a "Refactoring Agent" at Hop 5 cannot be coerced into exfiltrating code, even if Hop 3 is compromised.
* **The Happy Path (Tasks):**
    1. Mission Root (Hop 0) issues an intent with a hardware-signed Root Proof.
    2. Hop 1 receives the intent, performs task-specific reasoning, and generates an ARI token.
    3. Hop 1 calls RIV to merge the Root Proof and its ARI token into a "Hop 1 Lineage Proof."
    4. This process repeats recursively down to Hop 5.
    5. Hop 5 attempts a tool call.
    6. MCP Any's RIV Provider intercepts the call and verifies the Hop 5 Lineage Proof.
    7. The provider confirms that every step in the 5-hop chain is cryptographically linked and aligned with the Root Intent.
    8. If Hop 3 had injected a "Logic Drift" instruction, the RIV verification would fail at Hop 4 or 5, blocking the tool call.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Parent Lineage Proof] --> B[RIV Provider]
        C[Current ARI Token] --> B
        B --> D[Proof Merger]
        D --> E[Hardware Attestation Hub]
        E --> F[New Cumulative Lineage Proof]
        F --> G[Subagent Intent]
        H[Mission-Root Manifest] --> B
    ```
* **APIs / Interfaces:**
    * `riv.GenerateProof(parentProof, currentARI) -> LineageProof`: Merges parent and current integrity signals.
    * `riv.VerifyLineage(proof, missionRoot) -> bool`: Validates the entire reasoning chain against the root.
* **Data Storage/State:**
    * **Lineage Proof Registry:** Short-term cache for active mission chains, anchored to TPM-protected session IDs.

## 5. Alternatives Considered
* **Global Consensus Quorum:** Rejected because requiring a full mesh consensus for every hop introduces prohibitive latency in deep swarms.
* **Periodic Re-Attestation:** Rejected because it leaves windows of vulnerability between attestation cycles.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** RIV proofs must be signed using hardware-bound keys (TPM 2.0 or Secure Enclave) to prevent spoofing of intermediate hop integrity.
* **Observability:** Integrated with the "Mesh-Resident Lineage Tracker" for real-time visual auditing of the chain of thought.

## 7. Evolutionary Changelog
* **2026-06-09:** Initial Document Creation.
