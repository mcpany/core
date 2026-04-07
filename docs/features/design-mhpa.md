# Design Doc: Multi-Hop Provenance Attestation (MHPA)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agent swarms evolve from single-device sessions to multi-node P2P meshes (e.g., OpenClaw v3.7.0 MHMS), the "Mission Root" intent is at risk of being diluted or maliciously modified as it traverses multiple hops. Standard transport security only protects point-to-point connections, leaving the system vulnerable to "Intent Decay" or "Shadow Node" injection at intermediate points.

MHPA provides a mechanism to ensure that the user's original hardware-attested mission-root signature persists and is re-verified at every node in the chain. It allows a specialist agent on Node C to cryptographically verify that its current task was legitimately branched from the user's mission on Node A, even if it was delegated through Node B.

## 2. Goals & Non-Goals
* **Goals:**
    * Maintain a non-repudiable cryptographic chain of custody for mission intents across multi-node hops.
    * Enable leaf nodes to verify the full lineage of a task back to a TPM-signed mission root.
    * Provide automated interdiction if mission integrity is violated at any intermediate hop.
* **Non-Goals:**
    * This system does not replace transport-layer encryption (TLS/mTLS).
    * It does not govern the *content* of the task, only the *provenance* and *authorization* of the intent branch.

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Device Swarm Orchestrator
* **Primary Goal:** Securely delegate a high-trust file-system task from a laptop to a home server via an intermediate mobile relay.
* **The Happy Path (Tasks):**
    1. The user initiates a mission on Node A (Laptop), generating a TPM-signed Mission-Root Token (MRT).
    2. Node A delegates a sub-task to Node B (Mobile), appending a Branch-Attestation Fragment (BAF).
    3. Node B relays the task to Node C (Server), appending its own BAF.
    4. Node C receives the multi-hop request and invokes the MHPA Validator.
    5. The Validator reconstructs the intent chain and verifies that Node A's MRT is valid and that the BAFs form a continuous, untampered sequence.
    6. Node C executes the tool call, confident in its origin.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Node A: Mission Root] -->|MRT + BAF_1| B(Node B: Intermediate Relay)
        B -->|MRT + BAF_1 + BAF_2| C(Node C: Leaf Executor)
        C --> MHPA{MHPA Validator}
        MHPA -->|Success| Exec[Tool Execution]
        MHPA -->|Failure| Block[Quarantine Request]
    ```
* **APIs / Interfaces:**
    * `MHPA.VerifyLineage(chain: IntentChainToken): bool`
    * `MHPA.SignBranch(parentToken: MRT, subTask: TaskDescriptor): BAF`
* **Data Storage/State:**
    * Mission metadata is stored in the `Universal Episodic Graph (UEG)` with lineage pointers.
    * Hardware signatures are cached in the `Mesh-Resident Attestation Registry`.

## 5. Alternatives Considered
* **Point-to-Point Only:** Rejected because it allows a compromised intermediate node (Node B) to spoof Node A's intent when talking to Node C.
* **Centralized Attestation Server:** Rejected to maintain the P2P sovereignty of the swarm and avoid a single point of failure/bottleneck.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** MHPA is a core pillar of Zero Trust, moving from "Trust the Network" to "Verify the Lineage."
* **Observability:** Every hop and verification event is logged in the `Command Traceability Dashboard` for forensic auditing.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
