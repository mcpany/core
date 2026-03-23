# Design Doc: RIA Provider (Recursive Intent Attestation)
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
As agent swarms become deeper and more autonomous, the risk of "intent spoofing" and "hallucination hijacking" increases. Traditional attestation only verifies the identity of the immediate caller. MCP Any needs a way to mathematically prove that every sub-task in a chain is a legitimate descendant of the original, user-authorized mission root.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a cryptographic chain of custody for agentic intent.
    * Enable agents to verify the provenance of received instructions across multiple framework hops.
    * Support auto-revocation of sub-intents if the parent intent is pruned.
* **Non-Goals:**
    * Replacing existing transport-layer security (mTLS/TLSB).
    * Enforcing reasoning-level semantic alignment (handled by AIA Broker).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Swarm Architect
* **Primary Goal:** Ensure that a 3rd-hop subagent cannot perform unauthorized filesystem edits by claiming it was ordered to do so.
* **The Happy Path (Tasks):**
    1. Primary Agent requests a new sub-intent token from the RIA Provider, providing its own root-signed token.
    2. RIA Provider verifies the lineage and issues a cryptographically linked sub-token.
    3. Subagent presents this token to a local tool.
    4. Local tool validates the entire RIA chain against the mission-root via MCP Any before execution.

## 4. Design & Architecture
* **System Flow:**
    `User -> Mission Root (Signed) -> RIA Token A -> RIA Token B (Linked to A) -> Tool Execution (Verified Chain)`
* **APIs / Interfaces:**
    * `POST /v1/ria/token/issue`: Request a new sub-intent token.
    * `POST /v1/ria/token/verify`: Validate a full RIA chain.
* **Data Storage/State:** RIA tokens are ephemeral and stored in a versioned, hardware-attested shard of the Blackboard.

## 5. Alternatives Considered
* **Flat Session Tokens**: Rejected because they do not provide provenance beyond the immediate parent.
* **Full Monologue Signing**: Rejected due to prohibitive latency and token overhead.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** RIA tokens are hardware-bound and expire upon task completion.
* **Observability:** RIA chain events are logged to the tamper-evident audit log for full auditability.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
