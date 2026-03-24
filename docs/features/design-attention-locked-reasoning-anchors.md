# Design Doc: Attention-Locked Reasoning Anchors (ALRA)
**Status:** Draft
**Created:** 2026-07-01

## 1. Context and Scope
As context windows in frontier models expand to 1M+ tokens (Gemini CLI), a new class of vulnerability known as "Context-Window Flooding" (CWF) has emerged. Malicious subagents or "Injected Context" files (e.g., `GEMINI.md`) can flood the attention layer with high-entropy noise, forcing the model to "evict" the primary mission-root instructions. This leads to intent drift or unauthorized tool execution driven by the injected noise.

The **Attention-Locked Reasoning Anchors (ALRA)** system provides a hardware-attested mechanism to protect the mission root. It utilizes specialized headers and cryptographic binding to ensure that critical intent fragments remain "pinned" at the LLM's attention layer, regardless of the surrounding context density.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement hardware-bound attention locking for mission-critical intent fragments.
    * Provide a mechanism for users to "Audit & Pin" specific context fragments.
    * Neutralize CWF attacks by ensuring core instructions cannot be evicted.
    * Support cross-framework attestation for pinned anchors (Claude Code, OpenClaw).
* **Non-Goals:**
    * Modifying the underlying transformer architecture (ALRA works at the protocol/prompting layer).
    * Restricting legitimate reasoning expansion.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Agent Orchestrator
* **Primary Goal:** Ensure that "Zero-Deletions Policy" remains active even if a subagent generates 500,000 tokens of research data.
* **The Happy Path (Tasks):**
    1. The user defines a "Mission Root" with a strict `no-deletion` policy.
    2. MCP Any generates an ALRA-token for this fragment, signed by the local TPM.
    3. The agent begins a massive research task, filling the 1M token context.
    4. Malicious injected noise tries to shadow the `no-deletion` rule.
    5. The ALRA middleware detects the attention-density shift and re-injects the pinned anchor with hardware-verified priority.
    6. The model remains anchored to the user's original constraint.

## 4. Design & Architecture
* **System Flow:**
    `User Intent` -> `ALRA Anchor Generator` -> `TPM Signing` -> `Prompt Injection Engine` -> `LLM Attention Layer`
* **APIs / Interfaces:**
    * `ALRA.pinFragment(fragment_id, priority)`
    * `ALRA.verifyAttentionStability(context_state)`
    * `Header: x-mcpany-attention-lock: [attestation_payload]`
* **Data Storage/State:**
    * Anchors are stored in the **Mesh-Aware Blackboard** with an `is_pinned` attribute.

## 5. Alternatives Considered
* **Reprompting Every Turn:** Rejected due to excessive token cost and latency.
* **Logit Biasing:** Rejected as it is model-specific and lacks framework-neutrality.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Pinned anchors are immutable once signed by the user's TPM.
* **Observability:** Attention status is visualized in the "Visual Attention Dashboard."

## 7. Evolutionary Changelog
* **2026-07-01:** Initial Document Creation.
