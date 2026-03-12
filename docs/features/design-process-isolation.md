# Design Doc: Process-Isolated Execution Runtime
**Status:** Draft
**Created:** 2026-03-17

## 1. Context and Scope
With the rise of multi-agent swarms, agents frequently delegate tasks to subagents or execute automated hooks. Recent research (CVE-2026-30112) has identified a critical vulnerability known as "Context Bleeding," where subagents can inadvertently access sensitive environment variables, secrets, or memory segments of the parent process.

MCP Any must provide a hardened, process-isolated execution environment for all tool calls and configuration hooks. By moving from simple logical isolation to strict process-level boundaries, we ensure that each agent identity operates in a clean, restricted sandbox, preventing the leakage of high-privilege credentials to lower-privilege sub-tasks.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Mandate process-level isolation for all tool and hook executions.
    *   Ensure zero "Context Bleeding" between parent and child agent processes.
    *   Provide a standardized, ephemeral sandbox for local command execution.
    *   Enforce resource limits (CPU, Memory) on isolated processes to prevent DoS.
*   **Non-Goals:**
    *   Full virtual machine isolation (too heavy for local tool use).
    *   Rewriting existing MCP servers to be process-aware (isolation happens at the gateway level).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security-Conscious Agent Developer
*   **Primary Goal:** Execute an automated project hook (e.g., `npm test`) without the hook process having access to the agent's `ANTHROPIC_API_KEY` or host SSH keys.
*   **The Happy Path (Tasks):**
    1.  The user clones a repository containing a project-local MCP Any configuration with an `after_load` hook.
    2.  MCP Any detects the hook and prompts the user for attestation.
    3.  Upon approval, MCP Any spawns an ephemeral, isolated container (or restricted process).
    4.  The isolated runtime is injected with *only* the specific environment variables required for the hook, explicitly excluding parent secrets.
    5.  The hook executes, results are returned to MCP Any, and the isolated environment is immediately destroyed.

## 4. Design & Architecture
*   **System Flow:**
    `Agent Request` -> `MCP Any Gateway` -> `Policy Engine (Validation)` -> `Isolation Manager` -> `Ephemeral Worker Process` -> `Tool/Hook Execution` -> `Result Return` -> `Worker Destruction`.
*   **APIs / Interfaces:**
    *   Internal `IsolationProvider` interface to support multiple backends (Docker, WebAssembly, or restricted `fork/exec`).
    *   `ExecutionRequest` schema updated to include `IsolationProfile` and `EnvironmentAllowList`.
*   **Data Storage/State:**
    *   No persistent state within the isolated runtime.
    *   Logs are streamed back to MCP Any and stored in the central audit log.

## 5. Alternatives Considered
*   **Logical Env Filtering:** Simply unsetting environment variables before `exec`. Rejected because it doesn't protect against memory-level bleeding or sophisticated side-channel attacks in shared runtimes.
*   **Full VM Isolation:** Using Firecracker or similar. Rejected for being too resource-intensive and introducing significant latency for simple tool calls.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The Isolation Manager follows a "Deny-by-Default" posture. Only explicitly allow-listed resources (files, env vars) are mounted into the worker process.
*   **Observability:** Each isolated execution is assigned a unique `TraceID`, and resource consumption (CPU/Mem) is reported to the telemetry dashboard to detect "Spiral of Death" loops.

## 7. Evolutionary Changelog
*   **2026-03-17:** Initial Document Creation.
