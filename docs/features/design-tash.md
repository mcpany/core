# Design Doc: Teammate-Aware Safety Hooks (TASH)
**Status:** Draft
**Created:** 2026-11-02

## 1. Context and Scope
As agent swarms transition from hierarchical sessions to horizontal "teammate" meshes (e.g., Claude Code Agent Teams), the "one-size-fits-all" security hook model is failing. Currently, all teammates in a swarm inherit the same safety hook configuration from the lead agent. This prevents security architects from applying differentiated policies—such as allowing an "Implementer" agent to run shell commands while restricting a "Researcher" agent to read-only tools.

Furthermore, current hook payloads lack teammate identity, making it impossible for the MCP Any gateway to log or interdict actions based on which specific subagent within the mesh triggered them. TASH solves this by mandating hardware-attested teammate identity in every safety hook request.

## 2. Goals & Non-Goals
* **Goals:**
    * Mandate the inclusion of `teammate_id` and `teammate_role` in all PreToolUse safety hook payloads.
    * Provide a mechanism for the gateway to apply different Rego/CEL policies based on the attested role of the teammate.
    * Enable cryptographic lineage verification to ensure a teammate cannot spoof a higher-privilege role.
* **Non-Goals:**
    * Managing the internal logic of the subagents themselves.
    * Providing a UI for real-time teammate role reassignment (this is handled by the mission manifest).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Architect for Enterprise AI Swarms
* **Primary Goal:** Apply a "Read-Only" policy to researcher teammates while allowing "Full-Write" access to developer teammates within the same mission.
* **The Happy Path (Tasks):**
    1. The lead agent spawns two teammates: `ResearchAgent` (Role: Researcher) and `DevAgent` (Role: Developer).
    2. MCP Any issues hardware-attested identity tokens to both teammates, bound to their respective roles.
    3. `ResearchAgent` attempts to call `run_shell_command` to install a package.
    4. The TASH middleware intercepts the call, extracts the `ResearchAgent` identity, and identifies the "Researcher" role.
    5. The policy engine matches the "Researcher" role against a "Deny:Write" policy and blocks the call.
    6. `DevAgent` calls the same tool; TASH identifies the "Developer" role and allows the call after standard audit.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Teammate Agent->>TASH Middleware: Tool Call (with Attested Token)
        TASH Middleware->>Identity Provider: Verify Teammate ID & Role
        Identity Provider-->>TASH Middleware: Validated (Role: Researcher)
        TASH Middleware->>Policy Engine: Check Policy(Researcher, Tool)
        Policy Engine-->>TASH Middleware: Access Denied
        TASH Middleware-->>Teammate Agent: Error: Role Restricted
    ```
* **APIs / Interfaces:**
    * Update `PreToolUse` payload:
      ```json
      {
        "session_id": "mission-root-123",
        "teammate_id": "agent-xyz",
        "teammate_role": "researcher",
        "attestation": "tpm-signature-blob",
        "tool_name": "...",
        "tool_input": "..."
      }
      ```
* **Data Storage/State:** Teammate roles are anchored to the Hardware-Attested Mission Manifest (HAMM).

## 5. Alternatives Considered
* **Lineage-Only Tracking:** Rejected because lineage alone (Parent -> Child) doesn't capture the *intent* or *role* of the child, only its origin. A "Researcher" and "Developer" could have the same parent but require vastly different permissions.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All `teammate_id` and `role` claims must be hardware-attested to prevent role-escalation within the swarm.
* **Observability:** Audit logs will now include `teammate_id`, allowing for granular forensic analysis of swarm behavior.

## 7. Evolutionary Changelog
* **2026-11-02:** Initial Document Creation.
