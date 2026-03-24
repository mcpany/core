# Design Doc: Hierarchical Kill Switch Broker
**Status:** Draft
**Created:** 2026-05-23

## 1. Context and Scope
Autonomous agent swarms, especially those with deep recursive lineages, are difficult to stop once they deviate from a safe reasoning path. A "Runaway Swarm" can continue executing tasks and spawning descendants across multiple nodes even after a Parent agent has been flagged or terminated. Existing kill switches are often limited to single processes or depend on software-level polling that can be ignored or bypassed by a compromised agent.

The Hierarchical Kill Switch Broker (HKSB) provides hardware-attested, recursive revocation of agent capabilities. It implements the OpenClaw v2.0 global kill switch protocol. When a human supervisor or high-level security policy signs a "Kill Signal" for a Mission Root, HKSB triggers a recursive lockdown that propagates through the entire intent branch in under 50ms. By binding the revocation to hardware security modules (TPM/SEP), we ensure that descendant agents are physically blocked from tool calls and compute before they can even process the termination signal.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement recursive, hierarchical revocation of agent capabilities across a mission branch.
    * Achieve hardware-attested lockdown within a 50ms threshold.
    * Use TPM-bound "Termination Nonces" to physically invalidate session tokens.
    * Support "Emergency Swarm Freezing" for forensic analysis of compromised swarms.
* **Non-Goals:**
    * Managing process termination (HKSB blocks *capabilities*; OS handles process reaping).
    * Providing single-agent process management (HKSB is for hierarchical intent branches).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Operations Center (SOC) Specialist
* **Primary Goal:** Instantly halt a massive, multi-node agent swarm that has been hijacked by a "Monologue Pollution" attack.
* **The Happy Path (Tasks):**
    1. The Specialist identifies a compromised Parent agent in the dashboard.
    2. The Specialist signs a "Mission-Wide Revocation" command using their hardware key.
    3. MCP Any HKSB receives the signed signal and identifies the root Inode and Intent-ID.
    4. HKSB broadcasts a "Hardware Halt" to all connected nodes sharing that Intent lineage.
    5. The local TPM on every node invalidates all session keys and Resource Leases (TBRI) associated with the branch.
    6. All subagents are physically blocked from subsequent tool calls and LLM requests within 42ms.
    7. The swarm is "Frozen" in its current state, allowing for isolated monologue review.

## 4. Design & Architecture
* **System Flow:**
    [Human Supervisor] -> [Revocation Signal (Signed)] -> [HKSB Master] -> [Local HKSB Agents] -> [TPM/SEP Invalidation] -> [Gateways (Blocked)]
    1. Master Broker verifies Supervisor signature.
    2. Master identifies all Intent IDs in the hierarchy.
    3. Master sends "Kill Nonce" to all node-local HKSB listeners.
    4. Local HKSB calls TPM to invalidate the session-bound Memory Keys and Resource Quotas.
* **APIs / Interfaces:**
    * `BroadcastKillSignal(mission_id, signature) -> bool`
    * `RegisterIntentHierarchy(parent_id, child_id)`
* **Data Storage/State:**
    * Intent hierarchy is maintained in a hardware-protected segment of the MCP Any internal state.
    * Revocation nonces are stored in TPM "Kill Slots."

## 5. Alternatives Considered
* **SIGKILL Propagation:** Rejected; slow and can be blocked by "unkillable" or zombie processes. HKSB blocks the *ability* to act regardless of process state.
* **Soft Revocation (API-based):** Rejected; vulnerable to latency and "Last-Second Action" bypasses.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Kill signals require hardware-level multi-factor attestation.
* **Observability:** "Kill Propagation Latency" is tracked as a critical system performance metric.

## 7. Evolutionary Changelog
* **2026-05-23:** Initial Document Creation.
