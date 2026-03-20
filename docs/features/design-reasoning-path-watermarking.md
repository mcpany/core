# Design Doc: Reasoning-Path Watermarking Provider
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
In deep swarms, tracing the "Reasoning Lineage" of a high-risk tool call is
difficult, leading to "Reasoning Hijacking" (e.g., Gemini CLI v0.41.0
issues). Misaligned subagents can inject their own logic into a parent's
stream, altering the mission root without detection. The Reasoning-Path
Watermarking Provider ensures that every reasoning fragment is
cryptographically bound to the mission-root identity and non-repudiable.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement cryptographically signed watermarking for all reasoning
      fragments.
    * Mandate hardware-attested identity binding for every chain-of-thought
      step.
    * Provide real-time "Reasoning Hijack" detection via watermark
      validation.
* **Non-Goals:**
    * Redacting sensitive PII from monologues (handled by the PII-Sovereign
      Scrubber).
    * Summarizing long reasoning traces (handled by the ContextEngine).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Verify that a tool-calling instruction from a subagent
  is a direct descendant of the parent agent's reasoning path.
* **The Happy Path (Tasks):**
    1. Parent agent (Mission Root) generates a reasoning fragment.
    2. Reasoning-Path Watermarking Provider signs the fragment with the
       mission-root TPM key.
    3. Specialist subagent generates a sub-instruction.
    4. Reasoning-Path Watermarking Provider binds the sub-instruction to the
       parent's signed watermark.
    5. Tool execution is only granted if the watermarked lineage is verified.

## 4. Design & Architecture
* **System Flow:** LLM Output -> Watermarking Provider -> SRM Storage ->
  Mission-Root Identity Verification.
* **APIs / Interfaces:** `POST /api/v1/reasoning/watermark/sign` requiring
  a mission-root session-token.
* **Data Storage/State:** Cryptographic hash-chain storage for watermarked
  fragments.

## 5. Alternatives Considered
* **Lineage-Header Injection:** Rejected as headers are easily evictable
  during aggressive context compression.
* **Centralized Reasoning Audit:** Rejected due to performance and privacy
  constraints.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All watermarks must be hardware-attested and
  mission-bound.
* **Observability:** Every reasoning fragment is audit-ready for non-
  repudiation analysis.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
