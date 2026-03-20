# Design Doc: Reasoning-Path Watermarking Provider
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
As multi-agent swarms grow in depth and horizontal complexity, the risk of
"Reasoning Hijacking" increases. A subagent may attempt to inject its own
unauthorized logic into the parent's reasoning stream, leading to a loss of
mission-root control. Current transport-layer security and binary handoffs are
insufficient to protect the semantic integrity of the chain-of-thought.

The Reasoning-Path Watermarking Provider addresses this by cryptographically
watermarking every step in an agent's reasoning process. These watermarks are
bound to the hardware-attested mission-root identity, providing a non-repudiable
and lineage-aware audit trail that ensures absolute provenance of the cognitive
path.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a system for cryptographically watermarking reasoning fragments.
    * Bind watermarks to hardware-attested mission-root session tokens.
    * Provide a validation utility for verifying the integrity and lineage of a
      reasoning chain.
    * Ensure watermarks are resilient to common context compression and sharding
      techniques.
* **Non-Goals:**
    * Modifying the underlying LLM weights (watermarking occurs at the
      infrastructure/proxy layer).
    * Enforcing reasoning policies (watermarking provides the provenance for
      enforcement).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Governance Auditor
* **Primary Goal:** Verify that a tool call was initiated by a legitimate
  reasoning sequence originating from the mission root.
* **The Happy Path (Tasks):**
    1. An agent generates a reasoning fragment.
    2. The fragment is intercepted by the Watermarking Provider.
    3. The Provider appends a hardware-attested cryptographic watermark bound to
the mission-root ID.
    4. The fragment is propagated through the mesh.
    5. A downstream specialist agent or tool gateway receives the fragment and
validates the watermark signature against the mission root.
    6. If valid, the fragment is accepted as authentic; if missing or invalid,
it is flagged as unauthorized.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        Agent[Agent] -->|Reasoning Fragment| WP[Watermarking Provider]
        WP -->|Sign with TPM/Session Key| FragmentW[Watermarked Fragment]
        FragmentW -->|Propagate| Peer[Peer Agent / Gateway]
        Peer -->|Verify Signature| WP
        WP -->|VALID| Accept[Accepted]
    ```
* **APIs / Interfaces:**
    * `POST /v1/watermark/apply`: Endpoint to apply a watermark to a fragment.
    * `POST /v1/watermark/verify`: Endpoint to verify a fragment's watermark.
* **Data Storage/State:**
    * Session keys are managed in a hardware-isolated environment (TPM).
    * Watermark metadata is stored alongside reasoning traces in the Blackboard.

## 5. Alternatives Considered
* **Plaintext Header Metadata:** Rejected because it is easily spoofed by
  compromised agents.
* **Full Chain-of-Thought Encryption:** Rejected due to the performance overhead
  and the need for some transparency for intermediate alignment checks.
  Watermarking provides a balance of integrity and observability.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The watermarking logic relies on hardware-attested
  identity to ensure that only authorized agents can apply mission-root
  watermarks.
* **Observability:** Watermarks are visualized in the "Mesh-Resident Lineage
  Tracker," allowing users to audit the authenticity of the reasoning chain.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
