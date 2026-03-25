# Design Doc: Teammate Sovereignty Enforcer (TSE)
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
With the transition to "Agent Teams," parallel teammates require cryptographically isolated "Sovereignty Shards" to prevent "State Smearing" and logic bombs.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide cryptographically bound isolation for parallel teammates.
    * Enforce "Auth-at-the-Pipe" for inter-agent communication.
    * Bind identities to hardware-attested (TPM) tokens.
* **Non-Goals:**
    * OS-level virtualization for the LLM itself.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Share secure context between 3 agents without exposing local env vars.
* **The Happy Path (Tasks):**
    1. Lead Agent initializes a TSE mission with a TPM root token.
    2. Lead spawns teammates; TSE issues isolated identity shards.
    3. TSE mandates communication via isolated named pipes managed by the gateway.
    4. TSE Kernel enforces "Teammate-Locked Shards."
    5. Upon task completion, TSE Kernel revokes all capabilities.

## 4. Design & Architecture
* **System Flow:** Isolation Kernel manages the lifecycle of shards; Auth-at-the-Pipe validates mission-root tokens.
* **APIs:** `POST /v1/team/mission`, `POST /v1/team/teammate`.

## 5. Alternatives Considered
* **OS-Level Namespacing**: Rejected as too high-overhead.

## 6. Cross-Cutting Concerns
* **Security**: Zero Trust at the teammate level.
* **Observability**: Forensic-grade audit trails per mission.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
### Update: 2026-06-18 - Logic-Path Interdiction Integration
**Architecture Adjustment:** Mandatory integration with the Logic-Sovereignty Validator (LSV).
