# Design Doc: Dynamic Lease Escalation (DLE) Manager
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
In complex agent swarms, pre-flight authorization for every possible capability often leads to "Privilege Bloat" or "Authorization Fatigue." Agents frequently discover the need for specialized tools (e.g., `git push`, `cloud:deploy`) only after analyzing environment feedback. Re-authorizing the entire mission root for every sub-task is inefficient and breaks autonomous workflows.

The Dynamic Lease Escalation (DLE) Manager enables "Just-in-Time" (JIT) privilege upgrades. Subagents can request temporary, hardware-attested capabilities that are mathematically restricted to a specific task branch. This ensures that agents remain operational in dynamic environments while maintaining a "Zero-Privilege" baseline.

## 2. Goals & Non-Goals
* **Goals:**
    * Broker time-bound, task-specific privilege upgrades for subagents.
    * Enforce hardware-attested (TPM) boundaries for all escalated capabilities.
    * Automate the revocation of escalated leases upon task completion or mission decay.
    * Support "Recursive Lineage Validation" to ensure escalations don't exceed parent authority.
* **Non-Goals:**
    * Permitting permanent privilege upgrades.
    * Bypassing user-defined "Hard Deny" policies.
    * Modifying agent internal reasoning (DLE only gates the capability).

## 3. Critical User Journey (CUJ)
* **User Persona:** Autonomous DevOps Agent
* **Primary Goal:** Successfully fix a build error and push the correction to a protected branch without having pre-authorized `git push` access.
* **The Happy Path (Tasks):**
    1. The Agent fixed the code but hits a "Permission Denied" when attempting to push.
    2. The Agent issues a `DLE_REQUEST` for `git:push` limited to the current repository and task ID.
    3. The DLE Manager verifies the agent's lineage and current reasoning context via the `SRM Provider`.
    4. The Manager determines the request aligns with the mission root and issues a TPM-signed "Escalation Lease" valid for 5 minutes.
    5. The Agent pushes the code using the lease.
    6. The task completes, and the DLE Manager immediately revokes the lease and logs the event.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Subagent] -->|Permission Denied| B[DLE Manager]
        B -->|Verify Context| C[SRM Provider]
        B -->|Check Policy| D[Mission Manifest]
        C --> B
        D --> B
        B -->|Issue JIT Lease| E[TPM/Secure Enclave]
        E -->|Signed Token| A
        A -->|Authorized Tool Call| F[Target API]
    ```
* **APIs / Interfaces:**
    * `POST /v1/lease/escalate`: Requests a temporary capability upgrade.
    * `DELETE /v1/lease/{lease_id}`: Forcefully revokes an active lease.
* **Data Storage/State:**
    * Active leases are stored in a kernel-bound memory buffer managed by the EPM.

## 5. Alternatives Considered
* **Wide Pre-Authorization:** Rejected due to security risks (violates Principle of Least Privilege).
* **Manual HITL for every Fix:** Rejected as it creates a human bottleneck for high-frequency autonomous fixes.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All escalation tokens are hardware-bound. Any attempt to use a lease outside its specific task ID results in a hardware fault.
* **Observability:** Escalation events are tracked in the `JIT Privilege Lease Manager`.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
