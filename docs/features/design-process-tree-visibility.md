# Design Doc: Process Tree Visibility Middleware

**Status:** Draft
**Created:** 2026-03-09

## 1. Context and Scope
As AI agents gain more autonomy, they are increasingly entrusted with executing system-level tools (CLI commands, scripts, etc.). Recent research into frameworks like OpenClaw has shown that agents can sometimes bypass high-level security filters by spawning sub-processes or executing unexpected shell commands ("Shadow Agent" behavior). Current MCP logging typically only captures the initial tool call, leaving a massive observability gap. This middleware aims to close that gap by capturing and logging the full process tree of any tool execution initiated by an agent.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Capture full process lineage (PID, PPID, Command Line) for all tools executed via the `Command Adapter`.
    *   Log child processes spawned by the initial tool.
    *   Integrate process tree data into the `Unified Agent Audit Logger`.
    *   Provide real-time visibility into tool execution via the UI.
*   **Non-Goals:**
    *   Preventing sub-process spawning (this is the role of the `Policy Firewall`).
    *   Kernel-level auditing (we will rely on user-space process tracking like `ebpf` or platform-specific APIs).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security Operations Center (SOC) Analyst.
*   **Primary Goal:** Investigate why an agent-initiated script attempted to access a sensitive configuration file.
*   **The Happy Path (Tasks):**
    1.  The analyst opens the MCP Any "Audit Mesh" dashboard.
    2.  They locate a specific `tools/call` for a `deploy-script`.
    3.  The analyst expands the call to see the "Execution Tree."
    4.  The tree reveals that the `deploy-script` (PID 1234) spawned a sub-process `curl` (PID 1235) that attempted to exfiltrate `~/.aws/credentials`.

## 4. Design & Architecture
*   **System Flow:**
    - **Hooking**: The middleware wraps the `Command Adapter`'s execution loop.
    - **Tracking**: On Linux, it utilizes `ebpf` or `ptrace` (if allowed) to monitor `fork`/`exec` syscalls from the parent PID. On macOS/Windows, it uses platform-specific process accounting APIs.
    - **Data Aggregation**: Process events are correlated by `ToolCallID` and streamed to the `Audit Logger`.
*   **APIs / Interfaces:**
    - `ProcessTreeProvider` interface for platform-specific implementations.
    - `ProcessNode` data structure: `{pid: int, ppid: int, command: string, startTime: timestamp, endTime: timestamp}`.
*   **Data Storage/State:** Transient process mapping is kept in memory during execution; the final tree is persisted in the `Audit Logger` (SQLite/Postgres).

## 5. Alternatives Considered
*   **Basic Logging**: Only logging the top-level command. *Rejected* because it fails to capture "Shadow Agent" sub-processes.
*   **External Auditing (OSQuery/CrowdStrike)**: Relying on external tools for process visibility. *Rejected* because it lacks the context of which *AI agent* initiated the call, making correlation difficult.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The process tree itself becomes part of the attestation record. Any unauthorized sub-process can trigger an immediate "Kill Switch" in future iterations.
*   **Observability:** Visualizing trees in the UI requires a new "Execution Waterfall" component.

## 7. Evolutionary Changelog
*   **2026-03-09:** Initial Document Creation.
