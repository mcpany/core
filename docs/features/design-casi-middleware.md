# Design Doc: Context-Aware Shard Isolation (CASI)
**Status:** Draft
**Created:** 2026-06-07

## 1. Context and Scope
Parallel agent teams (Claude Code Agent Teams) share state via a common "Mailbox." However, without granular isolation, a specialist agent (e.g., a "Security Auditor") might accidentally leak its private session data into the shared mailbox, where a generalist "Task Runner" could ingest it.

CASI provides semantic boundaries for teammate shards, ensuring that only relevant, authorized fragments are synchronized.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement semantic filtering of state fragments crossing teammate boundaries.
    * Prevent "Shard Pollution" (exfiltration of private env vars or reasoning traces).
    * Enforce task-bound isolation for mailbox shards.
* **Non-Goals:**
    * Encryption of the local filesystem itself.
    * Real-time monitoring of agent token costs.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Share secure context between 3 agents without exposing local env vars or private SSH keys.
* **The Happy Path (Tasks):**
    1. Teammate A writes a state fragment to its mailbox shard.
    2. Teammate B requests a sync of the shared mailbox.
    3. CASI intercepts the sync request.
    4. CASI scans the fragment for "Sovereign Sensitive" data (keys, PII).
    5. Teammate B receives only the authorized, sanitized state fragment.

## 4. Design & Architecture
* **System Flow:**
    `Teammate A -> Local Shard -> CASI Filter -> Shared Mailbox -> Teammate B`
* **APIs / Interfaces:**
    - `Middleware.OnShardSync(ctx, fragment)`: Semantic hook for filtering fragments.
* **Data Storage/State:**
    Shard policies are stored in the Mesh-Resident Policy Synthesizer.

## 5. Alternatives Considered
* **Binary Access Control**: Rejected because it cannot perform the *semantic* analysis required to detect PII or keys hidden in natural language reasoning.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** CASI enforces "Least Privilege" for state sharing.
* **Observability:** Filtered/Redacted fragments are flagged in the FAMI (Fragment-Aware Mailbox Isolation) Auditor.

## 7. Evolutionary Changelog
* **2026-06-07:** Initial Document Creation.
