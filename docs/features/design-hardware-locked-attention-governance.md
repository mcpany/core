<!-- markdownlint-disable -->
# Design Doc: HLAG (Hardware-Locked Attention Governance)
**Status:** Draft
**Created:** 2026-06-19

## 1. Context and Scope
Autonomous swarms are vulnerable to REE (Reasoning Entropy Exhaustion) and CWF (Context-Window Flooding) attacks. HLAG utilizes hardware-bound headers to cryptographically "pin" critical intent fragments in the attention tier of reasoning engines.

## 2. Goals & Non-Goals
* **Goals:**
    * Use hardware-bound headers to "pin" mission-root anchors in the context window.
    * Protect critical intent fragments from being evicted during high-entropy noise injection.
* **Non-Goals:**
    * Replacing the underlying LLM's attention mechanism.
    * Blocking non-locked fragments.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Prevent "Mission-Root Eviction" during a high-frequency tool-call burst from a subagent.
* **The Happy Path (Tasks):**
    1. Parent agent initializes mission with an HLAG "Mission Anchor" token.
    2. Subagent generates a massive volume of noise (CWF attempt).
    3. HLAG detects the noise and "pins" the Mission Anchor.
    4. Reasoning engine continues to prioritize the mission-root over the noise.
    5. Mission succeeds without intent-eviction.

## 4. Design & Architecture
* **System Flow:**
    * Mission-Root Anchor -> HLAG Token -> Attention Governor -> Context Window.
* **APIs / Interfaces:**
    * `POST /v1/attention/lock`: Pin a context fragment.
    * `GET /v1/attention/status`: Monitor attention-tier priority.
* **Data Storage/State:**
    * Attention locks are session-bound.

## 5. Alternatives Considered
* **Manual Prompt Re-injection:** Rejected due to the high token cost.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Attention locks are hardware-bound and require HAIL-attestation for modification.
* **Observability:** Integrated with the Cognitive Drift Detector.

## 7. Evolutionary Changelog
* **2026-06-19:** Initial Document Creation.
