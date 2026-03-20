<!--
Copyright (C) 2026 Author(s) of MCP Any
SPDX-License-Identifier: Apache-2.0
-->

# Design Doc: Teammate Sovereignty Enforcer (TSE)
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
With the official transition of Claude Code to "Agent Teams," AI swarms are moving from sequential task execution to high-density parallel coordination. This shift introduces significant risks of "State Smearing," where parallel teammates accidentally pollute or exfiltrate each other's session state due to a lack of cryptographically bound isolation.

The Teammate Sovereignty Enforcer (TSE) aims to act as the authoritative "Isolation Kernel" for these horizontal meshes. It ensures that every teammate in a team (e.g., Claude lead, OpenClaw specialist) operates within a cryptographically locked, mission-anchored boundary that persists across all inter-agent communications and state handoffs.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide hardware-attested isolation for parallel teammate sessions.
    * Prevent cross-teammate "State Smearing" in the Shared Blackboard and Mailbox.
    * Mandate mission-anchored identity tokens for all inter-teammate requests.
    * Support heterogeneous teammate coordination (Claude Code + OpenClaw).
* **Non-Goals:**
    * Replacing existing framework-specific coordination protocols (e.g., Anthropic's internal mailbox). TSE acts as the secure *bus* for these protocols.
    * Providing full OS-level containerization for agents. TSE focuses on the *reasoning and state* isolation.

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Swarm Orchestrator
* **Primary Goal:** Coordinate a team of 5 parallel agents to refactor a legacy codebase without sensitive environment variables leaking between "Auditor" and "Executor" teammates.
* **The Happy Path (Tasks):**
    1. The lead agent (Claude) requests the creation of a new "Team Mission" through MCP Any.
    2. MCP Any initializes the TSE Kernel, generating a hardware-bound Mission Root token.
    3. Each parallel teammate (Auditor, Executor, etc.) is spawned with a unique, cryptographically bound Teammate Identity token.
    4. Teammates communicate via the T2T Encryption Bridge, which validates every mailbox message against the TSE Mission Root.
    5. The Shared Blackboard enforces "Teammate-Locked Shards," ensuring the Auditor cannot see the Executor's high-trust connection strings.
    6. Upon task completion, the TSE Kernel forcefully revokes all teammate-specific capabilities.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        participant Lead as Lead Agent (Claude)
        participant TSE as TSE Isolation Kernel
        participant TM1 as Teammate 1 (Auditor)
        participant BB as Shared Blackboard (Sharded)

        Lead->>TSE: Create Team Mission (MissionRoot: TPM-Signed)
        TSE-->>Lead: Mission Token
        Lead->>TSE: Spawn Teammate (Auditor)
        TSE-->>TM1: Teammate Identity (Bound to MissionRoot)
        TM1->>BB: Write Audit Trace
        BB->>TSE: Validate Teammate Token
        TSE-->>BB: Authorized (Auditor Shard Only)
    ```
* **APIs / Interfaces:**
    * `POST /v1/team/mission`: Initialize a new team mission.
    * `POST /v1/team/teammate`: Provision a new teammate with an isolation-aware identity.
    * `GET /v1/team/validate`: Internal endpoint for middleware to verify teammate-mission lineage.
* **Data Storage/State:**
    * TSE utilizes the "Shared KV Store" (Blackboard) but enforces mandatory sharding based on Teammate Identity.
    * Mission-root metadata is stored in a hardware-protected (TPM/Secure Enclave) state region.

## 5. Alternatives Considered
* **OS-Level Namespacing (Docker):** Rejected as too high-overhead for ephemeral agent sessions and difficult to map to the internal reasoning states of LLMs.
* **Flat Intent Isolation (RAMS):** Evolved into TSE. While RAMS isolated memory, it lacked the cross-teammate coordination awareness required for horizontal "Agent Teams."

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** TSE enforces the "Principle of Least Privilege" at the teammate level. A compromised Auditor cannot impersonate the Lead or Executor teammates.
* **Observability:** Every inter-teammate message and shard access is logged with the TSE Mission ID and Teammate ID, providing a forensic-grade audit trail.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
