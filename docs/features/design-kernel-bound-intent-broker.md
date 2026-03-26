# Design Doc: Kernel-Bound Intent (KBI) Broker
**Status:** Draft
**Created:** 2026-05-02

## 1. Context and Scope
As AI agent swarms move from simple automation to deep, multi-agent reasoning, the risk of "User-Space Intent Mutation" increases. If an agent framework is compromised, malicious code can intercept and modify signed intents before they reach the security gateway.

MCP Any needs to bridge the gap between application-level logic and OS-level enforcement. By integrating with OpenClaw's KBIA (Kernel-Bound Intent Attestation), we can ensure that an agent's intent is cryptographically bound to its execution process via eBPF hooks.

## 2. Goals & Non-Goals
* **Goals:**
    * Translate eBPF-attested intent tokens into MCP capability grants.
    * Provide hardware-bound integrity proofs for intent lineage.
    * Synchronize kernel-level resource quotas with application-level budgets.
* **Non-Goals:**
    * Replacing the OS kernel's native security (SELSM/AppArmor).
    * Providing general-purpose eBPF monitoring.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Swarm Orchestrator
* **Primary Goal:** Prevent a compromised subagent from mutating its intent to access "unauthorized" tools in a shared session.
* **The Happy Path (Tasks):**
    1. Parent agent signs a mission intent and broadcasts it via the KBI Broker.
    2. The Broker registers the intent hash with a local eBPF monitor.
    3. Subagent attempts to call a tool.
    4. KBI Broker verifies that the calling process ID (PID) matches the attested eBPF state before forwarding the request to the MCP Gateway.

## 4. Design & Architecture
* **System Flow:**
    [Agent Process] --(eBPF Hook)--> [Linux Kernel]
          |                           |
    (Tool Call)                  (Signed Intent)
          |                           |
          v                           v
    [MCP Any Gateway] <---(Verify)-- [KBI Broker]

* **APIs / Interfaces:**
    * `RegisterKernelIntent(intent_hash, pid)`: Binds a signed intent to a specific OS process.
    * `ValidateKernelBoundCall(pid, request)`: Verifies call integrity via kernel state.
* **Data Storage/State:**
    * Transient eBPF maps for PID-to-Intent mapping.
    * Secure Enclave storage for hardware-bound mission keys.

## 5. Alternatives Considered
* **User-Space Only Signing:** Rejected due to vulnerability to user-space interception and "Context-Mirroring" attacks.
* **Full VM Isolation:** Rejected due to the extreme performance overhead in deep swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Uses hardware-bound (TPM) signatures and kernel-resident monitors to enforce zero-trust at the process level.
* **Observability:** Logs eBPF violation events to the Local Security Audit Dashboard.

## 7. Evolutionary Changelog
* **2026-05-02:** Initial Document Creation.
