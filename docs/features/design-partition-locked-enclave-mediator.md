# Design Doc: Partition-Locked Enclave Mediator
**Status:** Draft
**Created:** 2026-05-27

## 1. Context and Scope
With the introduction of Mission-Root Branching (MRB), agent swarms can decompose complex tasks into hierarchical sub-missions. However, current hardware isolation models often use a shared enclave for all sub-missions under a single mission root. This leads to "Privilege Smearing," where a sub-mission root can accidentally inherit high-privilege policies (e.g., admin keys) from a sibling branch if TPM handles or memory segments are not strictly recycled or partitioned.

The Partition-Locked Enclave Mediator (PLEM) implements physical memory partitioning for each sub-mission root. It ensures that every branch in the intent tree is assigned a unique, hardware-isolated enclave segment that is physically inaccessible to other branches. MCP Any acts as the secure PLEM, managing the lifecycle and partitioning of these hardware resources to ensure that sub-mission safety remains absolute.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement hardware-enforced memory partitioning for sub-mission roots (MRB).
    * Provide a mandatory "Partition Gate" for cross-branch state access.
    * Prevent "Privilege Smearing" by physically isolating policy segments in the TPM.
    * Support "Atomic Branch Purging" where an entire enclave partition is wiped upon branch pruning.
* **Non-Goals:**
    * Replacing general-purpose OS memory management (PLEM is scoped to agent intent enclaves).
    * Managing non-hardware-backed memory segments (scoped to BSH and SCS shards).

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Security Swarm Orchestrator
* **Primary Goal:** Ensure a "Deployment Branch" with production access cannot leak its admin credentials to a "Testing Branch" under the same mission root.
* **The Happy Path (Tasks):**
    1. The Orchestrator defines a hierarchical mission with "Prod" and "Test" sub-branches.
    2. MCP Any PLEM provisions two physically isolated enclave partitions in the TPM.
    3. The Deployment subagent initializes in the "Prod Partition" and receives hardware-locked production keys.
    4. The Testing subagent initializes in the "Test Partition" with restricted developer keys.
    5. The Testing subagent attempts to read a memory shard belonging to the Prod Partition.
    6. Because the partitions are hardware-locked, the request results in an immediate TPM "Segmentation Fault."
    7. PLEM logs the violation and triggers an IRA (Intent Re-Alignment) handshake for the Testing branch.

## 4. Design & Architecture
* **System Flow:**
    [Sub-Mission Root] <--> [PLEM Broker] <--> [Partitioned HSM/TPM Memory]
    1. Branch initialization request includes `Partition:Isolated` flag.
    2. PLEM Broker allocates a unique Hardware Slot and Memory Range for the branch.
    3. All subsequent tool calls for that branch are cryptographically bound to the partition-ID.
    4. Hardware logic blocks any cross-ID memory or handle access.
* **APIs / Interfaces:**
    * `InitializePartitionedBranch(parent_id, branch_id) -> partition_handle`
    * `RequestCrossPartitionAccess(source_id, target_id, attestation) -> bool`
* **Data Storage/State:**
    * Partition manifests are stored in a hardware-protected segment of MCP Any internal state.
    * Master keys for partition-binding reside in the Secure Enclave.

## 5. Alternatives Considered
* **Logical Resource Labeling:** Rejected because it relies on software-level enforcement which can be bypassed by kernel-level escapes.
* **Full HSM Virtualization:** Rejected; too much overhead for ephemeral agent swarms. Partition-locking provides the required security with better performance.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Partitions are physically isolated. No cross-branch access is allowed without explicit, hardware-attested bridging.
* **Observability:** Track "Partition Violation" attempts as high-fidelity security alerts in the UI.

## 7. Evolutionary Changelog
* **2026-05-27:** Initial Document Creation.
