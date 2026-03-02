# Design Doc: Local Runtime Sandbox Adapter

**Status:** Draft
**Created:** 2026-03-01

## 1. Context and Scope
As highlighted in recent OpenClaw research, the primary risk in agentic workflows is "the last mile" of execution. When an agent calls a local tool (e.g., a python script to process data), it typically runs with the permissions of the host user. The Local Runtime Sandbox Adapter aims to decouple tool execution from the host environment by wrapping every tool call in a restricted, policy-enforced sandbox.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Isolate tool execution from the host filesystem and network by default.
    *   Support granular syscall filtering (e.g., blocking `execve` unless explicitly permitted).
    *   Provide ephemeral, per-call environments that are destroyed after the tool returns.
    *   Allow declarative resource limits (CPU, Memory) per tool.
*   **Non-Goals:**
    *   Replacing full-blown container orchestrators like Kubernetes.
    *   Providing a persistent OS-level sandbox (it is per-tool call).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security-conscious Agent Developer.
*   **Primary Goal:** Allow an agent to run a community-contributed "CSV Parser" tool without risking the agent reading the user's `.ssh` directory.
*   **The Happy Path (Tasks):**
    1.  Developer enables the `LocalRuntimeSandbox` middleware in `config.yaml`.
    2.  Agent receives a task to "Analyze `data.csv`".
    3.  MCP Any intercepts the tool call to `csv_parser`.
    4.  The adapter creates a gVisor-based or WASM-based container, mounts *only* `data.csv`, and executes the tool.
    5.  Tool finishes; the sandbox is immediately purged.

## 4. Design & Architecture
*   **System Flow:**
    - **Interception**: The `SandboxMiddleware` wraps the `mcpserver.ExecuteTool` call.
    - **Environment Prep**: Creates a temporary rootfs with only the necessary binaries and inputs.
    - **Execution**: Uses a lightweight runtime (e.g., `runsc` for gVisor, or a WASI runtime for supported tools).
    - **Cleanup**: Synchronous deletion of all temporary resources.
*   **APIs / Interfaces:**
    - Extends the `ToolDefinition` with a `sandbox_policy` field.
*   **Data Storage/State:** Statelss by design. All context must be passed via explicit mounts.

## 5. Alternatives Considered
*   **Docker-based isolation**: *Rejected* for per-call overhead and complex daemon requirements.
*   **OS-level User Namespaces**: *Rejected* because they don't provide sufficient syscall protection against kernel exploits.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Follows the "Default Deny" principle. No filesystem or network access is granted unless specified in the `sandbox_policy`.
*   **Observability:** Every sandbox lifecycle event (start, exec, stop) is logged to the Audit Trail.

## 7. Evolutionary Changelog
*   **2026-03-01:** Initial Document Creation.
