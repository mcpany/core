# Design Doc: Shard-Level Ephemeral Secret (SLES) Broker
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
AI agents frequently require temporary API keys or credentials to perform specific tasks. Current systems often leak these into global environment variables, leading to "Credential Sprawl" and high exfiltration risks. As swarms become more granular and sharded (e.g., OpenClaw's context sharding), security must move from the session level to the fragment level.

The SLES Broker provides hardware-attested, task-bound credentials that are cryptographically tied to specific context shards and automatically purged when no longer needed.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue ephemeral, task-bound credentials cryptographically bound to specific context shards.
    * Automate the purging of secrets upon shard unmounting or mission completion.
    * Provide a hardware-attested "Secret Mint" for specialist agents.
    * Neutralize "Credential Sprawl" in deep agent chains.
* **Non-Goals:**
    * Long-term persistent secret storage (use a standard Vault for that).
    * Managing non-agent secrets or system-level environment variables.

## 3. Critical User Journey (CUJ)
* **User Persona:** Specialist subagent developer.
* **Primary Goal:** Securely use a temporary S3 key for a specific data-retrieval shard without exposing it to the parent agent.
* **The Happy Path (Tasks):**
    1. Parent agent delegates a data-retrieval task to a specialist subagent.
    2. SLES Broker issues a task-bound S3 credential, cryptographically linked to the `data_retrieval_v1` context shard.
    3. The subagent executes the tool using the ephemeral secret.
    4. Upon task completion, the subagent unmounts the shard.
    5. SLES Broker detects the unmount signal and wipes the ephemeral secret from kernel-bound memory.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Mission Root] --> B[SLES Broker]
        B --> C{Context Shard?}
        C -->|Yes| D[Issue Shard-Bound Secret]
        C -->|No| E[Deny Request]
        D --> F[Specialist Agent]
        F --> G[Task Completion]
        G --> H[Purge Signal]
        H --> I[SLES Memory Wipe]
    ```
* **APIs / Interfaces:**
    * `sles.MintSecret(shardID, metadata) -> SecretToken`: Generates a shard-bound credential.
    * `sles.GetSecret(secretToken) -> Plaintext`: Authoritative retrieval for sandboxed tools.
    * `sles.RevokeByShard(shardID) -> Status`: Immediate revocation and memory wipe.
* **Data Storage/State:**
    * **Kernel-Bound Secret Buffer:** Non-swappable memory regions for plaintext secrets.
    * **Shard-to-Secret Mapping:** In-memory, hardware-attested registry.

## 5. Alternatives Considered
* **Environment Variable Injection:** Rejected because env vars are globally accessible to child processes and easily exfiltrated via shell commands.
* **Standard Vault Proxying:** Rejected because standard vaults lack the "Context Shard Awareness" needed for automatic lifecycle management in swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Secrets are never persisted to disk. Hardware-locked attestation ensures only the authorized shard owner can retrieve the secret.
* **Observability:** Integrated with the "Blackboard Lineage Inspector" to track secret usage within the cryptographic audit trail.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
