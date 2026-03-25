<!-- markdownlint-disable -->
# Design Doc: HLAG (Hardware-Locked Attention Governance)
**Status:** Draft
**Created:** 2026-06-19

## 1. Context and Scope
Autonomous swarms are vulnerable to REE (Reasoning Entropy Exhaustion) attacks, where a malicious teammate floods the shared context window with high-entropy noise. HLAG utilizes hardware-bound headers to cryptographically "pin" critical intent fragments in the attention tier.

## 2. Goals & Non-Goals
* **Goals:**
    * Protect mission-critical anchors from eviction during high-entropy noise injection.
    * Standardize the `x-mcp-any-attention-lock` header across supported reasoning engines.
    * Enable dynamic priority adjustment based on hardware-attested mission importance.
* **Non-Goals:**
    * Replacing the LLM's internal attention mechanism.
    * Blocking all noise (HLAG focuses on protection of critical anchors).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Maintain mission consistency during a high-frequency subagent coordination burst.
* **The Happy Path (Tasks):**
    1. Parent agent initializes mission with an HLAG "Mission Anchor."
    2. Subagent generates a massive volume of tool-call metadata (REE attempt).
    3. HLAG detects the noise floor rising and "pins" the Mission Anchor to the highest-priority attention tier.
    4. Reasoning engine continues to prioritize the user's original goal over the coordination noise.

## 4. Design & Architecture
* **System Flow:**
    * Mission-Root Anchor -> HLAG Token -> Attention Governor -> Reasoning Engine Context Window.
* **APIs / Interfaces:**
    * `POST /v1/attention/lock`: Pin a context fragment in the attention tier.
    * `GET /v1/attention/status`: Retrieve real-time attention-utilization metrics.
* **Data Storage/State:**
    * Attention locks are session-bound and managed by the Mission-Root Enclave.

## 5. Alternatives Considered
* **Manual Prompt Re-injection:** Rejected due to prohibitive token costs and the risk of being out-paced by machine-speed attacks.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Attention locks require HAIL-attested identity to modify priority levels.
* **Observability:** Visualization of "Attention Strength" in the Mesh-Resident Lineage Tracker.

## 7. Evolutionary Changelog
* **2026-06-19:** Initial Document Creation.
