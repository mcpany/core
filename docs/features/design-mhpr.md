# Design Doc: Multi-Hop Persistence Relay (MHPR)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agent swarms evolve from simple pairs to deep, multi-hop delegation chains (e.g., Agent A delegates to B, who delegates to C), a critical "Delegation Gap" has emerged. Current hardware-attested trust models often decay or require full re-attestation at each hop, introducing significant latency (100ms+ per hop) and leading to an 80% failure rate in deep chains. MCP Any needs to provide a mechanism that allows hardware-locked trust to persist and propagate across infinite hops without the prohibitive cost of repeated full handshakes.

## 2. Goals & Non-Goals
* **Goals:**
    * Facilitate sub-10ms trust propagation across delegation hops.
    * Maintain hardware-attested (TPM/Secure Enclave) lineage for every instruction.
    * Support cross-framework trust persistence (OpenClaw, Claude Code, AutoGen).
    * Provide non-repudiable audit trails for the entire delegation chain.
* **Non-Goals:**
    * Replacing the underlying hardware attestation primitives (TPM/SEP).
    * Providing long-term persistent storage for agent identities (focus is on session-bound delegation).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Cloud Swarm Orchestrator
* **Primary Goal:** Execute a 5-hop delegation chain (A->B->C->D->E) with hardware-level security and sub-50ms total coordination overhead.
* **The Happy Path (Tasks):**
    1. Agent A initiates a Mission Root and obtains a hardware-attested Root Lease from MCP Any.
    2. Agent A delegates a task to Agent B, including a cryptographically signed "Lineage Fragment."
    3. MCP Any validates the fragment and issues a "Persistence Relay Token" (PRT) to Agent B.
    4. Agent B uses the PRT to delegate to Agent C; MCP Any validates the lineage without a full hardware handshake.
    5. The chain continues to Agent E, with each hop maintaining the cryptographic link to Agent A's Root Lease.
    6. Agent E executes the final tool call; MCP Any verifies the complete non-repudiable chain.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        participant RA as Root Agent (A)
        participant MA as Middle Agent (B)
        participant TA as Target Agent (C)
        participant MHPR as MHPR Hub (MCP Any)
        participant TPM as Hardware TPM

        RA->>MHPR: Request Root Lease (Hardware Attested)
        MHPR->>TPM: Sign Root Intent
        TPM-->>MHPR: Root Lease Token
        MHPR-->>RA: Root Lease
        RA->>MA: Delegate Task + Lineage Fragment
        MA->>MHPR: Verify & Extend Lineage
        MHPR->>MHPR: Validate Fragment (Fast-Path)
        MHPR-->>MA: Persistence Relay Token (PRT)
        MA->>TA: Delegate Task + PRT
        TA->>MHPR: Final Tool Call + PRT
        MHPR->>MHPR: Validate Complete Chain
        MHPR->>Tool: Execute
    ```
* **APIs / Interfaces:**
    * `POST /v1/trust/lease/root`: Request initial hardware-attested lease.
    * `POST /v1/trust/lease/extend`: Extend an existing lineage fragment to a child agent.
    * `GET /v1/trust/lineage/verify`: Verify the complete chain of custody for a PRT.
* **Data Storage/State:**
    * Lineage fragments are stored in a session-bound, memory-mapped cache (memfd) for sub-millisecond retrieval.
    * Root leases are pinned to hardware-attested session IDs.

## 5. Alternatives Considered
* **Full Re-Attestation at Every Hop:** Rejected due to prohibitive latency (100ms+ per signature) which kills real-time coordination.
* **Flat Token Sharing:** Rejected as it lacks non-repudiable lineage; if Agent B is compromised, it could spoof Agent A's identity for unrelated tasks.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):**
    * PRTs are one-time-use or strictly time-bound (e.g., 60 seconds).
    * Lineage fragments include a hash of the *specific task* to prevent "Token Hijacking" for unauthorized actions.
* **Observability:**
    * Every hop in the lineage is logged with a cryptographically bound trace ID.
    * "Trust Decay" metrics are tracked to identify nodes with high attestation latency.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
