# Design Doc: Sandboxed Hook Execution Runtime
**Status:** Draft
**Created:** 2026-03-09

## 1. Context and Scope
MCP Any allows users to configure "Hooks" that execute at various stages of the tool call lifecycle (e.g., `pre_call`, `post_call`). Recent vulnerabilities in similar tools (e.g., Claude Code CVE-2025-59536) have demonstrated that if these hooks execute in the user's default shell with full privileges, a malicious repository can achieve Remote Code Execution (RCE) simply by including a crafted configuration file.

MCP Any must solve this by ensuring all configured hooks run in a restricted, sandboxed environment that lacks access to sensitive host resources unless explicitly granted.

## 2. Goals & Non-Goals
* **Goals:**
    * Isolate hook execution from the host shell.
    * Provide a "Zero Trust" environment for scripts defined in configuration files.
    * Allow granular resource access (e.g., specific directories, environment variables) via explicit policy.
    * Support multiple sandbox backends (e.g., WebAssembly, gVisor/Docker, or OS-level restricted processes).
* **Non-Goals:**
    * Providing a full container orchestration system.
    * Supporting long-running background processes within hooks.
    * Replacing the main command adapter's execution logic (this is specifically for lifecycle hooks).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer
* **Primary Goal:** Run an agent on an untrusted open-source repository without risking host compromise via malicious hooks.
* **The Happy Path (Tasks):**
    1. The user starts `mcpany` in a new repository.
    2. The repository contains a `.mcpany/config.yaml` with a `pre_call` hook.
    3. MCP Any detects the hook and prepares the **Sandboxed Hook Runtime**.
    4. The runtime initializes an isolated environment (e.g., a Wasm runtime or a restricted subprocess).
    5. The hook executes. If it tries to access `/etc/passwd` or `~/.ssh`, the sandbox blocks the attempt.
    6. The hook returns its result to the MCP Any pipeline.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        Core[MCP Any Core] --> HookMgr[Hook Manager]
        HookMgr --> SandboxFactory[Sandbox Factory]
        SandboxFactory -->|Wasm| WasmRunner[Wasm Sandbox]
        SandboxFactory -->|Process| RestrictedProc[NSJail / Restricted Process]

        WasmRunner -->|Limited VFS| HookScript[Hook Script]
        RestrictedProc -->|Cgroups/Namespaces| HookScript
    ```
* **APIs / Interfaces:**
    * `HookRuntime` interface with `Execute(script string, ctx Context) (Result, error)` method.
    * `SandboxPolicy` struct defining allowed syscalls, file paths, and env vars.
* **Data Storage/State:**
    * Ephemeral state only. Any persistent state must be explicitly handled via the `Shared KV Store` tool.

## 5. Alternatives Considered
* **Native Shell Execution (Current):** Rejected due to RCE risks.
* **Docker-only Sandbox:** Rejected due to high overhead and dependency on a running Docker daemon.
* **WebAssembly (Wasm):** Considered as the primary candidate for lightweight, cross-platform isolation.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Hooks must have "Deny All" permissions by default. Access to the filesystem or network must be declared in the MCP Any master configuration.
* **Observability:** All hook execution attempts, including sandbox violations, must be logged to the audit trail.

## 7. Evolutionary Changelog
* **2026-03-09:** Initial Document Creation.
