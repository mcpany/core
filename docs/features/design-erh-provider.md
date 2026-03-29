# Design Doc: Ephemeral Registry Hook (ERH) Provider
**Status:** Draft
**Created:** 2026-07-12

## 1. Context and Scope
Current AI agent discovery models often rely on persistent tool schemas or discovery hooks. This persistence creates a "Registry Persistence" exploit vector where a malicious subagent can register a "Shadow Tool" that overlaps with a high-trust system tool. If this shadow tool persists across sessions, it can hijack subsequent missions.

The ERH Provider introduces session-locked discovery schemas. Instead of persisting tool definitions, discovery hooks are issued as ephemeral tokens that expire immediately post-discovery, ensuring that the "Discovery Bus" remains clean and sovereign for every new mission.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement "Time-to-Live" (TTL) for all discovered tool schemas and configuration hooks.
    * Mandate session-locked signatures for all `discoveryCommand` outputs.
    * Ensure that "Shadowing" a legitimate tool results in a hardware-attested conflict alert.
    * Automate the purging of discovery metadata immediately after the agent reasoning loop initialization.
* **Non-Goals:**
    * Blocking the installation of new tools (handled by the Skill Registry).
    * Sandboxing the execution of the hooks (handled by the Discovery Sandbox).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Ensure that tools discovered in an untrusted repository do not persist and affect other projects.
* **The Happy Path (Tasks):**
    1. The agent enters a new workspace and triggers a `discoveryCommand`.
    2. The ERH Provider wraps the discovery process in a session-bound "Ephemeral Window."
    3. Discovered tools are loaded into the agent's context with a 10-minute expiration token.
    4. The agent completes its task.
    5. MCP Any automatically purges the discovery hooks and tool schemas from the active registry.
    6. A subsequent agent session in a different workspace starts with a clean registry.

## 4. Design & Architecture
* **System Flow:**
    `Discovery Trigger` -> `ERH Provider (Token Issuance)` -> `Tool Registry (Short-lived entry)`
    1. **Issuance**: The ERH Provider generates a cryptographic `Session-Nonce` for the discovery event.
    2. **Locking**: All tool schemas produced during this event are signed with the `Session-Nonce`.
    3. **Expiration**: The Tool Registry monitors the `Session-Nonce` and removes all associated schemas when the mission context is closed.
* **APIs / Interfaces:**
    * `POST /v1/erh/provision`: Start an ephemeral discovery session.
    * `POST /v1/erh/commit`: Finalize discovery and lock schemas to the session.
* **Data Storage/State:** In-memory tool registry with session-bound expiration logic.

## 5. Alternatives Considered
* **Workspace Isolation (Docker-per-project)**: Rejected due to the extreme resource overhead for simple tool discovery.
* **Manual Purge Commands**: Rejected because agents often fail to terminate cleanly, leaving "Ghost" tools in the registry.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** "Registry Cleansing" is the primary defense. Even if a malicious tool is discovered, it cannot survive to compromise the next user mission.
* **Observability:** The "Ephemeral Hook Monitor" UI provides real-time visibility into active discovery tokens and their expiration status.

## 7. Evolutionary Changelog
* **2026-07-12:** Initial Document Creation.
