# Design Doc: Teammate Integrity Token (TIT) Broker
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
In horizontal teammate coordination swarms (e.g., Claude Code Agent Teams), the shared task list (Blackboard) and mailbox have become primary attack surfaces. 'Teammate Impersonation' allows a compromised specialist agent to misdirect other teammates by reporting fake task completions or hijacking mission-root instructions.

The Teammate Integrity Token (TIT) Broker provides the authoritative infrastructure for issuing and verifying hardware-attested, session-bound integrity tokens for every inter-teammate status update or task claim.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement an authoritative broker for issuing session-bound TITs to horizontal teammates.
    * Mandate TIT-attestation for every task claim or status update in the shared mailbox.
    * Ensure absolute non-repudiation of inter-teammate coordination in sharded meshes.
    * Integrate with hardware-bound identity tokens (SMI) to ensure token lineage.
* **Non-Goals:**
    * Directly managing LLM inference (handled by providers).
    * Enforcing low-level transport security (handled by the Named-Pipe/WebSocket layer).
    * Sanitizing binary state (handled by the WASM-BSH Sanitizer).

## 3. Critical User Journey (CUJ)
* **User Persona:** Horizontal Swarm Orchestrator
* **Primary Goal:** Ensure the absolute integrity of task claiming and reporting in shared teammate coordination channels.
* **The Happy Path (Tasks):**
    1. Teammate Agent is spawned and receives an SMI identity token.
    2. Teammate Agent requests a TIT from the TIT Broker for the current mission session.
    3. Teammate Agent claims a task from the shared mailbox, attaching its TIT to the request.
    4. TIT Broker (integrated with the Mailbox Guard) verifies the token's validity and mission-root lineage.
    5. The task claim is recorded as a non-repudiable event on the Blackboard.
    6. Other teammates can verify the TIT to ensure the claim is legitimate.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Teammate Agent] --> B[TIT Broker]
        B --> C[Hardware Identity Validator]
        C --> D[Mission-Root Lineage Check]
        D --> E[Issue Integrity Token]
        E --> F[Attach to Task Claim/Update]
        F --> G[Mailbox Guard Verification]
    ```
* **APIs / Interfaces:**
    * `tit.IssueToken(missionToken, agentID) -> TIT`: Issues a new integrity token.
    * `tit.VerifyToken(token, missionToken) -> bool`: Validates a token's integrity and lineage.
* **Data Storage/State:**
    * **TIT Registry:** A persistent, cryptographically signed log of all active and expired integrity tokens, stored in a TEE-protected segment of the Blackboard.

## 5. Alternatives Considered
    * **Static Token Signing:** Rejected because it does not provide the hardware-bound, session-specific non-repudiation required for horizontal meshes.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The TIT Broker must utilize Hardware-Attested Identity (SMI) and TPM primitives to ensure tokens cannot be spoofed or reused across sessions.
* **Observability:** Integrated with the 'Mesh-Resident Lineage Tracker' for real-time visualization of teammate coordination events.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
