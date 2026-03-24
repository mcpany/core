<!--
Copyright (C) 2026 Author(s) of MCP Any
SPDX-License-Identifier: Apache-2.0
-->

# Design Doc: Teammate Sovereignty Enforcer (TSE)
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
With the transition of Claude Code and other frameworks to "Agent Teams," multiple agents now work in parallel on the same mission. Today's market sync revealed that these parallel teammates often share too much context or can be coerced into "State Smearing," where one teammate's lower-trust output pollutes the high-trust reasoning of another. Additionally, the emergence of "Logic Bombs" in agent-generated code (CVE-2026-30741) requires that teammate isolation extends to the reasoning path itself.

The Teammate Sovereignty Enforcer (TSE) provides a kernel-level isolation layer within MCP Any. It ensures that every teammate in a swarm operates within a cryptographically isolated "Sovereignty Shard," preventing unauthorized state access, cross-teammate impersonation, and un-sanitized logic propagation.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide cryptographically bound isolation for parallel teammates within a single mission.
    * Enforce "Auth-at-the-Pipe" for inter-agent communication using kernel-level named pipes.
    * Bind teammate identities to hardware-attested (TPM) mission-root tokens.
    * Integrate with the Logic-Sovereignty Validator (LSV) to detect cross-shard logic poisoning.
    * Automatically revoke all teammate-specific capabilities upon task completion.
* **Non-Goals:**
    * Replacing the primary mission-root security policy.
    * Providing OS-level virtualization (e.g., Docker) for the LLM itself.

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Orchestrator (Lead Agent)
* **Primary Goal:** Coordinate a specialist "Executor" and a "Security Auditor" teammate without the Auditor being able to see the Executor's privileged environment variables.
* **The Happy Path (Tasks):**
    1. The Lead Agent initializes a TSE mission with a TPM-signed root token.
    2. The Lead spawns an "Executor" teammate; TSE issues a bound identity shard.
    3. The Lead spawns an "Auditor" teammate; TSE issues a separate, isolated identity shard.
    4. TSE mandates that all inter-teammate communication occurs via isolated named pipes managed by the gateway.
    5. The TSE Kernel enforces "Teammate-Locked Shards," ensuring the Auditor cannot see the Executor's high-trust connection strings.
    6. Upon task completion, the TSE Kernel forcefully revokes all teammate-specific capabilities.

## 4. Design & Architecture
* **System Flow:**
    * **Isolation Kernel**: Manages the lifecycle of "Sovereignty Shards."
    * **Auth-at-the-Pipe**: Inter-teammate transport that validates mission-root tokens for every packet.
    * **Logic-Aware Shard Guards**: Real-time filters that interface with the LSV to block malicious reasoning fragments.

## 5. Alternatives Considered
* **OS-Level Namespacing (Docker)**: Rejected as too high-overhead for ephemeral agent sessions and difficult to map to the internal reasoning states of LLMs.
* **Flat Intent Isolation (RAMS)**: Evolved into TSE. While RAMS isolated memory, it lacked the cross-teammate coordination awareness required for horizontal "Agent Teams."

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: TSE enforces the "Principle of Least Privilege" at the teammate level. A compromised Auditor cannot impersonate the Lead or Executor teammates.
* **Observability**: Every inter-teammate message and shard access is logged with the TSE Mission ID and Teammate ID, providing a forensic-grade audit trail.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
### Update: 2026-06-18 - Logic-Path Interdiction Integration
**Context:** Today's research revealed a surge in "Logic Bombs" injected via Request-Side prompt injection (CVE-2026-30741).
**Architecture Adjustment:**
- Mandatory integration with the Logic-Sovereignty Validator (LSV) in Section 2 and 4.
- Introducing "Logic-Aware Shard Guards" to prevent un-attested reasoning fragments from being shared between teammates.
**Security Impact:** Prevents "Refinement Drift" and ensures that malicious logic from a single teammate cannot compromise the mission-root audit trace.
