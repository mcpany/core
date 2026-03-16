# Design Doc: Context Sealed-Fragment Hub
**Status:** Draft
**Created:** 2026-05-09

## 1. Context and Scope
With the discovery of the "EchoLeak" vulnerability, standard context segmentation is no longer sufficient. Attackers can exploit semantic side-channels to exfiltrate sensitive data from RAG-based systems. The Context Sealed-Fragment Hub (CSFH) aims to move beyond passive isolation by implementing "Active Fragment Sealing," where every context shard is cryptographically bound to a specific mission intent and session token.

## 2. Goals & Non-Goals
* **Goals:**
    * Cryptographically seal context fragments to prevent semantic side-channel exfiltration.
    * Enforce intent-bound retrieval, ensuring shards are only accessible if they align with the current verified sub-mission.
    * Provide a standardized interface for agents to "mount" and "unmount" sealed shards.
* **Non-Goals:**
    * Encryption of data at rest (handled by underlying storage).
    * Providing a new RAG engine (CSFH acts as a middleware for existing engines).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Swarm Orchestrator
* **Primary Goal:** Share sensitive PII between two specialized agents without risking exfiltration via "EchoLeak" side-channels.
* **The Happy Path (Tasks):**
    1. Parent agent creates a "Sealed Shard" containing the PII.
    2. Parent agent generates a "Shared Intent Token" (SIT) for the sub-mission.
    3. Parent agent delegates the task to Subagent A, providing the SIT.
    4. Subagent A requests access to the shard from the CSFH, presenting the SIT.
    5. CSFH verifies the SIT against the shard's seal and grants time-bound, read-only access.
    6. Upon task completion, the SIT is revoked and the shard is "unmounted."

## 4. Design & Architecture
* **System Flow:**
    - `Agent` -> `Request(IntentToken)` -> `CSFH`
    - `CSFH` -> `Verify(IntentToken, ShardSeal)` -> `Result(DecryptedView)`
* **APIs / Interfaces:**
    - `CreateShard(Data, IntentScope) -> ShardID`
    - `SealShard(ShardID, MissionRootKey) -> ShardSeal`
    - `AccessShard(ShardID, IntentToken) -> Data`
* **Data Storage/State:**
    - Metadata and seals are stored in a hardened SQLite instance.
    - Actual fragment data can be stored in the Shared KV Store (Blackboard).

## 5. Alternatives Considered
* **Plain Segmentation:** Rejected because it doesn't protect against semantic side-channel attacks.
* **Total Session Encryption:** Rejected due to high overhead and inability to share granular fragments between subagents.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Shards are sealed with keys derived from the Mission Root, ensuring only agents within the verified lineage can request access.
* **Observability:** Every access request is logged with its associated Intent Token and agent identity.

## 7. Evolutionary Changelog
* **2026-05-09:** Initial Document Creation.
