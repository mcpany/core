# Design Doc: Virtualized Configuration Sandbox (Config-Jail)
**Status:** Draft
**Created:** 2026-03-08

## 1. Context and Scope
With the rise of "Configuration-as-Code" in agentic tools like Claude Code and OpenClaw, project-level configuration files (`.mcp.json`, `.claude/settings.json`) have become a primary attack vector. Recent vulnerabilities (CVE-2025-59536, CVE-2026-21852) demonstrate that malicious hooks or overridden settings can lead to Remote Code Execution (RCE) on a developer's machine.

MCP Any needs to provide a "Safe-by-Default" execution environment for these project-level configurations. Config-Jail ensures that any command or hook defined in a non-global configuration is executed within an isolated, ephemeral sandbox.

## 2. Goals & Non-Goals
* **Goals:**
    * Isolate execution of hooks/commands from the host filesystem and network.
    * Provide a seamless "opt-in" for trusted repositories while defaulting to the sandbox.
    * Support multiple sandboxing backends (Docker, WebAssembly, gVisor).
* **Non-Goals:**
    * Providing a full-blown CI/CD environment.
    * Sandboxing the MCP Any core process itself (this is about the *upstream* tools it triggers).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer / Enterprise AI Architect.
* **Primary Goal:** Safely run an agent in a repo containing untrusted or third-party `.mcp.json` files without risking host RCE.
* **The Happy Path (Tasks):**
    1. Agent detects a project-level `.mcp.json` with a `pre-command` hook.
    2. MCP Any intercepts the hook execution request.
    3. MCP Any spawns an ephemeral Config-Jail container.
    4. The hook runs inside the container with restricted access (read-only project mount).
    5. Results are returned to the agent; the container is destroyed.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        Agent -->|Execute Hook| MCP_Any
        MCP_Any -->|Check Origin| Policy_Engine
        Policy_Engine -->|Untrusted| Sandbox_Manager
        Sandbox_Manager -->|Spawn| Config_Jail(Ephemeral Sandbox)
        Config_Jail -->|Command| Result
        Result --> MCP_Any
        MCP_Any -->|Result| Agent
    ```
* **APIs / Interfaces:**
    * `SandboxProvider`: Interface for different backends (`Run(cmd string, mount string) error`).
    * `ConfigIntercepter`: Middleware that identifies project-level config triggers.
* **Data Storage/State:**
    * Ephemeral; no state is persisted between hook executions unless explicitly configured via a persistent volume mount.

## 5. Alternatives Considered
* **Native Process Isolation (chroot/namespaces):** Rejected for being too complex to maintain cross-platform and less secure than containerization.
* **User Confirmation (MFA) only:** Rejected because users often "blindly click" through security prompts. Sandbox + MFA is better.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The sandbox should have no network access by default and limited CPU/Memory quotas.
* **Observability:** Logs from the sandbox are streamed back to MCP Any and tagged with the specific project origin.

## 7. Evolutionary Changelog
* **2026-03-08:** Initial Document Creation.
