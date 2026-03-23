# Design Doc: Attention-Locked Context Sharding (ALCS)
**Status:** Draft | **Created:** 2026-06-15

## 1. Context and Scope
In deep agentic swarms, mission-critical intent fragments are often evicted from
the context window due to "Attention Entropy" caused by high-volume subagent
outputs. ALCS provides a mechanism to "pin" specific context shards to the LLM's
attention layers using hardware-bound attestation.

## 2. Goals & Non-Goals
- **Goals:**
    - Cryptographically priority-tiering of context shards.
    - Real-time monitoring of attention occupancy.
- **Non-Goals:**
    - Increasing raw context window size.
    - Managing long-term agent memory.

## 3. Critical User Journey (CUJ)
- **User Persona:** Local LLM Swarm Orchestrator
- **Primary Goal:** Share secure context between 3 agents without exposing local env vars
- **The Happy Path (Tasks):**
    1. Define security shards containing environment constraints.
    2. Attach ALCS attestation to the shards.
    3. Initialize subagents with the locked context.
    4. Subagents perform recursive tasks while the security shard remains pinned
in the attention layer.

## 4. Design & Architecture
- **System Flow:**
    `[Agent Request] -> [ALCS Filter] -> [Hardware Attestation Service] ->
[Pinned LLM Context] -> [Response]`
- **APIs / Interfaces:**
    - `POST /v1/context/pin`: Pin a specific context fragment.
    - `GET /v1/context/attention-status`: Retrieve current attention layer occupancy metrics.
- **Data Storage/State:**
    - State is managed via isolated memory segments within the gateway proxy, ensuring that even if a subagent generates high-token-count output, the pinned segment is protected from eviction policies.

## 5. Alternatives Considered
- **Naive Token Pushing:** Simply repeating the critical context in every turn. Rejected due to excessive token consumption and potential "lost in the middle" effects.
- **Vector Retrieval-only:** Retrieving context only when needed. Rejected as it introduces latency and doesn't guarantee the shard is present during the reasoning step.

## 6. Cross-Cutting Concerns
- **Security (Zero Trust):** Pinned context shards are cryptographically signed. Only agents with verified identities can request pins, preventing prompt injection from hijacking the attention lock.
- **Observability:** ALCS exposes "Attention Eviction Logs" to identify when shards are nearing the attention threshold.

## 7. Evolutionary Changelog
- **2026-06-15:** Initial Document Creation.
