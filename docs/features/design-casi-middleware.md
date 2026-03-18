# Design Doc: Context-Aware Shard Isolation (CASI) Middleware
**Status:** Draft
**Created:** 2026-06-07

## 1. Context and Scope
Horizontal coordination in AI agent teams (e.g., Claude Code Agent Teams) relies on shared mailboxes and context shards. However, as teams scale, "Shard Pollution" becomes a critical failure point. A teammate working on a "Frontend" task may accidentally ingest or modify state fragments from a concurrent "Backend" shard, leading to "Reasoning Drift" and potential security leaks between logic boundaries.

The CASI Middleware solves this by enforcing semantic isolation at the fragment level. Teammates are restricted to "seeing" and "writing" only those state fragments that are semantically aligned with their active task context.

## 2. Goals & Non-Goals
* **Goals:**
    * Perform real-time semantic analysis of context shards during teammate synchronization.
    * Restrict teammate access to state fragments based on the active task UUID and mission-root intent.
    * Prevent cross-shard state pollution in horizontal swarms.
    * Maintain a mission-root "Semantic Anchor" for all sharded fragments.
* **Non-Goals:**
    * Replacing the underlying Shared KV Store (Blackboard).
    * Managing global coordination locks (handled by the ASLM).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator.
* **Primary Goal:** Coordinate two parallel teammates (one for Documentation, one for Code) without the Documentation agent accessing sensitive database schema fragments from the Code agent's shard.
* **The Happy Path (Tasks):**
    1. The Orchestrator assigns Task A (Docs) to Agent 1 and Task B (Code) to Agent 2.
    2. Agent 2 writes a schema fragment to the "Code" shard.
    3. Agent 1 requests a state sync for the "Docs" shard.
    4. The CASI Middleware intercepts the request, performs a semantic check, and filters out the schema fragment because it doesn't align with Task A's context.
    5. Agent 1 receives only the documentation-relevant fragments.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        A[Teammate Agent] --> B[CASI Middleware]
        B --> C{Semantic Check}
        C -- Match --> D[Active Task Shard]
        C -- Mismatch --> E[Filter/Block]
        D --> F[Shared Mailbox/Blackboard]
    ```
* **APIs / Interfaces:**
    * `GET /mailbox/sync`: Upgraded with `task_context_id` header; returns filtered shards.
    * `PUT /mailbox/shard`: Upgraded with `intent_fragment_tag` for CASI-aware indexing.
* **Data Storage/State:**
    * Shards are tagged with "Context Signatures" (hashes of the task intent).
    * The middleware utilizes the `Intent-Sealed Blackboard Shards` for physical isolation.

## 5. Alternatives Considered
* **Namespace-based Isolation:** Rejected because it is too rigid for dynamic swarms where agents may need to share specific subsets of state.
* **Manual Gating:** Rejected due to high latency and human-in-the-loop bottlenecks in machine-speed meshes.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** CASI ensures that even if an agent's transport is compromised, it cannot exfiltrate state from sibling shards without a matching mission token.
* **Observability:** Blocked sync attempts are logged as "Semantic Violations" in the CSAD Hub.

## 7. Evolutionary Changelog
* **2026-06-07:** Initial Document Creation.
