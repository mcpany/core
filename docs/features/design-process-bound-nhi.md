# Design Doc: Process-Bound NHI Provider
**Status:** Draft
**Created:** 2026-04-15

## 1. Context and Scope
With the shift towards headless agents (e.g., Claude Code Remote Control) and autonomous CI/CD pipelines, agent identities can no longer rely on transient terminal sessions or human-in-the-loop (HITL) presence. If a terminal session is disconnected, the agent often loses its attestation context, leading to "Sovereignty Decay."

The Process-Bound NHI Provider establishes a cryptographically secure identity that is bound to the execution process itself and hardware-anchored (TPM/Secure Enclave). This ensures that a headless agent maintains its permissions and mission-bound sovereignty even when running as a persistent background service.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue hardware-attested identity tokens bound to the PID and execution environment of the agent.
    * Maintain sovereignty across terminal session disconnects and process handoffs.
    * Implement "Process-Alive" heartbeats to automatically revoke tokens if the parent process is hijacked or terminated.
    * Support seamless migration of process identity across parallel Git worktrees.
* **Non-Goals:**
    * Replacing standard A2A identity tokens (it complements them).
    * Providing cross-machine process migration (limited to the local host).

## 3. Critical User Journey (CUJ)
* **User Persona:** DevOps Engineer / Headless Swarm Orchestrator
* **Primary Goal:** Run an autonomous codebase audit across 5 parallel branches without losing permission context when the SSH session closes.
* **The Happy Path (Tasks):**
    1. The orchestrator starts a headless MCP Any session bound to a specific mission.
    2. The Process-Bound NHI Provider generates a TPM-signed token linked to the orchestrator's PID.
    3. The agent spawns 5 subagents for different branches; each subagent inherits a "Process-Bound Fragment" of the parent token.
    4. The user closes the terminal. The background process continues.
    5. Subagents perform tool calls; the gateway validates each call against the process-resident identity.
    6. Upon mission completion, the main process terminates, triggering a hardware-locked revocation of all fragments.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Process Start] --> B[TPM-Bound ID Generation]
        B --> C[PID Binding]
        C --> D[Headless Session Active]
        D --> E{Process Alive?}
        E -- Yes --> F[Periodic Heartbeat]
        F --> D
        E -- No --> G[Immediate Token Revocation]
        G --> H[Audit Log: Process Death]
    ```
* **APIs / Interfaces:**
    * `BindProcessIdentity(pid, missionRoot) -> ProcessToken`
    * `ValidateProcessCall(token, callerPid) error`
    * `HandoffIdentity(newPid) -> error`
* **Data Storage/State:**
    * Tokens are stored in kernel-bound, origin-locked memory regions, inaccessible to other user-space processes.

## 5. Alternatives Considered
* **Persistent Disk-Based Tokens**: Rejected due to high risk of exfiltration. Disk tokens are not bound to a running execution context.
* **Long-Lived JWTs**: Rejected as they don't solve the session-disconnect problem and lack hardware-level process binding.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Integrates with the `HLES` (Hardware-Locked Environment Sovereignty) provider to ensure tokens are invisible to subagent environments.
* **Observability:** Process heartbeats and identity lineage are visualized in the "Mesh Identity Manager" UI.

## 7. Evolutionary Changelog
* **2026-04-15:** Initial Document Creation.
