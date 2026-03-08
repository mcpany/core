# Design Doc: Sandboxed Shell Execution

**Status:** Draft
**Created:** 2026-03-08

## 1. Context and Scope
Agentic AI tools frequently require shell access to perform system-level tasks (e.g., file manipulation, package installation, running tests). However, recent vulnerabilities like CVE-2026-2256 (MS-Agent) show that regex-based command sanitization is insufficient. MCP Any must provide a secure, isolated environment for shell-based tool execution to prevent host system compromise.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Execute all "Shell-type" tools within ephemeral OCI containers by default.
    *   Provide filesystem isolation with configurable mount points (Least Privilege).
    *   Implement resource limits (CPU, Memory, Network) for tool execution.
    *   Standardize the interface for shell tools to interact with the sandbox.
*   **Non-Goals:**
    *   Building a custom container runtime (will use Docker, Podman, or gVisor).
    *   Providing full persistent VM isolation (focus is on ephemeral task-based containers).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Developer using a local agent to refactor a codebase.
*   **Primary Goal:** Allow the agent to run `npm test` without risk of a malicious prompt executing `rm -rf /`.
*   **The Happy Path (Tasks):**
    1.  The agent requests a shell-based tool call.
    2.  MCP Any intercepts the call and identifies it as a "Sandboxed" tool.
    3.  A new ephemeral container is spun up with the project directory mounted as read-only (or read-write if specified).
    4.  The command is executed within the container.
    5.  Results are returned to the agent, and the container is immediately destroyed.

## 4. Design & Architecture
*   **System Flow:**
    - **Tool Classification**: Tools are tagged with `execution_mode: "sandboxed"`.
    - **Sandbox Provider**: A pluggable interface (e.g., `DockerProvider`, `PodmanProvider`) manages container lifecycles.
    - **Ephemeral Workspace**: A temporary volume is created for the tool's execution, including necessary environment variables and mounted project files.
*   **APIs / Interfaces:**
    - `mcp_sandbox.yaml`: Configuration for sandbox images and resource limits.
    - `ExecuteSandboxed(command string, options SandboxOptions) (Result, error)`
*   **Data Storage/State:** No persistent state within the sandbox; results are streamed back to the MCP Any host.

## 5. Alternatives Considered
*   **Regex Sanitization**: (MS-Agent approach) *Rejected* as it is prone to bypasses.
*   **Chroot/Jail**: *Rejected* for lack of robust resource isolation and dependency management.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The sandbox is the final line of defense against prompt injection leading to RCE.
*   **Observability:** Logs from the sandbox are captured and tagged with the `trace_id` for debugging.
*   **Performance**: Overhead of container startup (target < 500ms using pre-warmed pools).

## 7. Evolutionary Changelog
*   **2026-03-08:** Initial Document Creation.
