# Design Doc: Reasoning-Aware Redaction (RAR) Engine
**Status:** Draft
**Created:** 2026-07-19

## 1. Context and Scope
As AI agents increasingly utilize persistent state (e.g., SSP-compliant tool storage) and shared workspaces (e.g., `.scratchpad`), the risk of "Context-Stitching" has become critical. Adversaries or compromised subagents can re-compose sensitive mission-root intents by analyzing fragmented state left behind by specialized tools.

The RAR Engine is designed to solve this by providing a unified governance layer for cognitive privacy. It ensures that even if a skill's persistent state is compromised, the high-level reasoning path and mission-critical intents remain invisible to unauthorized entities.

## 2. Goals & Non-Goals
* **Goals:**
    * Automatically identify mission-critical intent fragments in outgoing tool state writes.
    * Perform real-time semantic redaction of sensitive reasoning monologues before they are persisted.
    * Provide a hardware-bound "Salt" to state fragments to prevent cross-fragment correlation.
* **Non-Goals:**
    * Encrypting all tool output (RAR is focused on *redacting reasoning*, not general encryption).
    * Modifying the internal weights of the LLM.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Share state between 5 specialized agents without exposing the parent's secret mission constraints.
* **The Happy Path (Tasks):**
    1. Parent agent initiates a mission with hardware-attested constraints.
    2. Parent delegates a task to a "Database Specialist" tool.
    3. The tool attempts to write its execution trace to its persistent state (SSP).
    4. The RAR Engine intercepts the write request.
    5. RAR performs semantic analysis, identifying parent-level constraints in the trace.
    6. RAR redacts the identified fragments and applies a "Cognitive Salt" to the remaining metadata.
    7. The sanitized fragment is written to the SSP shard.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent/Tool] -->|Write Request| B(RAR Proxy)
        B --> C{Semantic Analysis}
        C -->|Match Found| D[Redaction Engine]
        C -->|No Match| E[Salt Injector]
        D --> E
        E --> F[Persistent State Shard]
    ```
* **APIs / Interfaces:**
    * `RedactIntent(fragment string, missionID string) -> sanitized string`
    * `AttestStateLineage(shardID string) -> provenance proof`
* **Data Storage/State:**
    * Redaction policies are stored in a hardware-locked (TPM) secure enclave.

## 5. Alternatives Considered
* **Full Shard Encryption**: Rejected because it prevents legitimate subagents from performing cross-referenced reasoning on authorized data fragments.
* **Static Keyword Filtering**: Rejected because modern agents use varied natural language, requiring semantic-level detection.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** RAR relies on the hardware-attested mission root to define the "Redaction Baseline."
* **Observability:** Redaction events are logged with high-entropy hashes (not the raw data) to monitor for "Reasoning Mirroring" attacks.

## 7. Evolutionary Changelog
* **2026-07-19:** Initial Document Creation.

### Update: 2026-07-25 - Echo-Residue Sanitizer (ERS) Integration
**Context:** Today's market sync revealed CVE-2026-99102 (stylometric traces in kernel buffers).
**Architecture Adjustment:** * Integrated kernel-level memory scrubbing for all sharded buffers upon fragment deletion to neutralize "ghost traces."
**Security Impact:** Mitigates cross-mission stylometric mimicry by ensuring absolute erasure of sharded state fragments.
