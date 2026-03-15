# Design Doc: eBPF Boundary Guard
**Status:** Draft
**Created:** 2026-04-15

## 1. Context and Scope
As AI agents increasingly execute local tools (shell commands, scripts, binaries), the security boundary between the agent and the host filesystem/network becomes critical. Existing sandboxing (Docker, gVisor) often has high overhead or can be bypassed if misconfigured. The eBPF Boundary Guard provides kernel-level monitoring and enforcement for MCP Any tools, ensuring that they cannot exceed their declared capabilities.

## 2. Goals & Non-Goals
* **Goals:**
    * Real-time monitoring of system calls (file I/O, network, process execution) for local tools.
    * Immediate termination of any tool that attempts to access resources not declared in its manifest.
    * Minimal performance overhead (sub-microsecond latency).
    * Immutable audit logs of all denied system calls.
* **Non-Goals:**
    * Replacing existing sandboxing (it is an additional layer of defense).
    * Providing a full virtualized environment.
    * Hardening the entire host OS (scope is limited to MCP Any tool child processes).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Local Developer
* **Primary Goal:** Execute a community-sourced MCP tool for image processing without risk of it exfiltrating `~/.ssh/id_rsa`.
* **The Happy Path (Tasks):**
    1. User installs a new MCP tool with a manifest declaring `fs:read:/tmp` and `net:none`.
    2. MCP Any starts the tool process and attaches an eBPF probe to its PID and children.
    3. The tool attempts to read `/tmp/image.png` (Allowed).
    4. The tool attempts to read `~/.ssh/id_rsa`.
    5. The eBPF Boundary Guard intercepts the `openat` syscall, detects the violation, and returns `EACCES` or kills the process.
    6. MCP Any logs the security violation and alerts the user.

## 4. Design & Architecture
* **System Flow:**
    * MCP Any (User Space) -> eBPF Loader -> Linux Kernel (eBPF Programs).
    * When a tool is executed, MCP Any loads the tool's security manifest into a BPF Map keyed by PID.
    * eBPF programs attached to tracepoints (e.g., `sys_enter_openat`, `sys_enter_connect`) check the syscall arguments against the PID's allowed map.
* **APIs / Interfaces:**
    * `ManifestVerifier`: Internal Go service that translates MCP capability tokens into eBPF-compatible bitmasks.
    * `ViolationObserver`: Interface for receiving and logging real-time alerts from the kernel.
* **Data Storage/State:**
    * BPF Maps: LRU Hash for PID-to-Manifest mapping.
    * Ring Buffer: For sending security events from kernel to user space.

## 5. Alternatives Considered
* **gVisor:** Rejected due to performance overhead and complexity of running in standard local environments.
* **AppArmor/SELinux:** Rejected because dynamic, per-process policy generation is slow and difficult to manage at the granularity required by AI agent swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The eBPF programs themselves are verified by the kernel before loading. No tool can escape the guard even if it gains root (if the guard is properly hardened).
* **Observability:** Integration with Prometheus for tracking "Denied Syscall Rate."

## 7. Evolutionary Changelog
* **2026-04-15:** Initial Document Creation.
