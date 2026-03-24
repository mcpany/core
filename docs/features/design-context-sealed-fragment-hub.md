# Design Doc: Context Sealed-Fragment Hub
**Status:** Draft
**Created:** 2026-05-09

## 1. Context and Scope
With the discovery of the "EchoLeak" vulnerability, where sensitive data is exfiltrated from an agent's context via semantic side-channels, there is an urgent need to move beyond simple memory segmentation. The Context Sealed-Fragment Hub provides cryptographically bound, semantically sealed context shards to ensure data integrity and confidentiality within agent swarms.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement "Active Fragment Sealing" for context shards.
    *   Cryptographically bind shards to specific subagent identities.
    *   Prevent semantic side-channel exfiltration during RAG retrieval.
*   **Non-Goals:**
    *   Encrypted storage for long-term archival (focus is on active reasoning memory).
    *   Replacement for the primary LLM's internal attention mechanism.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Security Architect
*   **Primary Goal:** Ensure that PII retrieved from a local database by a "Data Fetcher" agent cannot be "seen" or "inferred" by a "Public Reporter" agent in the same swarm.
*   **The Happy Path (Tasks):**
    1.  The supervisor agent spawns two subagents with distinct Intent-Sealed tokens.
    2.  The "Data Fetcher" retrieves sensitive data into a "Sealed Fragment."
    3.  The gateway enforces that only the "Data Fetcher" (and authorized supervisor) can mount this fragment.
    4.  The "Public Reporter" attempts to access the fragment and is blocked by the gateway's cryptographic layer.

## 4. Design & Architecture
*   **System Flow:**
    `Agent Request` -> `Identity Provider` -> `Sealing Engine` -> `Blackboard Shard`
*   **APIs / Interfaces:**
    *   `SealFragment(shard_id, agent_token, policy)`
    *   `MountSealedFragment(shard_id, agent_token)`
*   **Data Storage/State:**
    Uses the Shared KV Store (Blackboard) but wraps specific keys in a cryptographic envelope.

## 5. Alternatives Considered
*   **Passive Isolation:** Rejected because it doesn't protect against semantic side-channels or compromised agents with escalated permissions.
*   **Full Context Encryption:** Rejected due to the overhead of decrypting/encrypting for every LLM token generation.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Mandatory origin and identity validation before any sealing operation.
*   **Observability:** Audit logs for every shard mount and access attempt.

## 7. Evolutionary Changelog
*   **2026-05-09:** Initial Document Creation.
