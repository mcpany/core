# Design Doc: HARC Validator
**Status:** Draft
**Created:** 2026-07-12

## 1. Context and Scope
As AI agent swarms move from linear task execution to deep, parallel reasoning chains, the risk of "Hallucination Cascades" has become a critical system-level threat. A single specialized subagent, through reasoning drift or adversarial injection, can produce high-confidence but erroneous state fragments that "smear" the shared blackboard, leading to catastrophic mission failure.

MCP Any needs to provide a deterministic, hardware-locked mechanism to validate high-entropy reasoning before it is allowed to trigger destructive or high-risk tool calls. The **Hardware-Attested Reasoning Consensus (HARC) Validator** solves this by mandating a multi-agent quorum for sensitive operations, anchored to hardware-attested identities.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a mandatory multi-agent quorum (Root, Auditor, Validator) for high-risk actions.
    * Mandate hardware-attested signatures for every reasoning fragment in the consensus loop.
    * Provide sub-millisecond coordination for HARC-compliant state handoffs.
    * Neutralize "Hallucination Cascades" by requiring cross-agent reasoning alignment.
* **Non-Goals:**
    * Validating every low-risk read-only tool call (to prevent MTTC explosion).
    * Replacing the primary agent's reasoning engine (HARC is an external governance layer).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Swarm Security Architect
* **Primary Goal:** Prevent an autonomous subagent from executing an unauthorized `rm -rf` on the production filesystem due to a hallucinated instruction.
* **The Happy Path (Tasks):**
    1. The primary agent proposes a high-risk tool call (e.g., `execute_destructive_command`).
    2. MCP Any intercepts the request and identifies the "HARC Required" policy flag.
    3. The gateway spawns/assigns an independent "Auditor" and "Validator" subagent.
    4. Each subagent reviews the parent's reasoning trace and provides a hardware-signed attestation token.
    5. The HARC Validator aggregates the tokens and verifies the hardware-bound quorum.
    6. The tool call is authorized and executed only upon reaching the 3/3 quorum.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        participant Agent
        participant Gateway as HARC Validator
        participant Auditor
        participant Validator
        participant Tool

        Agent->>Gateway: Propose High-Risk Tool Call (Signed Intent)
        Gateway->>Auditor: Request Attestation (Trace + Intent)
        Gateway->>Validator: Request Attestation (Trace + Intent)
        Auditor->>Gateway: Hardware-Signed "ACK"
        Validator->>Gateway: Hardware-Signed "ACK"
        Gateway->>Gateway: Verify Quorum (3/3)
        Gateway->>Tool: Execute Authorized Call
        Tool-->>Agent: Result
    ```
* **APIs / Interfaces:**
    * `/v1/harc/propose`: Endpoint for agents to submit reasoning for quorum review.
    * `/v1/harc/attest`: Endpoint for specialist auditors to submit signed approvals.
* **Data Storage/State:**
    * Consensus state is stored in an ephemeral, RAM-mapped "Quorum Buffer" to ensure low latency and zero persistence post-decision.

## 5. Alternatives Considered
* **Soft-Voting (Non-Hardware):** Rejected due to the risk of "Identity Shadowing" where one compromised process spoofs all three votes.
* **Single Auditor Model:** Rejected because a single specialist agent remains a single point of failure for hallucination propagation.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All votes must be accompanied by a TPM-bound session token. Any mismatch in hardware lineage triggers an immediate mission-wide revocation.
* **Observability:** Quorum decisions, timing, and reasoning alignment scores are exported to the Mesh-Resident Lineage Tracker.

## 7. Evolutionary Changelog
* **2026-07-12:** Initial Document Creation.
