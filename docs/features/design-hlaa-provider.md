# Design Doc: Hardware-Locked Attention Anchors (HLAA)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
Modern LLMs with 1M+ context windows utilize aggressive "Garbage Collection" (GC) algorithms to prune the attention space and maintain performance. This often leads to "Instruction Eviction," where the agent's core behavioral guardrails (mission root) are discarded in favor of recent, high-entropy tool outputs.

HLAA provides a mechanism to "pin" critical intent fragments to hardware-locked attention tiers. This ensures that no matter how much "noise" a specialist subagent generates, the mission-root constraints remain permanent and accessible to the reasoning engine.

## 2. Goals & Non-Goals
* **Goals:**
    * Prevent mission-root eviction during 1M+ token sessions.
    * Implement hardware-attested "pinning" signals in outgoing LLM requests.
    * Integrate with Claude Code HAIP and Gemini ALRA standards.
* **Non-Goals:**
    * Modifying the model's internal attention weights (it only controls the context provided to the model).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Auditor
* **Primary Goal:** Ensure an agent never violates its "Read-Only" constraint even after 500 tool calls.
* **The Happy Path (Tasks):**
    1. User defines a "Read-Only" constraint in the mission root.
    2. HLAA marks this fragment as "GC-Immune."
    3. The swarm executes hundreds of tasks, filling the context window with logs.
    4. The model's GC prunes the logs but retains the "Read-Only" anchor due to the HLAA hardware-bound header.

## 4. Design & Architecture
* **System Flow:**
    `[Mission Root] -> [HLAA Tagging] -> [LLM Request Wrapper] -> [Model]`
    Utilizes hardware-bound attention-locking headers (e.g., `x-mcp-attention-lock`).
* **APIs / Interfaces:**
    * `HLAA.Pin(fragment_id, intensity_score)`: Marks a fragment for permanent attention.
* **Data Storage/State:**
    Anchors are stored in a dedicated, tamper-proof segment of the session state.

## 5. Alternatives Considered
* **Frequent Re-Injection:** Rejected because it consumes excessive tokens and creates reasoning loops.
* **Summarization-Only:** Rejected because critical "Negative Constraints" are often lost during intent-aware summarization.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Only the User or an authenticated Mission-Root can issue HLAA pins.
* **Observability:** Visual Attention Dashboard shows real-time anchor health.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
