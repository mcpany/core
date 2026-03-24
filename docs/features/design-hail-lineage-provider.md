<!-- markdownlint-disable -->
# Design Doc: HAIL (Hardware-Attested Intent Lineage)
**Status:** Draft
**Created:** 2026-06-19

## 1. Context and Scope
As agent swarms evolve, specialized subagents are increasingly prone to "Logic Grafting" and identity spoofing. HAIL provides a standardized protocol for cryptographically signing every sub-instruction and linking it back to a hardware-attested mission-root intent.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide hardware-bound (TPM) signatures for all reasoning fragments.
    * Establish a non-repudiable "Chain of Command" for sub-delegations.
* **Non-Goals:**
    * Real-time content sanitization.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Verify that a subagent tool call was authorized by the parent mission-root.
* **The Happy Path (Tasks):**
    1. Parent agent initializes a session with a TPM-signed mission-root token.
    2. Parent spawns a subagent and issues a "Lineage Fragment" token.

## 4. Design & Architecture
* **System Flow:** `[Parent] -> [Lineage Token] -> [Subagent] -> [HAIL Gateway] -> [Tool]`

## 7. Evolutionary Changelog
* **2026-06-19:** Initial Document Creation.
