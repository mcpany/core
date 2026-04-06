# Design Doc: Optimistic Execution Gate
**Status:** Draft
**Created:** 2026-04-05

## 1. Context and Scope
The "Attestation Tax" (latency) introduced by multi-agent security quorums and hardware signatures can cause "Cognitive Stall" in high-speed swarms. MCP Any needs to implement an optimistic loading pattern where agents can speculatively prepare tool contexts while discovery quorums perform background attestation in parallel.

## 2. Goals & Non-Goals
* **Goals:**
    * Minimize the perceived latency of tool discovery and preparation.
    * Allow speculative reasoning while maintaining strict post-discovery validation.
    * Prevent speculative results from leaking into persistent state until attested.
* **Non-Goals:**
    * Bypassing security quorums entirely.
    * Optimistic execution of write-heavy or high-risk tools (e.g., `rm -rf`).

## 3. Critical User Journey (CUJ)
* **User Persona:** Developer using Gemini CLI Swarms
* **Primary Goal:** Reduce the cold-start delay of subagent delegation without compromising mesh security.
* **The Happy Path (Tasks):**
    1. Agent signals intent to call a tool.
    2. MCP Any identifies the tool and begins background attestation.
    3. MCP Any provides a "Speculative Buffer" to the agent with tool metadata.
    4. Agent begins preparing the tool call context based on the speculative metadata.
    5. Quorum completes attestation.
    6. MCP Any "commits" the tool call, allowing execution to proceed.

## 4. Design & Architecture
* **System Flow:**
    [Agent] --(Intent)--> [Optimistic Gate] --> [Parallel: Attestation & Preparation]
* **APIs / Interfaces:**
    * `GET /v1/speculative/preflight`: Get early tool metadata.
    * `POST /v1/speculative/commit`: Finalize the call after attestation.
* **Data Storage/State:**
    * Speculative results are held in a "Probabilistic Buffer" (ephemeral) until confirmed.

## 5. Alternatives Considered
* **Sync-only Attestation**: Rejected because it creates a performance ceiling for real-time agent coordination.
* **Trust Leases (LFTA)**: Complementary; Optimistic gates handle the first-call latency while LFTA handles subsequent burst calls.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Probabilistic buffers are cryptographically isolated and purged on attestation failure.
* **Observability:** Track "Speculation Success Rate" metrics.

## 7. Evolutionary Changelog
* **2026-04-05:** Initial Document Creation.
