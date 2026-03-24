# Design Doc: Relational PoI Validator
**Status:** Draft
**Created:** 2026-03-24

## 1. Context and Scope
The "Identity-Only" security model is failing against "Context-Mirroring" (CVE-2026-34015), where subagents are coerced into actions by echoing parent state. Relational PoI Validator extends Proof-of-Intent to verify the entire "Intent Chain," ensuring subagents remain anchored to the parent's verified goal across multiple delegation hops.

## 2. Goals & Non-Goals
* **Goals:**
    * Verify the cryptographically signed "Intent Chain" for every tool call.
    * Implement relational scoping to dynamically narrow permissions based on intent lineage.
    * Detect and block "Context-Mirroring" attempts.
* **Non-Goals:**
    * Replace the base PoI Validator (this is an extension).
    * Enforce resource-level quotas (handled by Quota Monitor).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Security Architect
* **Primary Goal:** Prevent a specialized subagent from being tricked into using a "System Shell" tool if its parent intent was only "Read Only Analysis."
* **The Happy Path (Tasks):**
    1. Parent agent issues a "Relational Intent Token" to a subagent.
    2. Subagent attempts to delegate a task to a third agent, appending its own intent fragment.
    3. The resulting "Intent Chain" is passed with a tool call.
    4. MCP Any Relational PoI Validator intercepts the call and verifies all signatures in the chain.
    5. Validator checks if the tool call aligns with the most restrictive intent in the lineage.
    6. Tool execution is permitted or blocked.

## 4. Design & Architecture
* **System Flow:**
    `[Subagent] -> [Intent Chain] -> [Relational Validator] -> [Permission Guard] -> [MCP Server]`
* **APIs / Interfaces:**
    * `X-UACO-Intent-Chain` header containing a sequence of signed JWS tokens.
    * `validateIntentChain(tokens[])` internal API.
* **Data Storage/State:**
    * Ephemeral cache for verified chain segments to optimize multi-hop performance.

## 5. Alternatives Considered
* **Flat Intent Tokens:** Rejected as they don't provide sufficient lineage for deep swarms.
* **Centralized Intent Registry:** Rejected due to scaling and privacy concerns for decentralized swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Use hardware-bound keys (TPM/Secure Enclave) where available for signing intents.
* **Observability:** Track chain depth and verification latency metrics.

## 7. Evolutionary Changelog
* **2026-03-24:** Initial Document Creation.
