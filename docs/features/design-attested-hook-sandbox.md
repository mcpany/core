# Design Doc: Attested Hook Sandbox
**Status:** Draft
**Created:** 2026-03-17

## 1. Context and Scope
The rise of "Configuration-as-Execution" vulnerabilities (e.g., in Claude Code and OpenClaw) has turned project-local settings files into a dangerous RCE vector. Malicious actors can inject shell commands as "hooks" or "auto-execute" tasks into `.claude/settings.json` or `.mcp.json`. When an agent loads these projects, these commands execute with the full privileges of the user. MCP Any must intercept these hooks and execute them in a "Detached Sandbox" that is resource-isolated and network-restricted by default.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept all command execution requests originating from project-local configurations.
    * Execute hooks in an isolated OCI-compliant container or WebAssembly (Wasm) runtime.
    * Enforce strict resource limits (CPU, Memory, Disk IO) on hook execution.
    * Require explicit user attestation (MFA) via the HITL Middleware for any host-level filesystem access.
    * Block all network access by default for hooks.
* **Non-Goals:**
    * Providing a general-purpose sandbox for *all* agent activities (only for config-triggered hooks).
    * Fixing the underlying security flaws in third-party agent frameworks.

## 3. Critical User Journey (CUJ)
* **User Persona:** Developer collaborating on an open-source project.
* **Primary Goal:** Safely load a project containing automation hooks without risking host compromise.
* **The Happy Path (Tasks):**
    1. User `git pulls` a repository that includes a `.mcp.json` with an `on_load` hook.
    2. Agent attempts to initialize the project and calls the hook.
    3. MCP Any intercepts the call and identifies it as a "Project Hook."
    4. MCP Any pauses execution and prompts the user: "Allow project hook `npm install` to run in a sandbox?"
    5. User approves.
    6. MCP Any spawns a detached sandbox, executes the command, and returns the output to the agent.

## 4. Design & Architecture
* **System Flow:**
    - **Hook Interceptor**: A middleware that monitors configuration loading and identifies executable blocks.
    - **Sandbox Manager**: Orchestrates the lifecycle of the isolated runtime (e.g., using `runc`, `nsjail`, or `wasmtime`).
    - **Policy Engine**: Validates the hook's requested capabilities against the user's global and project-specific security policies.
* **APIs / Interfaces:**
    - `sandbox.execute(command, capabilities, timeout)`
    - Internal Protocol: `mcpany-sandbox://v1`
* **Data Storage/State:**
    - Persistent "Attestation Registry" of approved hook hashes to avoid redundant prompts.

## 5. Alternatives Considered
- **Virtual Machines (VMs)**: Too heavy and slow for short-lived automation hooks.
- **Static Analysis**: Hard to predict the full behavior of shell scripts; sandboxing is a more robust "Defense-in-Depth" strategy.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The sandbox is the primary enforcement point for Zero Trust hook execution.
* **Observability:** Hook logs and resource usage are streamed to the "Config Sandbox Monitor" in the UI.

## 7. Evolutionary Changelog
* **2026-03-17:** Initial Document Creation.
