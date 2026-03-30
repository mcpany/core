# Design Doc: Hardware-Attested Intent Recovery (HAIR) Provider
**Status:** Draft
**Created:** 2026-07-12

## 1. Context and Scope
Mission stability in autonomous swarms is currently threatened by "Ephemeral State Loss." When an agent process crashes or a mesh node fails, the cryptographically bound "Chain of Thought" is often severed, requiring a full mission restart. This is particularly catastrophic in multi-day reasoning tasks where "Mission-Root Continuity" is essential for legal and security compliance.

The Hardware-Attested Intent Recovery (HAIR) Provider moves MCP Any toward "Persistent Mission Sovereignty." By utilizing TPM-bound state snapshots and monotonic lineage counters, it allows agents to reconstruct and resume their cognitive path after a failure, ensuring that the mission root remains authoritative regardless of infrastructure churn.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement hardware-locked "Intent Snapshots" that are immune to system-clock tampering or rollback attacks.
    * Provide a monotonic recovery bridge that verifies the "Chain of Reason" before mission resumption.
    * Integrate with the Mission-Root Continuity Provider (MRCP) for cross-node recovery.
    * Neutralize "Shadow-Attestation" vulnerabilities by normalizing TPM-to-System clock drift during recovery.
* **Non-Goals:**
    * Providing general-purpose backup/restore for application data (focus is on the reasoning lineage and task-claim state).
    * Recovering "Blackboard" data that was not committed before the crash (handled by atomic state rollbacks).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Compliance Officer
* **Primary Goal:** Ensure an autonomous "Audit Swarm" resumes exactly where it left off after a datacenter power event, without losing the hardware-attested lineage of its findings.
* **The Happy Path (Tasks):**
    1. An agent team is halfway through a 10,000-file security audit.
    2. The primary orchestration node experiences a hardware failure.
    3. Upon node reboot, the HAIR Provider initializes and pulls the last TPM-signed "Intent Snapshot."
    4. The Provider verifies the monotonic counter and validates the "Chain of Thought" against the Hardware-Attested Mission Manifest (HAMM).
    5. The mission is resumed on a new node; the HAIR Provider broadcasts a "Mission-Resumption Certificate" to the mesh.
    6. Teammates re-attest to the resumed root and continue the audit with zero "Cognitive Amnesia."

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        Agent[Active Agent] -->|Periodic Snapshot| HAIR[HAIR Provider]
        HAIR -->|TPM Sign| Storage[Secure Snapshot Storage]
        Failure((Node Crash)) -.-> Recovery
        Recovery[Node Init] -->|Pull Snapshot| HAIR
        HAIR -->|Verify Monotonicity| TPM[Hardware TPM]
        TPM -->|Resumption Token| Agent
    ```
* **APIs / Interfaces:**
    * `POST /hair/snapshot`: Submit an intent-bound state fragment for hardware-locked storage.
    * `POST /hair/recover`: Request mission reconstruction after a failure signal.
    * `GET /hair/lineage`: View the hardware-verified resumption history of a mission.
* **Data Storage/State:**
    * "Intent Snapshots" are stored in an encrypted, off-node resilience shard, with keys bound to the mission-root identity.

## 5. Alternatives Considered
* **JSON-based Checkpointing:** Rejected as it is vulnerable to "Chain-of-Thought Spoofing" and lack hardware-level integrity proofs.
* **Full VM Snapshots:** Rejected due to excessive storage overhead and slow recovery times for high-frequency reasoning swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Recovery is only granted to a node that can prove the same mission-bound attestation lineage as the failed node.
* **Observability:** Resumption events are flagged in the Mesh Audit Log as high-severity sovereignty transitions.

## 7. Evolutionary Changelog
* **2026-07-12:** Initial Document Creation.
