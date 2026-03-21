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
  - Integration with HAAL (Hardware-Attested Attention Locking).
- **Non-Goals:**
  - Increasing raw context window size.
  - Managing long-term agent memory.

## 3. Critical User Journey (CUJ)

- **User Persona:** Swarm Orchestrator
- **Primary Goal:** Ensure security shards remain in focus during recursive
  execution.
- **The Happy Path (Tasks):**
  1. Mark shard as "Mission Root".
  2. Generate hardware-bound header.
  3. Inject into LLM context.
  4. Monitor attention weights.

## 4. Design & Architecture

- **System Flow:** Shard -> Hardware Verifier -> Attention Header -> LLM
  Context.
- **APIs / Interfaces:** `POST /v1/attention/lock`, `GET /v1/attention/status`.

## 5. Alternatives Considered

- Simple top-k retrieval (rejected due to lack of security).

## 6. Cross-Cutting Concerns

- **Security:** Hardware attestation.
- **Observability:** Real-time attention dashboard.

## 7. Evolutionary Changelog

- **2026-06-15:** Initial Document Creation.
