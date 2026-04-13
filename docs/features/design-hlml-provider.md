# Design Doc: Hardware-Locked Mission Leases (HLML) Provider
**Status:** Draft
**Created:** 2026-07-24

## 1. Context and Scope
The transition to "Lifecycle-Bound Agency" demands that privilege is not just time-bound, but strictly task-bound. Static capability tokens are insufficient for horizontal Agent Teams, where a specialist may require high-privilege access (e.g., `sudo`) for a specific sub-task but should not retain that access for the remainder of the session.

The Hardware-Locked Mission Leases (HLML) Provider issues TPM-signed, task-specific leases that are automatically revoked by the hardware root upon detection of mission-root task completion, neutralizing persistent privilege escalation and "Credential Squatting."

## 2. Goals & Non-Goals
* **Goals:**
    * Issue TPM-signed capability leases cryptographically bound to a specific `TaskID` and `MissionRoot`.
    * Implement autonomous revocation based on task-completion signals from the coordination hub.
    * Ensure leases are non-transferable between agent framework processes.
    * Support "Mission-Bound Hardware Attestation" for remote tool calls (via AMT).
* **Non-Goals:**
    * Managing the user's primary identity (this is handled by the FSI Provider).
    * Providing long-term persistent credentials.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Grant a "DevOps Specialist" agent temporary `write` access to a restricted config directory for one specific PR update.
* **The Happy Path (Tasks):**
    1. The Orchestrator delegates the "Config Update" task to the DevOps Specialist.
    2. The HLML Provider issues a TPM-signed lease: `[Task-102, /etc/app/config, RW, Expires: Task-102-Complete]`.
    3. The Specialist agent presents the lease to the MCP Any filesystem tool.
    4. MCP Any verifies the TPM signature and the active status of Task-102.
    5. The Specialist completes the update and reports task completion to the coordination hub.
    6. The HLML Provider receives the completion signal and immediately revokes the lease in the hardware root.
    7. Subsequent attempts by the Specialist to use the same lease are blocked.

## 4. Design & Architecture
* **System Flow:**
    `[Task Orchestrator] -> (Task Event) -> [HLML Provider] -> (TPM Signature) -> [Lease Token] -> [Specialist Agent] -> [Secure Tool Gateway]`
* **APIs / Interfaces:**
    * `hlml.IssueLease(taskID, capability) -> LeaseToken`
    * `hlml.RevokeLease(taskID) -> bool`
    * `hlml.VerifyLease(token) -> bool`
* **Data Storage/State:**
    * **Hardware Lease Registry**: Secure, kernel-bound memory or TPM-backed NVRAM for tracking active lease states.

## 5. Alternatives Considered
* **Time-Bound JWTs**: Rejected because task completion time is non-deterministic, leading to either premature expiration or "Squatting" windows.
* **Manual Revocation**: Rejected due to high latency and risk of human error in complex swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** HLML is the primary defense against the "BoryptGrab" style persistent access vulnerabilities.
* **Observability:** Active leases and revocation events are visualized in the "Mission Lease Manager."

## 7. Evolutionary Changelog
* **2026-07-24:** Initial Document Creation.

### Update: 2026-07-25 - Epistemic Lease Throttling
**Context:** Today's research into "Epistemic Watermarking" reveals that low-confidence reasoning increases the risk of "Lease Abuse."
**Architecture Adjustment:**
* Integrating with the **Epistemic Confidence Broker (ECB)** in Section 4.
* HLML will now dynamically throttle the *scope* or *duration* of a lease if the requesting agent's reasoning confidence falls below mission thresholds.
**Security Impact:** Adds a "Cognitive Safety" layer to privilege escalation, ensuring only confident agents receive high-stakes capabilities.
