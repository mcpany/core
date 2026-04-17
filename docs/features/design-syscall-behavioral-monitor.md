# Design Doc: Syscall-Level Behavioral Monitor (SLBM)
**Status:** Draft
**Created:** 2026-04-17

## 1. Context and Scope
As AI agents gain higher autonomy, traditional security boundaries like API-level filtering and filesystem allow-lists are becoming insufficient. Malicious subagents or compromised tools can exploit shell-fallbacks and binary "hook" injections to perform unauthorized actions directly at the operating system level. Today's market analysis confirms that syscall-level instrumentation (eBPF) is the new frontier for detecting these "malicious agent" patterns.

MCP Any needs a kernel-resident monitoring service that can observe tool execution in real-time, matching process behavior against a set of "Agentic Safety Profiles" and blocking deviating syscalls before they impact the host.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement real-time syscall monitoring for all MCP tool executions using eBPF/Falco-style signals.
    * Provide a set of default "Agentic Safety Profiles" (e.g., "Standard CLI", "Database Access", "ReadOnly").
    * Enable sub-millisecond interdiction (blocking) of unauthorized syscalls.
    * Generate hardware-attested behavioral logs for compliance auditing.
* **Non-Goals:**
    * Full OS-level sandboxing (handled by external providers like gVisor or Docker). SLBM is a monitoring and policy enforcement layer *within* the execution context.
    * Monitoring processes not initiated by the MCP Any gateway.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Enterprise Architect
* **Primary Goal:** Prevent a coding agent from exfiltrating environment variables via a rogue socket connection disguised as a standard tool call.
* **The Happy Path (Tasks):**
    1. The architect defines a "ReadOnly-NoNetwork" safety profile for the project.
    2. An agent initiates a tool call to `list_files`.
    3. The tool (compromised) attempts to execute `socket()` to connect to an external IP.
    4. SLBM detects the `socket` syscall, matches it against the "ReadOnly-NoNetwork" profile, and blocks the call.
    5. The tool call fails with a "Permission Denied" error at the kernel level.
    6. MCP Any logs the violation and alerts the user via the Security Dashboard.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent] --> B[MCP Any Gateway]
        B --> C[Tool Executor]
        C --> D[Target Tool Process]
        E[SLBM eBPF Probe] -- Monitor --> D
        E -- Signal --> F[Policy Engine]
        F -- Block/Allow --> D
        F -- Log --> G[Behavioral Audit Store]
    ```
* **APIs / Interfaces:**
    * `POST /v1/security/profiles`: Create or update a behavioral profile.
    * `GET /v1/security/violations`: Retrieve behavioral violation logs.
    * Internal Hook: `slbm.AttachProcess(pid, profile_id)`
* **Data Storage/State:**
    * Profiles are stored in the local configuration database.
    * Violation logs are stored in a hardware-attested SQLite buffer, exportable to SIEM.

## 5. Alternatives Considered
* **User-Space Only Monitoring**: Rejected due to high latency and susceptibility to "unshare" or shell-fallback escapes that bypass user-space hooks.
* **Full Docker Isolation**: Docker provides isolation but not granular behavioral monitoring. SLBM provides the *visibility* required to understand *why* an agent is acting maliciously, even within a container.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** SLBM operates as a "Kernel Guardrail" that doesn't trust the target tool process. All policy decisions are made outside the tool's address space.
* **Observability:** Integrated with the "Visual Attention Dashboard" to show users which syscalls were blocked in real-time.

## 7. Evolutionary Changelog
* **2026-04-17:** Initial Document Creation.
