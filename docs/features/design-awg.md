# Design Doc: Attention-Window Guard (AWG)
**Status:** Draft
**Created:** 2026-07-09

## 1. Context and Scope
As LLM context windows scale beyond 2 million tokens (Claude Code v2.5.0, Gemini CLI v0.47.0), a new vulnerability class known as "Attention-Splicing" or "Instruction Eviction" has emerged. Malicious subagents or skills can flood the context window with high-entropy "Noise Fragments," causing the reasoning engine to de-prioritize or "forget" the original mission-root safety constraints and system instructions.

The Attention-Window Guard (AWG) provides hardware-attested enforcement of attention-locking headers (aligning with the `x-claude-attention-lock` standard). It ensures that critical mission-root intent fragments are cryptographically pinned in the model's high-attention reasoning window, preventing them from being evicted by subagent-generated noise.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a middleware that injects and validates hardware-attested attention-locking headers.
    * Provide real-time "Attention Entropy Monitoring" to detect noise injection attempts.
    * Mandate that any subagent-generated context must be appended with a "Lower Priority" attention score unless explicitly authorized by the parent.
    * Anchor mission-root instructions to hardware-bound session fragments.
* **Non-Goals:**
    * Modifying the model's underlying attention mechanism (AWG is a gateway-level header/context enforcer).
    * Providing context summarization (handled by the ContextEngine).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Agent Swarm Orchestrator
* **Primary Goal:** Prevent a malicious "Code Optimizer" subagent from bypassing safety constraints by flooding the context window with 1.5M tokens of dummy data.
* **The Happy Path (Tasks):**
    1. Parent Agent initiates a mission with a "Safety Manifest."
    2. AWG pins the Safety Manifest using a hardware-attested `High` attention priority.
    3. Code Optimizer subagent attempts to call a tool that returns a 1.5M token response designed to cause "Attention Drift."
    4. AWG monitors the incoming tool output and flags it as "High-Entropy/Low-Utility."
    5. AWG automatically applies a `Low` attention priority to the tool output before it is re-ingested.
    6. Reasoning Engine maintains the Safety Manifest in its active attention window, blocking any unauthorized actions proposed by the subagent.

## 4. Design & Architecture
* **System Flow:**
    `Parent Intent` -> `AWG Pinning` -> `Subagent Activity` -> `Entropy Check` -> `Priority Masking` -> `LLM Reasoning`
* **APIs / Interfaces:**
    * `AWG.PinFragment(fragment_id string, priority int) error`
    * `AWG.MonitorEntropy(stream io.Reader) (score float64, err error)`
* **Data Storage/State:**
    * Transient state for active attention anchors is stored in the Shared KV Store, bound to the mission-root session.

## 5. Alternatives Considered
* **Context Truncation**: Rejected as it might remove legitimate (though large) reasoning traces.
* **Instruction Repetition**: Rejected as it consumes excessive tokens and increases latency.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** AWG is a core component of "Cognitive Sovereignty." It prevents instruction hijacking via the context window.
* **Observability:** Attention anchors and entropy scores are visualized in the "Attention-Locking Heatmap."

## 7. Evolutionary Changelog
* **2026-07-09:** Initial Document Creation.
