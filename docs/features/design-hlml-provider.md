# Design Doc: Hardware-Locked Mission Lease (HLML) Provider
**Status:** Draft
**Created:** 2026-07-24

## 1. Context and Scope
With the shift toward "Agent Teams" (Claude Code) and multi-node meshes, coarse session-bound privileges are no longer sufficient. High-privilege operations must be tied to specific, user-authorized mission tasks and backed by hardware attestation (TPM) to prevent lateral movement by compromised subagents.

The HLML Provider issues TPM-signed, task-specific leases that are automatically revoked upon mission-root task completion, ensuring absolute lifecycle-bound sovereignty.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue hardware-attested (TPM) capability leases bound to specific Mission IDs.
    * Enforce automated, hardware-triggered revocation upon mission completion.
    * Support "Lease Chaining" for sub-delegated tasks.
    * Integrate with the HMLO for hierarchical management across nodes.
* **Non-Goals:**
    * Defining the agent's internal task-planning logic.
    * Providing a persistent database for expired leases; leases are ephemeral.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Swarm Developer
* **Primary Goal:** Grant a "Sudo" lease to a DevOps subagent that is hardware-locked to the "Update Web Server" task.
* **The Happy Path (Tasks):**
    1. Parent Mission-Root authorizes the "Update Web Server" task.
    2. HLML Provider generates a TPM-signed lease for the `sudo` tool, bound to the Task ID.
    3. The DevOps subagent executes the update using the hardware-locked lease.
    4. Upon task completion, the hardware root revokes the lease.
    5. Any subsequent attempts by the subagent to use `sudo` result in hardware-level rejection.

## 4. Design & Architecture
* **System Flow:**
    `Task Authorization` -> `TPM Lease Minting` -> `Subagent Execution` -> `Hardware Revocation Signal`
* **APIs / Interfaces:**
    * `hlml.MintLease(taskID, capabilities) -> TPMLease`: Generates a hardware-signed lease.
    * `hlml.VerifyLease(tpmLease) -> bool`: Hardware-level verification of lease validity.
    * `hlml.ChainLease(parentLease, subScope) -> SubLease`: Support for hierarchical sub-tasks.
* **Data Storage/State:**
    * **Secure Enclave (TPM):** Holds the master signing keys and active lease metadata.
    * **Lease Status Buffer:** High-speed in-memory state for real-time validation.

## 5. Alternatives Considered
* **Software-Signed JWTs:** Rejected because they are susceptible to exfiltration. HLML requires the secret keys to never leave the hardware enclave.
* **Manual HITL for Every Call:** Rejected as it creates prohibitive friction for autonomous swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Lineage is verified back to the user's primary hardware-root at every hop.
* **Observability:** Tracked in the "Mission Lease Manager" UI with real-time revocation status.

## 7. Evolutionary Changelog
* **2026-07-24:** Initial Document Creation.

### Update: 2026-07-25 - Integration with Hierarchical Lease Orchestrator
**Context:** Today's ecosystem shift toward "Lease Chaining" (Claude Code v3.2.1) requires the HLML Provider to support hierarchical inheritance.
**Architecture Adjustment:**
- Evolving `hlml.ChainLease` to natively support the HMLO protocol.
- Implementing subsetted capability enforcement at the hardware level.
**Security Impact:** Ensures that sub-leases cannot escalate beyond parent bounds, even if the sub-agent is compromised.
