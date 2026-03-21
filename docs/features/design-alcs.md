# Design Doc: Attention-Locked Context Sharding (ALCS)

**Status:** Draft | **Created:** 2026-06-15

## 1. Context and Scope

In deep agentic swarms, mission-critical intent fragments are often evicted from
the context window due to "Attention Entropy" caused by high-volume subagent
outputs. ALCS provides a mechanism to "pin" specific context shards to the LLM's
attention layers using hardware-bound attestation.

## 2. Goals & Non-Goals

- **Goals:**
  - Hardware-attested pinning of intent-critical context shards.
  - Real-time monitoring of attention occupancy across sub-graphs.
- **Non-Goals:**
  - Increasing raw context window size.
  - Long-term agent memory management.

## 3. Critical User Journey (CUJ)

- **Primary Goal:** Ensure "Zero-Trust" directives remain in the LLM's focus
  during recursive execution.
- **The Happy Path:**
  1. Mark shard as "Mission Root".
  2. Generate hardware-bound header.
  3. Inject into LLM context.
  4. Monitor attention weights.

## 4. Design & Architecture

- **System Flow:** Shard -> Hardware Verifier -> Attention Header -> LLM
  Context.

## 7. Evolutionary Changelog

- **2026-06-15:** Initial Document Creation.
