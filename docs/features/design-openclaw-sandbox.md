# Design Doc: OpenClaw Security Sandbox

**Status:** Draft
**Created:** 2026-02-22

## 1. Context and Scope
The viral growth of OpenClaw and similar local-first agents has highlighted a massive security vulnerability: these agents often run with full user permissions on the host machine, executing shell commands and accessing files without any isolation. MCP Any, as a universal gateway, is uniquely positioned to solve this by providing a hardened execution environment for these agents. This feature aims to provide a "Secure Sandbox" adapter that transparently wraps local CLI/FS tools in an isolated environment.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide strict process-level isolation for local command execution.
    * Enforce granular filesystem access (chroot, bind-mounts, or virtual volumes).
    * Prevent unauthorized network access from sandboxed agents.
    * Maintain compatibility with existing MCP Command and Filesystem adapters.
* **Non-Goals:**
    * Replacing the host OS security model.
    * Providing a full GUI virtualization solution.
    * Support for proprietary, non-Linux execution environments (e.g., native Windows binaries without WSL2/Docker).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local Developer / OpenClaw Power User
* **Primary Goal:** Execute potentially untrusted local skills (e.g., browser automation, data scripts) without risking host machine integrity.
* **The Happy Path (Tasks):**
    1. User configures a new `sandbox` upstream in MCP Any.
    2. User defines allowed directories and specific command whitelists.
    3. The AI Agent (OpenClaw via MCP Any) requests a tool execution (e.g., `ls /project`).
    4. MCP Any spawns a transient, isolated container (e.g., via Docker or gVisor) to run the command.
    5. The result is returned to the agent, while the environment is instantly destroyed.

## 4. Design & Architecture
* **System Flow:**
    [Agent] -> [MCP Any Core] -> [Sandbox Adapter] -> [Transient Container/MicroVM] -> [Command Exec]
* **APIs / Interfaces:**
    * New Upstream Type: `sandbox_command`
    * Configuration Schema:
        ```yaml
        type: sandbox_command
        image: "mcpany/agent-runtime:latest"
        allowed_paths: ["/data", "/tmp"]
        network_access: none
        ```
* **Data Storage/State:**
    * Ephemeral storage by default.
    * Persistent volumes can be mounted via `bind-mount` with `read-only` flags.

## 5. Alternatives Considered
* **Direct Shell Execution (Current):** Rejected due to lack of security (the "OpenClaw Problem").
* **Virtual Machines (Firecracker):** Considered for maximum isolation, but rejected for V1 due to startup latency and complexity compared to Docker-based sandboxing.
* **Wasm (WebAssembly):** Excellent for isolation but limited for general CLI tools that expect a POSIX environment.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All execution is denied by default. Commands must be whitelisted. Filesystem access is strictly limited to explicitly defined paths.
* **Observability:** Every sandboxed command execution is logged with its full environment, input, output, and resource usage (CPU/Memory).

## 7. Evolutionary Changelog
* **2026-02-22:** Initial Document Creation.
