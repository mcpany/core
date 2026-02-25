# Design Doc: Immutable Agentic Config Store
**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
As agents gain more autonomy, they are increasingly targeted with "Config-Injection" attacks, where an attacker tricks an LLM into modifying its own configuration files or security policies. Recent investigations into OpenClaw (MITRE ATLAS) have shown that if an agent has write access to its own orchestration parameters, it can be coerced into disabling its safety filters or granting itself unauthorized tool access.

MCP Any needs to solve this by decoupling the **Agentic Execution Layer** from the **Configuration Management Layer**. By moving security policies and tool definitions into an immutable "Infrastructure Plane," we ensure that even a compromised agent cannot lower its own security guards.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a read-only configuration interface for agents.
    * Ensure that tool definitions (schemas, endpoints, auth) cannot be modified by the agent at runtime.
    * Implement "Configuration Pinning" where a session is bound to a specific hash of the configuration.
    * Allow hot-reloading by administrators without granting agents write access.
* **Non-Goals:**
    * Replacing existing secret managers (e.g., HashiCorp Vault).
    * Providing a full-blown agent orchestration framework (like AutoGen).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Agent Developer
* **Primary Goal:** Ensure that an autonomous coding agent cannot accidentally or maliciously delete the "Policy Firewall" or add unauthorized tools to its repertoire.
* **The Happy Path (Tasks):**
    1. The developer starts MCP Any with a signed `config.yaml`.
    2. The agent attempts to call a `modify_config` tool to add a new malicious MCP server.
    3. MCP Any intercepts the call and rejects it because the configuration plane is in `Immutable Mode`.
    4. The agent is forced to operate within the predefined, secure boundaries.

## 4. Design & Architecture
* **System Flow:**
    ```
    [Agent Swarm] --(Tool Call)--> [MCP Any Gateway]
                                          |
                                          |--[Check Policy (Read-Only)]
                                          |--[Fetch Tool Schema (Read-Only)]
                                          |
    [Admin Console] --(Update)--> [Immutable Config Store (Write-Only for Admin)]
    ```
* **APIs / Interfaces:**
    * `GET /v1/config/policy`: Returns current active policy (ReadOnly for agents).
    * `POST /v1/admin/config/reload`: Admin-only endpoint to update the immutable state via a signed manifest.
* **Data Storage/State:**
    * State is stored in a memory-mapped file that is mounted as read-only to the agent execution environment.
    * SHA256 checksums are used to verify integrity on every tool execution request.

## 5. Alternatives Considered
* **Agent-Managed Policies**: Rejected because it creates a circular trust problem. If the agent manages its own safety, it can be tricked into disabling it.
* **Database-Backed Config**: Considered, but adds complexity and latency. File-based immutable stores are easier to audit and faster for high-frequency tool calls.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The config store is the "Root of Trust." It must be protected by OS-level permissions (e.g., root-owned files in a container).
* **Observability:** Every attempt by an agent to modify the immutable store must be logged as a "High Severity" security event.

## 7. Evolutionary Changelog
* **2026-02-27:** Initial Document Creation.
