# Design Doc: Neural-Active Shard Validator (NASV)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the introduction of Neural-Active Sharding (NAS) in OpenClaw, agent environments now perform predictive state swapping to eliminate latency. However, "Shard Shadowing" (CVE-2026-10101) allows malicious subagents to spoof shard metadata, redirecting the memory bus to unauthorized context fragments.

NASV provides the "Metadata Guardian" layer within MCP Any, ensuring that every predictive shard transition is cryptographically verified against the mission-root intent before the hardware memory map is updated.

## 2. Goals & Non-Goals
* **Goals:**
    * Perform real-time validation of NAS metadata signatures.
    * Bind shard transitions to hardware-attested reasoning intents.
    * Neutralize metadata-spoofing attacks (Shard Shadowing).
* **Non-Goals:**
    * Implementing the predictive sharding logic itself (handled by OpenClaw).
    * Managing the physical allocation of RAM (handled by the OS/DME Broker).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Architect
* **Primary Goal:** Prevent a specialist subagent from accessing parent context via NAS metadata spoofing.
* **The Happy Path (Tasks):**
    1. OpenClaw agent predicts next tool requirement and issues a NAS "Pre-warm" signal.
    2. MCP Any intercepts the NAS metadata packet.
    3. NASV verifies the shard's SHA-256 hash and metadata signature against the mission-root manifest.
    4. NASV performs a "Semantic Fit" check to ensure the shard aligns with the active reasoning step.
    5. Upon verification, NASV signals the DME Broker to update the memory map.

## 4. Design & Architecture
* **System Flow:**
    [OpenClaw NAS Engine] -> (NAS Metadata) -> [NASV Middleware] -> (Verification) -> [DME Broker] -> [Hardware Enclave]
* **APIs / Interfaces:**
    * `VerifyShardTransition(shard_id, metadata_sig, reasoning_intent)`
* **Data Storage/State:**
    * Transient verification cache linked to the Mission-Root Registry.

## 5. Alternatives Considered
* **Disabling NAS:** Rejected due to unacceptable latency impacts on deep swarms.
* **Logical-Only Validation:** Rejected because it cannot prevent low-level memory-mapped escapes once the bus is re-routed.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Mandates that no shard is "Pre-warmed" without a valid, mission-bound cryptographic anchor.
* **Observability:** Logs all "Shadowing" attempts in the Security Dashboard.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
