# Design Doc: Atomic Lease Revocation (ALR) Controller
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
The disclosure of "Lease-Racing" vulnerabilities in mission-bound hardware leases reveals a critical gap in agentic security: the temporal window between software-level mission completion and hardware-level capability revocation. Currently, subagents can exploit a 50ms race condition to execute unauthorized tool calls after a mission is technically over but before the TPM has processed the revocation signal.

MCP Any needs to bridge this gap by introducing an authoritative hardware execution monitor that ensures revocation is physically atomic and precedes any post-mission execution.

## 2. Goals & Non-Goals
* **Goals:**
    * Ensure physically atomic revocation of task-bound capabilities.
    * Utilize TPM-bound monotonic heartbeats to synchronize software and hardware state.
    * Neutralize TOCTOU vulnerabilities in high-privilege tool execution.
* **Non-Goals:**
    * Replacing the primary orchestration layer for task assignment.
    * Managing the content of tool-specific security policies (delegated to Policy Firewall).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Enterprise Swarm Orchestrator
* **Primary Goal:** Prevent "final-second" unauthorized tool execution by subagents during mission teardown.
* **The Happy Path (Tasks):**
    1. Orchestrator issues a mission-bound capability lease via MCP Any.
    2. ALR Controller initializes a TPM-bound monotonic heartbeat for the subagent session.
    3. Subagent completes the task and signals mission completion.
    4. ALR Controller interdicts the heartbeat, triggering immediate physical revocation of the capability in the hardware enclave.
    5. MCP Any verifies revocation status before acknowledging mission termination to the subagent.

## 4. Design & Architecture
* **System Flow:**
    [Subagent Process] --(Heartbeat)--> [ALR Controller] --(TPM Signal)--> [Secure Enclave]
                                          |
                                [Mission Control] --(Revoke)--> [ALR Controller]
* **APIs / Interfaces:**
    * `ALR.InitializeLease(taskID, capabilityToken)`
    * `ALR.Heartbeat(sessionToken, monotonicCounter)`
    * `ALR.AtomicRevoke(taskID)`
* **Data Storage/State:**
    * TPM-backed session registry in kernel-bound memory.

## 5. Alternatives Considered
* **Software-only Timers:** Rejected due to lack of physical atomicity and susceptibility to process-scheduling delays.
* **Kernel-level Kill Signals:** Rejected because they do not provide cryptographic proof of capability withdrawal required for Zero Trust auditing.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All revocation signals are hardware-signed and monotonic, preventing replay attacks.
* **Observability:** Revocation latency is tracked via the Performance-Optimized Side-Channel Defense metrics.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
