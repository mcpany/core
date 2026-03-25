# Design Doc: RIA Provider (Recursive Intent Attestation)
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
As agent swarms become deeper (multi-hop delegation), the risk of "Intent-Grafting" increases, where a compromised subagent injects unauthorized instructions into a valid chain. MCP Any needs a way to mathematically prove that every sub-intent in a multi-hop chain is a direct, authorized descendant of the user's root mission.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a cryptographic hash-chaining mechanism for intent lineage.
    * Provide a verification API for agents to validate peer lineage.
    * Support hardware-bound (TPM) root signing.
* **Non-Goals:**
    * Encrypting the actual intent content (handled by T2T).
    * Managing LLM context windows (handled by CWP).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Enterprise Architect
* **Primary Goal:** Ensure that a 3rd-level subagent cannot call a destructive "Delete DB" tool without a verified path from the user.
* **The Happy Path (Tasks):**
    1. Orchestrator creates a TPM-signed Root Intent.
    2. Subagent A requests a derived sub-intent token from RIA Provider.
    3. Subagent B receives the token and verifies the chain back to the root before execution.

## 4. Design & Architecture
* **System Flow:**
    `[Mission Root] -> [RIA Token (Depth 0)] -> [Sub-Task A] -> [RIA Token (Depth 1)]`
* **APIs / Interfaces:**
    * `POST /ria/issue`: Generate a derived token.
    * `POST /ria/verify`: Validate a multi-hop token chain.
* **Data Storage/State:**
    Tokens are stateless but rely on the hardware-bound Root Key for verification.

## 5. Alternatives Considered
* **Flat JWTs:** Rejected because they don't capture the recursive "derivation" required for multi-hop swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Tokens include monotonic nonces to prevent replay attacks.
* **Observability:** Every RIA derivation is logged in the CoC Lineage Tracker.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
