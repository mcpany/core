# Design Doc: Recursive Intent Attestation (RIA) Provider
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
As AI agent swarms grow in depth and complexity, the "Intent-Grafting" vulnerability (CVE-2026-65002) has emerged as a critical threat. This exploit allows a subagent to append unauthorized goals to a verified mission root, effectively bypassing parent-imposed sandbox restrictions.

MCP Any needs a mechanism to ensure that every sub-instruction issued by any agent in the mesh is mathematically and cryptographically derived from the original user intent. The RIA Provider solves this by implementing a recursive proof system that validates the entire chain of custody for an intent before any tool execution occurs.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a recursive cryptographic attestation root for all mission intents.
    * Mathematically derive sub-agent "Intent Tokens" from the verified Mission Root.
    * Provide sub-millisecond validation of intent lineage during tool calls.
    * Neutralize the "Intent-Grafting" exploit pattern.
* **Non-Goals:**
    * This system will NOT perform semantic reasoning (handled by the AID Hub).
    * It will NOT manage transport-level encryption (handled by TLSB/IBET).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Architect
* **Primary Goal:** Prevent a specialized "Code Search" subagent from being coerced into "Exfiltrating Secrets" via a grafted sub-intent.
* **The Happy Path (Tasks):**
    1. The primary orchestrator issues a "Mission Root" token signed by MCP Any.
    2. The Code Search subagent receives a derived "Sub-Intent" token.
    3. The subagent attempts to call a `git-clone` tool.
    4. The RIA Provider validates the token against the Mission Root's derivation path.
    5. Validation succeeds; tool execution proceeds.
    6. (Attack Path): A rogue subagent attempts to graft a `curl-upload` sub-intent.
    7. The RIA Provider detects a derivation mismatch and halts execution immediately.

## 4. Design & Architecture
* **System Flow:**
    `[Mission Root (User)] -> [MCP Any Master Key] -> [Derived Intent Token (v1)] -> [Sub-Intent Token (v1.1)]`
    Lineage is verified using HMAC-based Extract-and-Expand (HKDF) or similar one-way derivation chains.
* **APIs / Interfaces:**
    * `DeriveIntent(parent_token, sub_goal) -> intent_token`
    * `VerifyLineage(intent_token) -> bool`
* **Data Storage/State:**
    Intent lineage paths are stored in an ephemeral, hardware-bound cache (TPM-backed).

## 5. Alternatives Considered
* **Flat Token Scoping:** Rejected because it cannot handle deep, dynamic swarms where sub-goals are generated at runtime.
* **Full Semantic Deconstruction (AID):** Too slow for per-hop validation; RIA provides the fast-path cryptographic "permission to thought."

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Lineage proofs are hardware-bound and expire with the session.
* **Observability:** Every derivation event is logged to the CoC Lineage Tracker for forensic auditing.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
