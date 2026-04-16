# Design Doc: Reasoning-Enclave Sharding (RES) Broker
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
Current multi-agent coordination models rely on logical isolation (e.g., separate process namespaces or software-based memory segmentation) to prevent state pollution. However, the emergence of "Memory-Mapped Escape" vulnerabilities and speculative side-channel attacks in high-density swarms (e.g., OpenClaw) prove that software isolation is insufficient. Agents can speculatively probe or time the execution of sibling agents to exfiltrate mission context.

The RES Broker solves this by moving state isolation from the software layer into hardware-locked memory enclaves (TPM/SEP/SGX). Each specialist agent reasoning path is physically sharded at the hardware level.

## 2. Goals & Non-Goals
* **Goals:**
    * Physically isolate specialist agent reasoning paths using hardware-locked memory enclaves.
    * Neutralize speculative side-channel probes between peer agents in a shared bus.
    * Provide a unified broker for managing enclave lifecycle (allocation, attestation, reclamation).
    * Support hardware-accelerated "Enclave Resumption" to minimize coordination latency.
* **Non-Goals:**
    * Replacing existing software sandboxing (Docker/gVisor); it acts as an additional hardware layer.
    * Managing non-agent system memory.
    * Providing cross-device memory sharing (this is handled by the AMT Broker).

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Trust Swarm Orchestrator
* **Primary Goal:** Securely execute a PII-handling specialist agent alongside a public-web-searching subagent without risking context leakage via memory side-channels.
* **The Happy Path (Tasks):**
    1. Orchestrator requests a "Sensitive Mission" branch.
    2. RES Broker allocates two hardware-locked enclaves (Shards A and B).
    3. PII Specialist is bound to Shard A; Web Specialist is bound to Shard B.
    4. Specialist agents reason locally within their enclaves.
    5. RES Broker mediates all state handoffs via the "Zero-Copy Memory Broker," ensuring that binary fragments are cryptographically re-keyed for the target enclave.
    6. Any speculative probe from the Web Specialist into the PII Specialist's memory space is physically blocked by the hardware CPU gates.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        subgraph Hardware Layer (TPM/SEP)
            E1[Enclave Shard A]
            E2[Enclave Shard B]
        end
        subgraph MCP Any Core
            B[RES Broker] -->|Manages| E1
            B -->|Manages| E2
            Z[Zero-Copy Memory Broker] <--> B
        end
        A1[Specialist Agent 1] --> E1
        A2[Specialist Agent 2] --> E2
    ```
* **APIs / Interfaces:**
    * `res.AllocateShard(agentID, securityLevel) -> ShardID`: Provisions a new hardware-locked enclave.
    * `res.BindProcess(shardID, pid)`: Hardens a process's memory space using the hardware enclave.
    * `res.AttestShard(shardID) -> AttestationReport`: Generates a hardware-signed proof of shard integrity.
* **Data Storage/State:**
    * **Enclave Mapping Table:** Kernel-bound registry of active shards and their bound agent IDs.
    * **Hardware Key Store:** Securely manages the ephemeral encryption keys used to "wrap" shards for handoffs.

## 5. Alternatives Considered
* **Pure Logical Sharding (Software-based):** Rejected due to vulnerability to Spectre/Meltdown style side-channel attacks in shared agent buses.
* **Virtual Machine Isolation:** Rejected due to high cold-start latency (1s+) and resource overhead. RES provides sub-millisecond isolation.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Every enclave allocation requires a hardware-attested mission root token.
* **Observability:** Integrated with the "Enclave Isolation Monitor" UI component for real-time visualization of hardware memory boundaries.
* **Performance:** Uses "Fast-Path Enclave Resumption" to reduce the 15% attestation tax introduced by hardware calls.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
