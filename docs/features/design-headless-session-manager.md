# Design Doc: Headless Session Manager (HSM)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
The emergence of Claude Code's "Remote Control" and "Dispatch" indicates a market shift from local, terminal-bound agents to distributed, headless agentic workers. Currently, MCP Any provides local tool execution but lacks a unified control plane for monitoring and steering agents running in headless or remote contexts. HSM is designed to bridge this gap, allowing developers to manage remote agent sessions as first-class infrastructure components.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a unified dashboard for monitoring active headless agent sessions across the mesh.
    * Implement "Steer" capability to inject corrective intents or manual tool approvals into remote sessions.
    * Enforce hardware-attested lineage for all remote control instructions.
    * Support session pausing and state snapshots for long-running background tasks.
* **Non-Goals:**
    * Providing a general-purpose remote terminal (HSM is intent-oriented, not character-oriented).
    * Bypassing local sandboxing on remote nodes.

## 3. Critical User Journey (CUJ)
* **User Persona:** Remote DevOps Orchestrator
* **Primary Goal:** Steer a headless coding agent running in a CI pipeline from a local mobile terminal.
* **The Happy Path (Tasks):**
    1. A headless agent session is initialized on a remote build server via MCP Any HSM.
    2. The session hits a high-risk tool call (e.g., `delete_database`) and pauses for attestation.
    3. The orchestrator receives a notification on their local terminal.
    4. The orchestrator reviews the agent's internal reasoning via the SRM Provider.
    5. The orchestrator issues a "Proceed" command, cryptographically signed by their local TPM.
    6. HSM on the remote node verifies the signature and mission-root lineage.
    7. The agent resumes execution and completes the task.

## 4. Design & Architecture
* **System Flow:**
    `[Remote Terminal] -> (Signed Intent) -> [AMT Broker] -> [Headless Session Manager] -> [Agent Runtime]`
* **APIs / Interfaces:**
    * `hsm.v1.ListSessions()`: Returns active headless worker IDs and statuses.
    * `hsm.v1.SteerSession(session_id, instruction)`: Injects a signed intent into the agent's monologue.
    * `hsm.v1.SnapshotSession(session_id)`: Triggers a PLSS snapshot of the worker environment.
* **Data Storage/State:**
    * **Session Registry**: Persistent store of active session metadata, bound to mission-root IDs.
    * **Steer Log**: Hardware-attested audit trail of all manual interventions.

## 5. Alternatives Considered
* **SSH-based Proxying**: Rejected as it lacks agentic awareness (cannot inspect reasoning traces or enforce mission-bound constraints).
* **Web-based Agent GUIs**: Often lack hardware attestation and are vulnerable to browser-origin hijacking (CVE-2026-25253). HSM utilizes the AMT Broker for secure P2P steering.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All steering commands require "Recursive Mission-Root Attestation." A controller cannot steer an agent unless they prove they own the mission root.
* **Observability:** Real-time session heartbeat and reasoning-intensity monitoring via the "Headless Trust Status Widget."

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
