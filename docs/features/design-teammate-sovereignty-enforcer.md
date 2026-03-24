<!--
Copyright (C) 2026 Author(s) of MCP Any
SPDX-License-Identifier: Apache-2.0
-->

# Design Doc: Teammate Sovereignty Enforcer (TSE)
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
With the transition to "Agent Teams," multiple agents now work in parallel on the same mission. Today's market sync revealed that these parallel teammates often share too much context or can be coerced into "State Smearing."

The Teammate Sovereignty Enforcer (TSE) provides a kernel-level isolation layer within MCP Any. It ensures that every teammate in a swarm operates within a cryptographically isolated "Sovereignty Shard."

## 2. Goals & Non-Goals
* **Goals:**
    * Provide cryptographically bound isolation for parallel teammates.
    * Enforce "Auth-at-the-Pipe" for inter-agent communication.
    * Bind teammate identities to hardware-attested (TPM) tokens.
    * Integrate with the Logic-Sovereignty Validator (LSV).
* **Non-Goals:**
    * Replacing the primary mission-root security policy.
    * Providing OS-level virtualization for the LLM itself.

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Orchestrator (Lead Agent)
* **Primary Goal:** Coordinate an "Executor" and a "Security Auditor" teammate without shared context leaks.
* **The Happy Path (Tasks):**
    1. Lead initializes a TSE mission with a TPM root token.
    2. Lead spawns teammates; TSE issues isolated identity shards.
    3. TSE mandates communication via isolated named pipes.
    4. TSE Kernel enforces "Teammate-Locked Shards."
    5. Upon task completion, TSE revokes all teammate-specific capabilities.

## 4. Design & Architecture
* **System Flow:**
    * **Isolation Kernel**: Manages the lifecycle of "Sovereignty Shards."
    * **Auth-at-the-Pipe**: Inter-teammate transport that validates mission-root tokens.
    * **Logic-Aware Shard Guards**: Real-time filters that interface with the LSV.

## 5. Alternatives Considered
* **OS-Level Namespacing (Docker)**: Rejected as too high-overhead.
* **Flat Intent Isolation (RAMS)**: Evolved into TSE; added cross-teammate coordination awareness.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: TSE enforces the "Principle of Least Privilege" at the teammate level.
* **Observability**: Every inter-teammate message is logged with TSE Mission ID and Teammate ID.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
### Update: 2026-06-18 - Logic-Path Interdiction Integration
**Architecture Adjustment:** Mandatory integration with the Logic-Sovereignty Validator (LSV).
