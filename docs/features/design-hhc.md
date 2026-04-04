# Design Doc: Headless Handoff Continuity (HHC) Bridge
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agent workflows move to remote, headless environments (Claude Code Dispatch), a critical gap exists when ownership of a session needs to transition from an automated trigger (CI/CD) to a human reviewer, or between different remote controllers. Without a continuity bridge, the "Mission Root" intent and granular reasoning state (Blackboard) are often lost or fragmented during the handoff.

The HHC Bridge provides the infrastructure to "hibernate" and "resume" agent sessions across network and controller boundaries, ensuring that the new controller inherits the full cryptographic lineage and context of the mission.

## 2. Goals & Non-Goals
* **Goals:**
    * Facilitate the transfer of hardware-attested mission tokens between controllers.
    * Synchronize the "Internal Monologue" and Blackboard shards during controller rotation.
    * Provide RDH-native WebSocket hooks for real-time steering re-attachment.
    * Support "Mission Checkpointing" to allow rollbacks after a handoff.
* **Non-Goals:**
    * Automating the handoff decision (this is triggered by RDH or ACSM).
    * Synchronizing the entire host filesystem (HHC focuses on the agent reasoning state).

## 3. Critical User Journey (CUJ)
* **User Persona:** On-Call SRE
* **Primary Goal:** Take over a failing autonomous migration task initiated by a headless worker and steer it to completion without losing the previous 2 hours of reasoning.
* **The Happy Path (Tasks):**
    1. A headless worker in RDH triggers an "Uncertainty Escalation."
    2. HHC Bridge creates a "Continuity Snapshot" of the mission-root state.
    3. SRE receives an alert and re-attaches to the session via the RDH Console.
    4. HHC Bridge verifies the SRE's hardware token and restores the reasoning shards.
    5. SRE reviews the interactive "Reasoning Lineage" and issues a corrective intent.
    6. Mission continues seamlessly under the new controller.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Headless Controller] -->|Handoff Signal| B[HHC Bridge]
        B -->|State Checkpoint| C[(Shared KV Store)]
        D[Human Controller] -->|Attach Request| B
        B -->|Verify Lineage| E[Hardware Enclave]
        E -->|Success| B
        B -->|Restore State| D
    ```
* **APIs / Interfaces:**
    * `ContinuityProvider`: `CreateCheckpoint(sessionID string) (CheckpointID, error)`
    * `HandoffBroker`: `RotateController(oldID, newID string, token Attestation) error`
* **Data Storage/State:**
    * Continuity manifests are stored in the `mrcp_continuity` vault in SQLite.

## 5. Alternatives Considered
* **Stateless Handoffs:** Rejected because agents lose the "Why" behind previous tool calls, leading to redundant execution and token waste.
* **Persistent SSH Sessions:** Rejected as they do not scale across multi-cloud environments and lack fine-grained intent-locking.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Controller rotation requires a "Chain of Trust" attestation where the new controller must prove authority over the original mission-root.
* **Observability:** "Controller Latency" and "Handoff Integrity" scores are tracked in the Swarm Coherence Dashboard.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
