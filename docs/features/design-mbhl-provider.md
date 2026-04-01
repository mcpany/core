# Design Doc: Mission-Bound Hardware Leases (MBHL) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agent swarms move toward high-autonomy environments (e.g., Claude Code Agent Teams), the risk of "Persistent Privilege Escalation" has become a primary security concern. If a specialist subagent is granted a powerful capability (like `run_shell_command` or `db:admin`) and is subsequently compromised or enters a recursive hallucination loop, it can cause catastrophic damage beyond its intended task scope.

The MBHL Provider is required to implement the Claude Code v3.2.0 standard, where capabilities are issued as hardware-locked, TPM-signed leases that are strictly bound to the lifecycle of a specific mission-root task.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue hardware-attested (TPM-signed) capability leases for all high-privilege subagent operations.
    * Enforce strict temporal and task-bound boundaries for every lease.
    * Facilitate automated "Lease Revocation" immediately upon sub-mission termination or intent drift.
    * Neutralize persistent privilege escalation by ensuring no capability survives its mission context.
* **Non-Goals:**
    * Managing the user's primary identity (this manages tool-specific leases).
    * Providing a general-purpose secret manager.
    * Hardening the underlying MCP tool execution (that is the job of the sandbox).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Swarm Orchestrator
* **Primary Goal:** Grant a "DevOps Specialist" agent temporary SSH access to a specific staging server only for the duration of a deployment task.
* **The Happy Path (Tasks):**
    1. Parent agent initiates a "Deployment" mission and requests a lease for the `ssh_exec` tool.
    2. MBHL Provider verifies the request against the mission-root manifest and issues a TPM-signed lease token (TTL: 10m).
    3. The DevOps specialist receives the lease and executes the tool calls through the MCP Any gateway.
    4. Upon each call, MCP Any validates the lease signature and its binding to the active task ID.
    5. The deployment task completes; the parent agent signals mission termination to the MBHL Provider.
    6. MBHL Provider immediately revokes the lease across the mesh; subsequent attempts by the specialist to use `ssh_exec` are interdicted.

## 4. Design & Architecture
* **System Flow:**
    [Parent Agent] --(Request Lease)--> [MBHL Provider] --(TPM Sign)--> [Lease Token]
    [Subagent] --(Lease Token + Tool Call)--> [MCP Any Gateway] --(Verify Lease)--> [MCP Tool]
    [Mission Monitor] --(End Signal)--> [MBHL Provider] --(Revoke)--> [Mesh Revocation List]
* **APIs / Interfaces:**
    * `mbhl.IssueLease(capabilities[], missionID, taskID, ttl) -> LeaseToken`
    * `mbhl.VerifyLease(leaseToken, currentTaskID) -> Valid/Invalid`
    * `mbhl.RevokeLease(leaseToken/missionID) -> Success`
* **Data Storage/State:**
    * **Lease Registry:** Secure, kernel-bound memory storage for active lease metadata.
    * **Revocation List:** High-frequency, mesh-synchronized list of blacklisted lease hashes.

## 5. Alternatives Considered
* **Time-Bound JWTs:** Rejected because standard JWTs are not hardware-locked and can be replayed if exfiltrated from the subagent's process memory.
* **Static Capability Tokens:** Rejected because they lack the granular task-binding required for horizontal Agent Teams.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All leases are cryptographically bound to the hardware enclave (TPM/SEP) of the initiating node.
* **Observability:** Integrated with the "Mission Lease Manager" UI for real-time visualization of active and expired leases.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
