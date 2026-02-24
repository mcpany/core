# Design Doc: Isolated Subprocess Runner

**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
As the MCP ecosystem grows, many agents rely on local command-line tools (e.g., shell scripts, Python utilities, CLI binaries) to perform their tasks. Current implementations often execute these tools directly on the host machine using standard subprocess calls. This poses a significant security risk: a malicious or hallucinating agent could craft a "command injection" payload or access unauthorized files. The Isolated Subprocess Runner aims to mitigate this by executing all local tools within a hardened, sandboxed environment.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Execute command-based MCP tools in a restricted sandbox (e.g., using `nsjail`, `gVisor`, or lightweight Docker containers).
    *   Restrict file system access to a per-tool "Workspace Directory."
    *   Limit network access for sandboxed processes (default to disabled).
    *   Enforce resource limits (CPU, Memory, Timeout) to prevent DOS attacks by rogue tools.
*   **Non-Goals:**
    *   Providing a full virtual machine for tool execution.
    *   Replacing cloud-based sandboxes (e.g., Anthropic's). This is specifically for local tool hardening.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Security Architect.
*   **Primary Goal:** Ensure that local tools called by an OpenClaw agent cannot access the host's `/etc/passwd` or execute `rm -rf /`.
*   **The Happy Path (Tasks):**
    1.  Architect enables the `Isolated Subprocess Runner` in the MCP Any configuration.
    2.  An agent calls a local tool `analyze-logs --path /tmp/logs`.
    3.  MCP Any intercepts the call and spawns the tool inside a restricted `nsjail` container.
    4.  The container only has read-only access to `/tmp/logs` and no access to the rest of the host.
    5.  The tool completes, and the output is returned to the agent securely.

## 4. Design & Architecture
*   **System Flow:**
    - **Interception**: The `IsolatedRunnerMiddleware` catches any tool call tagged as `type: command`.
    - **Sandbox Selection**: Based on configuration and host capabilities, it selects a driver (e.g., `NsJailDriver`, `DockerDriver`).
    - **Environment Preparation**: It mounts the required workspace directories and injects a sanitized set of environment variables.
    - **Execution**: The command is executed within the sandbox.
    - **Cleanup**: Ephemeral files and the sandbox environment are destroyed after execution.
*   **APIs / Interfaces:**
    - **Config**: New `isolation` block in `config.yaml` to define sandbox levels and allowed paths.
*   **Data Storage/State:** Per-execution workspace states are stored in a temporary, encrypted volume that is wiped upon completion.

## 5. Alternatives Considered
*   **User-Based Isolation**: Running tools as a low-privilege system user. *Rejected* because it doesn't provide strong enough file system or network isolation compared to modern containerization/namespacing.
*   **Full Virtual Machines (Firecracker)**: *Rejected* for local execution due to overhead and complexity; namespacing/gVisor provides a better balance for short-lived tool calls.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The runner follows the principle of least privilege. No tool has access to anything on the host by default. All access must be explicitly granted in the tool's security contract.
*   **Observability:** All sandbox events (start, stop, resource violations) are logged and visualized in the "Security Dashboard."

## 7. Evolutionary Changelog
*   **2026-02-27:** Initial Document Creation.
