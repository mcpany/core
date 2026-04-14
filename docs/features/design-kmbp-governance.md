# Design Doc: Kernel-Mediated Budget Pinning (KMBP)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
Current agent resource governance relies heavily on "Application-Level Throttling," where the gateway or framework monitors token usage and terminates sessions upon limit violation. However, the emergence of "Reasoning-Budget Hijacking" (RBH) has shown that compromised subagents can exploit latencies in application-level monitoring or bypass them entirely via out-of-band communication, leading to "Agentic DoS" and significant token exhaustion costs.

KMBP moves the enforcement of reasoning and compute budgets from the application layer to the Operating System kernel. By utilizing Linux \`cgroups\` (Control Groups) and macOS \`sandbox-exec\` (App Sandbox), MCP Any can "pin" a subagent process to a hardware-attested budget that is physically impossible to exceed, regardless of the subagent's internal reasoning state or attempts to evade monitoring.

## 2. Goals & Non-Goals
* **Goals:**
    * Enforce compute and memory limits for tool-executing subagents at the kernel level.
    * Synchronize hardware-attested token budgets with OS-level process priority and CPU quotas.
    * Provide sub-millisecond interdiction of subagents that attempt to "over-reason" beyond mission-root constraints.
    * Ensure resource sovereignty in horizontal teammate meshes by preventing sibling "Resource Squatting."
* **Non-Goals:**
    * Managing the LLM provider's internal quotas (focused on the local execution of the agent/runtime).
    * Optimizing the reasoning process itself (focused only on physical resource constraints).
    * Replacing existing application-level RBF (KMBP acts as the "hard" fallback).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Run 10 parallel subagents on a single workstation without a single "rogue" subagent consuming all available CPU or exceeding the total project token budget.
* **The Happy Path (Tasks):**
    1. The Orchestrator defines a "Mission Budget" containing a max CPU quota (e.g., 2 cores) and a max reasoning duration.
    2. MCP Any spawns the subagent within a dedicated Linux \`cgroup\` or macOS sandbox.
    3. The KMBP middleware "pins" the hardware-attested budget to the process ID (PID).
    4. The subagent attempts an infinite reasoning loop to bypass mission constraints.
    5. The Kernel detects the CPU quota violation and automatically throttles or "freezes" the process.
    6. MCP Any detects the "Frozen" state, revokes the subagent's identity tokens, and alerts the mission-root.

## 4. Design & Architecture
* **System Flow:**
    ```
    Mission Root -> [Budget Manifest] -> KMBP Middleware
    KMBP Middleware -> [Spawn Process] -> Linux Cgroups / macOS Sandbox
    Kernel -> [Continuous Enforcement] -> Subagent Process
    KMBP Middleware -> [Status Feedback] -> Reasoning-Budget Firewall (RBF)
    ```
* **APIs / Interfaces:**
    * `SetBudget(pid int, budget BudgetManifest) error`: Kernel-level binding of a process to a budget.
    * `GetUsage(pid int) (UsageStats, error)`: Direct kernel-read of consumed resources.
* **Data Storage/State:**
    * State is primarily maintained in the kernel (\`/sys/fs/cgroup\` or sandbox kernel state).
    * MCP Any maintains a mapping of Mission-Root Fragments to OS-level Group IDs.

## 5. Alternatives Considered
* **Pure Application-Level Throttling:** Rejected because it is vulnerable to "Reasoning Stall" attacks where the monitor itself is delayed by the agentic load.
* **Virtual Machine Isolation:** Rejected for local swarms due to the high memory and startup latency overhead. KMBP provides "VM-like" isolation at "Process-like" speeds.
* **User-Mode Throttling:** Rejected because it can be bypassed by subagents that exploit library-level vulnerabilities or direct syscalls.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** KMBP utilizes hardware-attested monotonic counters to ensure that "Usage" metrics cannot be tampered with by the subagent.
* **Observability:** Real-time "Kernel Throttling" metrics are exported to the "Mission Budget Dashboard" and the "Resource Attribution Overlay."

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
