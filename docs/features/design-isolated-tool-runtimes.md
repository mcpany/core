# Design Doc: Isolated Tool Runtimes (Sandboxed Execution)

**Status:** Draft
**Created:** 2026-03-06

## 1. Context and Scope
The "Claude Code Security" report and recent kernel escape vulnerabilities highlight the danger of allowing AI agents to execute arbitrary code or shell commands directly on a host machine. While MCP Any provides the connectivity, it must also provide the **Safety**. Isolated Tool Runtimes allow MCP Any to spin up ephemeral, hardened sandboxes for executing high-risk tools, ensuring that even if a tool is compromised or an agent is "jailbroken," the host system remains protected.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Provide a standard interface for "Pluggable Runtimes" (Docker, gVisor, Wasm, firecracker).
    *   Automatically route "high-risk" tool calls to isolated environments.
    *   Manage the lifecycle (creation, execution, cleanup) of ephemeral sandboxes.
    *   Support secure file-system mounting and environment variable injection into sandboxes.
*   **Non-Goals:**
    *   Building a new container orchestrator (leverage existing ones like Docker/Podman).
    *   Replacing host-level security (it's a layer of defense-in-depth).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Developer using an autonomous coding agent.
*   **Primary Goal:** Allow the agent to run tests and install dependencies without risking the local machine.
*   **The Happy Path (Tasks):**
    1.  User configures the "Shell" tool to use the `docker-sandbox` runtime.
    2.  Agent calls `execute_shell(command="npm install && npm test")`.
    3.  MCP Any intercepts the call, creates a temporary Docker container based on a "Safe Image."
    4.  The command runs inside the container.
    5.  Results are returned to the agent, and the container is immediately destroyed.

## 4. Design & Architecture
*   **System Flow:**
    - **Risk Assessment**: MCP Any identifies tools marked with `runtime_isolation: true`.
    - **Runtime Selection**: The `SandboxManager` selects the appropriate runtime based on configuration.
    - **Execution**: The tool logic is executed via the runtime's driver (e.g., `docker exec`).
    - **Cleanup**: Ephemeral resources are purged after the call returns.
*   **APIs / Interfaces:**
    - Runtime Driver Interface: `Run(image, cmd, env, mounts) (output, error)`
    - Configuration: `runtimes: { "safe-python": { "driver": "wasm", "image": "python:3.11-slim" } }`
*   **Data Storage/State:** Ephemeral state is stored in temporary volumes or memory-backed filesystems.

## 5. Alternatives Considered
*   **User-Provided Sandboxes**: Forcing users to manually set up containers. *Rejected* as it adds too much friction.
*   **Virtual Machines**: Too slow for per-tool execution. *Rejected* in favor of container/Wasm isolation for performance.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Limits the "Blast Radius" of a compromised agent or tool.
*   **Observability:** Logs should clearly indicate which calls were sandboxed and provide resource usage metrics (CPU/Mem) for the sandbox.

## 7. Evolutionary Changelog
*   **2026-03-06:** Initial Document Creation.
