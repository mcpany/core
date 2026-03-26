# Design Doc: Teammate Sovereignty Enforcer (TSE)
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
Parallel teammates in agent teams require cryptographically isolated "Sovereignty Shards" to prevent "State Smearing" and logic bombs.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide cryptographically bound isolation for parallel teammates.
    * Enforce "Auth-at-the-Pipe" for inter-agent communication.
    * Bind identities to TPM tokens.
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
* **System Flow:**
    * Isolation Kernel manages shard lifecycles.
    * Auth-at-the-Pipe validates mission-root tokens.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
### Update: 2026-06-18 - Logic-Path Interdiction Integration
**Architecture Adjustment:** Mandatory integration with the Logic-Sovereignty Validator (LSV).
