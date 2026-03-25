<!--
Copyright (C) 2026 Author(s) of MCP Any
SPDX-License-Identifier: Apache-2.0
-->

# Design Doc: Teammate Sovereignty Enforcer (TSE)
**Status:** Draft | **Created:** 2026-06-18

## 1. Context and Scope
Parallel teammates in agent teams require cryptographically isolated "Sovereignty Shards" to prevent "State Smearing" and logic propagation.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide cryptographically bound isolation for parallel teammates.
    * Enforce "Auth-at-the-Pipe" for inter-agent communication.
    * Bind teammate identities to hardware-attested tokens.
* **Non-Goals:**
    * OS-level virtualization.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Share secure context without exposing local env vars.
* **The Happy Path (Tasks):**
    1. Lead Agent initializes mission with TPM token.
    2. Lead spawns teammates; TSE issues isolated shards.
    3. Comms occur via isolated named pipes managed by the gateway.

## 7. Evolutionary Changelog
- **2026-06-18:** Initial Document Creation.
### Update: 2026-06-18 - Logic-Path Interdiction Integration
**Architecture Adjustment:** Mandatory integration with the Logic-Sovereignty Validator (LSV).
