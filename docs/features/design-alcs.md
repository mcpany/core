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

- Ensure security shards remain in focus during recursive execution.

## 4. Evolutionary Changelog

- **2026-06-15:** Initial Document Creation.
