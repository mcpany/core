# Design Doc: Teammate Sovereignty Enforcer (TSE)
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
With the transition to "Agent Teams," parallel teammates require cryptographically isolated "Sovereignty Shards" to prevent "State Smearing" and "Logic Bombs."

## 2. Goals & Non-Goals
* **Goals:**
    * Provide cryptographically bound isolation.
    * Enforce "Auth-at-the-Pipe" for teammate comms.
    * Bind identities to TPM mission-root tokens.
* **Non-Goals:**
    * OS-level virtualization.

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Orchestrator
* **Primary Goal:** Coordinate parallel teammates without shared context leaks.
* **The Happy Path (Tasks):**
    1. Initialize mission with TPM token.
    2. Spawn teammates with isolated shards.
    3. Mandate comms via isolated named pipes.
    4. Enforce shard locking.

## 4. Design & Architecture
* **Isolation Kernel**: Manages Sovereignty Shards.
* **Logic-Aware Shard Guards**: Real-time reasoning filters.

## 5. Alternatives Considered
* **OS Namespacing**: Too high overhead.

## 6. Cross-Cutting Concerns
* **Security**: Zero Trust at the teammate level.
* **Observability**: Forensic-grade audit trails.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
### Update: 2026-06-18 - Logic-Path Interdiction Integration
**Architecture Adjustment:** Mandatory integration with LSV.
