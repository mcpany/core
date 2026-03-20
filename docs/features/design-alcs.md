# Design Doc: Attention-Locked Context Sharding (ALCS)

**Status:** Draft | In Review | Approved **Created:** [2026-06-15]

## 1. Context and Scope

As agent swarms grow in complexity, "Attention Entropy" leads to the eviction of
critical intent fragments from the context window. ALCS provides a
hardware-bound guarantee that mission-critical directives remain "pinned" to the
LLM's active reasoning tiers, preventing cognitive stall or drift.

## 2. Goals & Non-Goals

- **Goals:**
  - Hardware-attested pinning of intent-critical context shards.
  - Real-time monitoring of attention occupancy across sub-graphs.
  - Isolation of adversarial "noise" intended to trigger context eviction.
- **Non-Goals:**
  - General-purpose vector retrieval (RAG).
  - Long-term memory persistence beyond the active reasoning session.

## 3. Critical User Journey (CUJ)

- **User Persona:** Multi-Swarm Orchestrator
- **Primary Goal:** Ensure that "Zero-Trust" directives are never evicted from
  the context window during deep recursive tool execution.
- **The Happy Path (Tasks):**
  1. Orchestrator marks a security shard as "Locked".
  2. ALCS generates a hardware-bound attention header for the shard.
  3. The shard is injected into the LLM context with a "Priority 0" weight.
  4. ALCS monitors the LLM's attention weights to ensure the shard remains above
     the eviction threshold.

## 4. Design & Architecture

- **System Flow:** Shard -> Hardware Verifier -> Attention Header -> LLM
  Context.
- **APIs / Interfaces:** `POST /v1/attention/lock`, `GET /v1/attention/status`.
- **Data Storage/State:** Volatile attention state managed via hardware-bound
  secure enclaves.

## 5. Alternatives Considered

- **Vector Eviction Management:** Rejected due to high latency and lack of
  hardware-bound security guarantees.
- **Context Compression:** Rejected as it tends to lose the "Zero-Trust"
  metadata required for security enforcement.

## 6. Cross-Cutting Concerns

- **Security (Zero Trust):** Multi-factor attestation for all "Locked" shards.
- **Observability:** Integration with the ALCS Dashboard for real-time entropy
  monitoring.

## 7. Evolutionary Changelog

- **[2026-06-15]:** Initial Document Creation.
