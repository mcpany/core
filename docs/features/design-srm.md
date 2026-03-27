# Design Doc: Spectral Reasoning Mitigator (SRM)
**Status:** Draft
**Created:** 2026-07-08

## 1. Context and Scope
Spectral Reasoning attacks utilize timing analysis of LLM reasoning monologues (ARE headers) to infer mission-root constraints, even when the final output is sanitized. MCP Any needs a defense that decouples reasoning latency from state visibility.

The SRM provides "Reasoning Jitter" and "Temporal Attention Masking" to neutralize these side-channels.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Inject hardware-attested timing jitter into all mission-bound responses.
    *   Mask attention-utilization metrics from subagent observation.
    *   Maintain <5ms overhead for non-adversarial reasoning paths.
*   **Non-Goals:**
    *   Will NOT modify the core LLM reasoning weights.
    *   Will NOT perform textual sanitization (handled by AID Hub).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Swarm Architect
*   **Primary Goal:** Prevent a specialist subagent from inferring project-local secrets via reasoning latency.
*   **The Happy Path (Tasks):**
    1.  Architect enables `SRM_ENABLED` in global policy.
    2.  Subagent executes a high-stakes reasoning loop.
    3.  SRM intercepts the ARE headers and calculates a "Reasoning-Aware Jitter" profile.
    4.  Response is held in a kernel-bound buffer and released according to the jitter profile.
    5.  Attestation tokens are appended to the response to verify the defense was active.

## 4. Design & Architecture
*   **System Flow:**
    `[Subagent] -> [ARE Headers] -> [SRM Jitter Engine] -> [Kernel Buffer] -> [MCP Any Gateway]`
*   **APIs / Interfaces:**
    *   `POST /v1/srm/jitter-profile`: Internal endpoint for profile calculation.
    *   `GET /v1/srm/attestation/{session_id}`: Retrieve hardware-bound proof of mitigation.
*   **Data Storage/State:**
    *   Timing profiles are stored in ephemeral, kernel-locked memory.

## 5. Alternatives Considered
*   **Random Jitter:** Rejected due to statistical predictability; reasoning-aware jitter provides higher entropy against timing attacks.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** SRM profiles are cryptographically bound to the mission-root identity.
*   **Observability:** SRM latency and jitter variance are exported via the Telemetry Sink.

## 7. Evolutionary Changelog
*   **2026-07-08:** Initial Document Creation.
