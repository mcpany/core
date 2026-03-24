# Design Doc: Reasoning-Path Validator (RPV)
**Status:** Draft
**Created:** 2026-05-05

## 1. Context and Scope
As multi-agent swarms become increasingly autonomous, the risk of "Reasoning Spoofing" and "Mission Drift" grows. Frameworks like Gemini CLI (RPW) and OpenClaw (SIA) are introducing cryptographic watermarks and reasoning audits. The Reasoning-Path Validator (RPV) is MCP Any's implementation of a unified coordination layer that verifies the cryptographic integrity and semantic alignment of an agent's internal monologue before allowing high-stakes tool interactions.

## 2. Goals & Non-Goals
* **Goals:**
    * Verify Gemini-style "Reasoning Path Watermarks" (RPW) on incoming UACO task delegations.
    * Perform "Sovereign Intent Auditing" (SIA) by semantically comparing "Reasoning Proofs" against the signed Mission Root.
    * Provide a unified RPV status for the `Risk-Adaptive CQ Controller`.
    * Log verified reasoning lineages for post-mission auditability.
* **Non-Goals:**
    * Generating watermarks (handled by the agent framework/LLM).
    * Modifying the agent's internal monologue (RPV is read-only validation).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Swarm Architect
* **Primary Goal:** Ensure that a "Specialist Subagent" cannot execute a file-deletion tool unless it provides a verifiable reasoning path that links the deletion to the primary user-signed mission.
* **The Happy Path (Tasks):**
    1. A subagent requests a high-risk tool call via UACO.
    2. The request includes an RPW watermark and an SIA reasoning proof.
    3. MCP Any's RPV middleware intercepts the request.
    4. RPV verifies the RPW signature against the parent's identity key.
    5. RPV semantically audits the SIA reasoning proof against the Mission Root (stored in the Blackboard).
    6. Upon successful validation, the tool call is forwarded to the local handler.
    7. If validation fails, the request is quarantined and the parent agent is notified of the "Reasoning Mismatch."

## 4. Design & Architecture
* **System Flow:**
    `UACO Request` -> `RPW Signature Check` -> `SIA Semantic Audit` -> `IBA Attestation` -> `Tool Handler`
* **APIs / Interfaces:**
    * `RPVMiddleware`: UACO interceptor for reasoning validation.
    * `WatermarkVerifier`: Backend service for verifying cryptographic monologues.
    * `IntentAuditor`: Semantic engine for SIA auditing.
* **Data Storage/State:**
    * RPW public keys and Mission Root intents are stored in the "Mesh-Aware Blackboard."

## 5. Alternatives Considered
* **Behavior-Only Attestation**: Rejected because it cannot detect "Reasoning Divergence" until an unauthorized action is already attempted.
* **Centralized Reasoning Auditor**: Rejected due to the "Reasoning Latency Tax" and single-point-of-failure risks.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All monologues must be cryptographically bound to a hardware-attested session. We implement "Watermark Striping" defense to prevent reasoning replay.
* **Observability:** The "Reasoning Path Explorer" in the UI allows users to visualize the verified lineage of any tool call.

## 7. Evolutionary Changelog
* **2026-05-05:** Initial Document Creation.
