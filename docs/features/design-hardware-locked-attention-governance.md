<!-- markdownlint-disable -->

# Design Doc: HLAG (Hardware-Locked Attention Governance)

**Status:** Draft
**Created:** 2026-06-19

## 1. Context and Scope

Autonomous swarms are vulnerable to Reasoning Entropy Exhaustion (REE) attacks, where malicious subagents flood the shared context with noise to "push out" mission-critical anchors. HLAG utilizes hardware-bound headers to cryptographically "pin" critical intent fragments in the attention tier.

## 2. Goals & Non-Goals

* **Goals:**

    * Prevent context-window eviction of hardware-locked mission anchors.

    * Neutralize REE attacks by prioritizing "HAAL-locked" fragments.

    * Provide a standardized "Attention Budget" for subagents.

* **Non-Goals:**

    * General-purpose context compression (handled by the ContextEngine).

## 3. Critical User Journey (CUJ)

* **User Persona:** Sovereign Agent Security Auditor

* **Primary Goal:** Ensure that "Zero Trust" policy anchors remain in the context window even during high-frequency teammate coordination.

* **The Happy Path (Tasks):**

    1. Security policy is loaded into the HLAG provider.
    2. HLAG generates a hardware-locked "Attention Pin" for the policy fragment.
    3. Specialist agents coordinate and generate high-entropy noise.
    4. The LLM provider (or local buffer) prioritizes the "Attention Pin" during pruning.
    5. The mission remains anchored to the security policy.

## 4. Design & Architecture

* **System Flow:**

    `[Sovereign Policy] -> [HLAG Pinning Engine] -> [Attention Tier (Locked)] <- [Subagent Noise]`

* **APIs / Interfaces:**

    * `x-gemini-attention-lock`: Header for pinning intent fragments.

* **Data Storage/State:** Attention pins are managed in the local HLAG state-store, anchored to the mission-root.

## 5. Alternatives Considered

* **Software-only pinning**: Rejected because the agent reasoning engine could be coerced into "un-pinning" the anchor. Hardware-locking requires external revocation.

## 6. Cross-Cutting Concerns

* **Security (Zero Trust):** HLAG prevents "Intent Smearing" by maintaining the dominance of the parent mission-root.

* **Observability:** Attention-utilization metrics are exported to the telemetry proxy.

## 7. Evolutionary Changelog

* **2026-06-19:** Initial Document Creation.
