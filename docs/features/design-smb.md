# Design Doc: Sovereign Migration Broker (SMB)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
Production swarms are increasingly shifting from static local environments to hybrid-cloud "Teammate Meshes." A critical failure point in this shift is "Migration Drift," where agents lose their hardware-attested "Lineage of Authority" when moving between physical nodes (e.g., Laptop to Cloud GPU Cluster).

The Sovereign Migration Broker (SMB) facilitates the hardware-attested teleportation of agent sessions, ensuring that the mission-root intent and the cryptographically signed reasoning path remain atomic and non-repudiable across device boundaries.

## 2. Goals & Non-Goals
* **Goals:**
    * Facilitate hardware-locked "teleportation" of agent state between nodes.
    * Neutralize "Context Amnesia" during sandbox migration.
    * Enforce mission-bound attestation persistence across physical boundaries.
    * Integrate with zero-copy `memfd` handoffs for atomic state transfer.
* **Non-Goals:**
    * Providing a general VM/Container migration service (focus is on Agentic State).
    * Managing non-agentic filesystem synchronization.

## 3. Critical User Journey (CUJ)
* **User Persona:** Hybrid-Cloud Specialist
* **Primary Goal:** Seamlessly move an "Autonomous Data Scientist" agent from a local laptop to a cloud-based H100 cluster without restarting the mission.
* **The Happy Path (Tasks):**
    1. Local agent identifies a need for high-compute resources.
    2. SMB on the Laptop initiates a "Migration Proposal" to the Cloud SMB node.
    3. The Cloud node verifies the Laptop's hardware identity and mission-root authority.
    4. SMB triggers an AST (Atomic State Teleportation) event, capturing the reasoning path and Blackboard state into a TPM-signed `memfd` buffer.
    5. The buffer is tunneled securely to the Cloud node.
    6. Cloud SMB "unpacks" the state into the target enclave and resumes the agent.
    7. The agent continues reasoning instantly, with its lineage preserved.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        participant L as Laptop (SMB)
        participant C as Cloud (SMB)
        participant T as TPM/SEP
        L->>T: Sign State Snapshot
        L->>C: Migration Proposal (Signed)
        C->>C: Verify Hardware Identity
        L->>C: Atomic State Transfer (memfd)
        C->>C: Restore Context Shards
        C-->>L: Migration Complete
    ```
* **APIs / Interfaces:**
    * `smb.ProposeMigration(targetNode, missionID) -> ProposalID`: Requests a teleport.
    * `smb.CaptureState(proposalID) -> StateBlob`: Generates a TPM-signed snapshot.
    * `smb.RestoreState(stateBlob) -> SessionID`: Resumes an agent on the new node.
* **Data Storage/State:**
    * **Migration Log:** Audit trail of cross-node teleportation events.
    * **Enclave Shard Buffer:** Secure memory region for staging incoming state.

## 5. Alternatives Considered
* **Snapshot and Push (Git-based):** Rejected due to high latency and the risk of "Instruction Injection" in natural-language snapshots.
* **Remote Debugging Resumption:** Rejected as it requires continuous network connectivity and does not provide hardware-bound attestation of the recipient node.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** SMB mandates "Hardware-Locked Handshakes" at both ends. The target node must prove its own attestation status before receiving the state blob.
* **Observability:** Integrated with the "Service Mesh Topology Monitor" to show agents "moving" between nodes in real-time.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
