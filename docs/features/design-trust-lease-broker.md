# Design Doc: Distributed Trust Lease Broker
**Status:** Draft
**Created:** 2026-04-19

## 1. Context and Scope
As agent swarms grow deeper and more autonomous, the "Attestation Tax"—the latency introduced by performing a hardware-bound security check for every single tool call—has become the primary bottleneck for real-time agentic reasoning. The Distributed Trust Lease Broker implements the UACO v2.5 "Low-Frequency Trust Attestation" (LFTA) standard, allowing agents to "lease" an attested security posture for a time-bound burst of activity.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a high-performance broker for UACO v2.5 LFTA trust leases.
    * Reduce inter-agent coordination latency by 90% via ephemeral capability caching.
    * Bind all leases to a cryptographically verified "Mission Intent" and hardware state (via RIM).
    * Provide automatic revocation of all active leases upon detection of "Resident Integrity Drift."
* **Non-Goals:**
    * Providing permanent, long-term access to tools (leases must be ephemeral).
    * Replacing the per-call "Policy Firewall" for high-risk actions (high-risk still requires individual attestation).

## 3. Critical User Journey (CUJ)
* **User Persona:** Deep Swarm Orchestrator
* **Primary Goal:** Execute a sequence of 50 local file-reads and 20 code-analyses across a 4-level agent hierarchy in under 2 seconds.
* **The Happy Path (Tasks):**
    1. Parent agent initiates a "Mission Intent" with a Root Attestation.
    2. MCP Any's Lease Broker issues an LFTA token (Trust Lease) bound to this intent and the current hardware Inode state.
    3. Subagents inherit this LFTA token.
    4. Subagents perform tool calls by presenting the lease token; MCP Any validates the token in sub-milliseconds without re-triggering TPM signatures.
    5. The Resident Integrity Monitor (RIM) continuously validates the host state in the background.
    6. Once the 5-minute lease expires or the mission completes, the token is invalidated.

## 4. Design & Architecture
* **System Flow:**
    `[Root Attestation] -> [Lease Broker] -> [LFTA Token] -> [Subagent Tool Calls]`
    `[RIM] --(Revocation Signal)--> [Lease Broker]`
* **APIs / Interfaces:**
    * `/v1/lease/issue`: Generates a time-bound LFTA token bound to an intent.
    * `/v1/lease/verify`: Fast-path validator for lease tokens.
    * `/v1/lease/revoke`: Explicitly invalidates a lease or intent-branch.
* **Data Storage/State:**
    * Leases are stored in a high-speed, memory-mapped "Lease Registry" within MCP Any, synchronized with the Shared KV Store (Blackboard).

## 5. Alternatives Considered
* **Per-Call Hardware Attestation**: Rejected due to 100ms+ latency per call, which causes "Cognitive Stall" in deep swarms.
* **Static API Keys**: Rejected as they lack intent-binding and cannot be automatically revoked upon environment drift.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All LFTA tokens are cryptographically bound to the hardware state. If the RIM detects a symlink-race or config swap, all associated leases are instantly nukated.
* **Observability:** Performance gains and lease lifecycle events are tracked in the "Fast-Path Attestation Visualizer."

## 7. Evolutionary Changelog
* **2026-04-19:** Initial Document Creation.
