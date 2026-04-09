# Design Doc: Shard-Level Redaction (SLR) Hub
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agent swarms scale horizontally, the sharing of context shards (e.g., mailboxes, scratchpads) becomes a primary source of "Intent Leakage" and PII exposure. Specialist subagents often have access to sensitive fragments (PII, internal monologues, system constraints) that should not be visible to other teammates or parent reasoning loops.

The Shard-Level Redaction (SLR) Hub provides a centralized, hardware-attested sanitization layer that automatically redacts unauthorized intents and sensitive data from shards before they are committed to the shared teammate mesh.

## 2. Goals & Non-Goals
* **Goals:**
    * Perform real-time, semantic redaction of PII and private monologues from context shards.
    * Provide hardware-attested proofs of redaction integrity.
    * Neutralize "Monologue Smearing" across heterogeneous swarms.
    * Support pluggable redaction policies (e.g., regex, LLM-based, or heuristic).
* **Non-Goals:**
    * Managing the underlying shard storage (handled by UMMB).
    * Providing general-purpose PII scrubbing for non-agentic data.
    * Directly modifying the LLM's attention weights (handled by HAAL).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Orchestrator
* **Primary Goal:** Share a task-relevant context shard with a specialist agent without exposing the parent's private reasoning monologue.
* **The Happy Path (Tasks):**
    1. Parent agent prepares a context shard for a delegated task.
    2. Shard is routed through the SLR Hub before being shared.
    3. SLR Hub performs semantic analysis, identifying fragments marked as "Private" or "Sensitive".
    4. SLR Hub redacts the identified fragments, replacing them with mission-bound placeholders.
    5. SLR Hub issues a hardware-attested "Redaction Receipt" linked to the shard.
    6. Specialist agent receives the sanitized shard and the receipt, ensuring it only operates on authorized state.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Raw Context Shard] --> B[SLR Hub]
        C[Redaction Policy] --> B
        B --> D[Semantic Scanner]
        D --> E[Redacted Shard]
        B --> F[TPM Attestor]
        F --> G[Redaction Receipt]
        E --> H[Teammate Mailbox]
        G --> H
    ```
* **APIs / Interfaces:**
    * `slr.RedactShard(shard Shard, identity Identity) -> (SanitizedShard, Receipt)`: Redacts sensitive content.
    * `slr.VerifyReceipt(receipt Receipt, shard SanitizedShard) -> bool`: Verifies the redaction attestation.
* **Data Storage/State:**
    * **Redaction Policy Registry:** A hardware-locked store of authorized redaction patterns and rules.

## 5. Alternatives Considered
* **Client-Side Redaction:** Rejected as specialist subagents cannot be trusted to redact their own state. SLR provides a mandatory infrastructure-level gate.
* **Simple Regex Scrubbing:** Insufficient for complex reasoning traces. SLR requires semantic analysis to detect "Intent Leakage."

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Redaction scanners are isolated in hardware enclaves to prevent leakage during the sanitization process itself.
* **Observability:** Integrated with the "RSR Redaction Auditor" in the UI for visual verification of redacted fragments.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
