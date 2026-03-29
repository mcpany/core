# Design Doc: Ephemeral Registry Hook (ERH) Provider
**Status:** Draft
**Created:** 2026-03-29

## 1. Context and Scope
The tool discovery phase has emerged as a primary attack vector for autonomous agents. Malicious subagents often attempt "Registry Persistence," where they inject long-lived configuration hooks (e.g., in `.claude/settings.json`) that shadow legitimate tools or exfiltrate state in subsequent sessions.

The Ephemeral Registry Hook (ERH) Provider transitions tool discovery from a static, filesystem-resident process to a session-locked, ephemeral event. It ensures that discovery schemas and capability cards expire immediately after the handshake phase, neutralizing the ability for "Shadow Tools" to persist across mission boundaries.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue time-bound, session-locked discovery tokens for all MCP tools.
    * Mandate the deletion of project-local configuration hooks post-discovery.
    * Provide hardware-attested "Absence Proofs" for stale registry entries.
    * Align with the Claude Code v3.2 ERH standard.
* **Non-Goals:**
    * Replacing the Protocol-Neutral Task Discovery (PNTD) hub (ERH is a security layer for PNTD).
    * Managing the execution of the tools themselves (focus is strictly on discovery/mapping).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer
* **Primary Goal:** Prevent a malicious repository from injecting a permanent tool hook that exfiltrates environment variables.
* **The Happy Path (Tasks):**
    1. The agent boots in a new project and requests tool discovery via MCP Any.
    2. The ERH Provider generates an ephemeral, hardware-locked discovery schema.
    3. The agent maps capabilities and performs the mission.
    4. Upon mission completion or session timeout, the ERH Provider invalidates the discovery token.
    5. A subsequent session attempting to use the "shadowed" tool fails because the ERH-backed schema no longer exists in the kernel-bound registry.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        Agent[Agent Discovery Req] --> PNTD[Discovery Hub]
        PNTD --> ERH[ERH Provider]
        ERH --> Token[Issue Ephemeral Token]
        Token --> Discovery[Active Mapping]
        Discovery --> Timeout[Session End]
        Timeout --> Invalidate[Evict from Registry]
    ```
* **APIs / Interfaces:**
    * `POST /registry/ephemeral/init`: Start an ephemeral discovery session.
    * `GET /registry/ephemeral/status`: Check the TTL of an active schema.
* **Data Storage/State:**
    * Ephemeral schemas are stored in non-persistent, kernel-bound memory (memfd-backed), never touching the physical disk.

## 5. Alternatives Considered
* **Read-only Filesystem Mounts:** Rejected because many agents require write access to specific project-local paths for legitimate reasons.
* **Manual Approval for Every Discovery:** Rejected due to "Discovery Fatigue" in complex swarms with 100+ tools.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** ERH relies on the Local-Only WebSocket Auth (LOWA) for session binding.
* **Observability:** Expired discovery attempts are logged in the Ephemeral Hook Monitor.

## 7. Evolutionary Changelog
* **2026-03-29:** Initial Document Creation.
