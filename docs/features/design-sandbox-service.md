# Design Doc: Sandbox-as-a-Service for Config Hooks
**Status:** Draft
**Created:** 2026-03-10

## 1. Context and Scope
With the discovery of critical vulnerabilities like CVE-2025-59536, it is clear that AI agents' reliance on project-local configuration files (e.g., `.claude/settings.json`) introduces severe Remote Code Execution (RCE) risks. These files often contain "hooks" or "auto-execute" commands that run with the user's privileges. MCP Any needs a native, isolated execution environmenta "Sandbox-as-a-Service"where approved hooks can run without compromising the host system.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a lightweight, containerized execution environment for project-local hooks.
    * Enforce strict resource limits (CPU, Memory, Disk) on hook execution.
    * Restrict network access to only explicitly authorized endpoints.
    * Ensure zero host filesystem access unless a specific directory is mounted with read-only permissions.
* **Non-Goals:**
    * Providing a general-purpose sandbox for all agent activities (focused specifically on config-driven hooks).
    * Supporting long-running background services within the sandbox.

## 3. Critical User Journey (CUJ)
* **User Persona:** Developer using an agent in a new, potentially untrusted repository.
* **Primary Goal:** Safely execute a project-init hook (e.g., `npm install` in a constrained way) without risk of machine takeover.
* **The Happy Path (Tasks):**
    1. Agent detects a hook in `.mcpany/hooks.json`.
    2. MCP Any validates the hook's signature and the user's previous attestation.
    3. MCP Any spawns a transient "Sandbox-as-a-Service" instance.
    4. The hook executes within the isolated container.
    5. Results are streamed back to the agent, and the sandbox is immediately destroyed.

## 4. Design & Architecture
* **System Flow:**
    `Agent` -> `MCP Any Hook Manager` -> `Sandbox Orchestrator (Docker/Firecracker)` -> `Isolated Hook Execution`
    1. **Trigger**: Hook Manager identifies an approved hook to run.
    2. **Isolation**: Sandbox Orchestrator provisions a new, ephemeral container with a "No-Network, No-Disk" default policy.
    3. **Execution**: The command is executed. Output (stdout/stderr) is captured.
    4. **Teardown**: The container is forcefully removed after a short timeout or completion.
* **APIs / Interfaces:**
    * `POST /v1/sandbox/execute`: Internal API to request a sandboxed execution.
    * `Stream /v1/sandbox/logs`: Real-time log stream from the running sandbox.
* **Data Storage/State:**
    * Stateless execution. Any necessary state must be passed in via environment variables or a specific mounted "Blackboard" volume.

## 5. Alternatives Considered
* **User-level `sudo` restricted accounts**: Rejected due to complexity of setup across different OSs and risk of privilege escalation.
* **Wasm-based execution**: Considered but rejected for now due to the need to run standard shell commands and scripts (e.g., `bash`, `npm`) which are harder to port to Wasm than a container.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: The sandbox is the ultimate fallback. Even if a malicious hook is attested by mistake, the sandbox prevents it from exfiltrating API keys or installing persistent malware on the host.
* **Observability**: Every sandbox execution is logged, including its resource usage and any attempted policy violations (e.g., blocked network calls).

## 7. Evolutionary Changelog
* **2026-03-10:** Initial Document Creation.
