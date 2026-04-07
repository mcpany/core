# Design Doc: Hardware-Locked Mission Lease (HLML) Provider
**Status:** Approved
**Created:** 2026-07-25

## 1. Context and Scope
As agents transition from short-lived sessions to long-running autonomous swarms, static tool permissions become a liability. Permanent access to high-privilege tools (e.g., `run_shell_command`) increases the blast radius of a compromised subagent. The industry is moving toward "Mission-Bound" agency, where privileges are tied to specific, pre-declared objectives.

The HLML Provider implements a JIT (Just-in-Time) capability model where tool access is issued as TPM-signed leases that are cryptographically bound to a specific mission fragment and automatically expire upon task completion.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue hardware-attested (TPM/Secure Enclave) tool capability leases.
    * Bind leases to specific mission-root task IDs and pre-declared manifests.
    * Enforce automated, hardware-level revocation upon mission completion or timeout.
    * Neutralize persistent privilege escalation in specialist agents.
* **Non-Goals:**
    * Managing human user login sessions (handled by the OIDC/SSO layer).
    * Providing a general-purpose secret manager.

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Architect
* **Primary Goal:** Ensure a "Deployment Agent" only has `ssh` access during the 10-minute window of a production rollout, and that access is physically revoked immediately after.
* **The Happy Path (Tasks):**
    1. The Mission-Root signs a "Rollout Manifest" containing the `ssh` tool.
    2. The HLML Provider issues a 15-minute, TPM-signed lease for the Deployment Agent.
    3. The agent executes the deployment tools; the gateway validates the lease signature and task ID for every call.
    4. Upon task completion, the agent issues a `MISSION_COMPLETE` signal.
    5. The HLML Provider invalidates the lease in the hardware root, and subsequent tool calls are blocked even if the token is leaked.

## 4. Design & Architecture
* **System Flow:**
    * [Mission Root] --(Attested Manifest)--> [HLML Provider]
    * [HLML Provider] --(TPM Signing)--> [Mission Lease Token]
    * [Agent] --(Token + Tool Call)--> [Security Gateway]
    * [Security Gateway] --(Lease Validation)--> [Tool Execution]
* **APIs / Interfaces:**
    * `POST /v1/leases/mint`: Issue a new hardware-bound lease.
    * `POST /v1/leases/revoke`: Forceful revocation of an active mission lease.
* **Data Storage/State:**
    * Active lease metadata is stored in a secure, kernel-bound memory region accessible only to the HLML Provider and validated by the hardware root.

## 5. Alternatives Considered
* **Time-Bound JWTs**: Rejected because they cannot be forcefully revoked at the hardware layer and are vulnerable to clock-skew attacks.
* **Standard RBAC**: Rejected due to lack of task-specific granularity and the risk of "Permission Creep."

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The lease signature must include the hardware Inode of the mission manifest to prevent TOCTOU manifest swapping.
* **Observability:** Audit logs will record lease duration vs. actual usage to optimize future lease windows via the Adaptive Mission Lease (AML) service.

## 7. Evolutionary Changelog
* **2026-07-24:** Feature proposed in strategic sync.
* **2026-07-25:** Initial Document Creation.
