# Design Doc: Kernel-Level Inode Watcher (KLIW)
**Status:** Draft
**Created:** 2026-05-01

## 1. Context and Scope
[The "BoryptKernel" variant has introduced a catastrophic new vulnerability: the manipulation of project-local file buffers in kernel memory before they are written to disk. This bypasses all existing user-space Inode-pinning and fsnotify mechanisms. MCP Any needs to move its filesystem validation layer into the kernel to provide deterministic, hardware-bound protection for agent configurations.]

## 2. Goals & Non-Goals
* **Goals:**
    * Integrate with OpenClaw v2026.4.1 kernel-level hooking for real-time Inode verification.
    * Provide kernel-enforced attestation for project-local configurations (e.g., `.claude/settings.json`).
    * Neutralize buffer-manipulation attacks performed by kernel-aware malware.
* **Non-Goals:**
    * Replacing the entire filesystem driver.
    * Managing general-purpose file encryption for the host.

## 3. Critical User Journey (CUJ)
* **User Persona:** [Enterprise Security Architect]
* **Primary Goal:** [Ensure that an agent's project-local settings cannot be modified by kernel-level malware during a high-stakes mission.]
* **The Happy Path (Tasks):**
    1. The user enables "Kernel-Enforced Attestation" in the MCP Any security policy.
    2. MCP Any initializes the KLIW middleware and registers project-local Inodes with the kernel-hooking module.
    3. A kernel-aware malware attempts to intercept and modify the file buffer for `.claude/settings.json` in memory.
    4. The kernel-hooking module detects the unauthorized buffer mutation and triggers a KLIW violation signal.
    5. MCP Any immediately halts the active agent swarm and alerts the user of a kernel-level compromise.

## 4. Design & Architecture
* **System Flow:**
    [Agent Runtime] <-> [MCP Any KLIW Middleware] <-> [OpenClaw Kernel Module] <-> [Filesystem VFS]
* **APIs / Interfaces:**
    * `RegisterKernelWatch(path string, expected_hash []byte)`: Registers a path for kernel-level monitoring.
    * `onKernelViolation(signal ViolationSignal)`: Callback for kernel-level attestation failures.
* **Data Storage/State:**
    * Kernel-space Inode table (managed by the module).
    * User-space attestation registry (managed by MCP Any pkg/security).

## 5. Alternatives Considered
[User-space Inode Pinning: Rejected because it is vulnerable to buffer-level manipulation in kernel memory.]
[eBPF-based monitoring: Considered, but direct VFS hooking via the OpenClaw module provides lower latency and more deterministic enforcement for this specific attack vector.]

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** [How do we prevent a compromised kernel from poisoning the KLIW monitor? We utilize hardware-bound attestation (TPM) to verify the integrity of the kernel module itself at boot.]
* **Observability:** [Standardized logging of kernel-level violation signals to the local security audit log.]

## 7. Evolutionary Changelog
* **2026-05-01:** Initial Document Creation.
