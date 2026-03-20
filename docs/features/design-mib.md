# Design Doc: Multi-modal Integrity Bridge (MIB)
**Status:** Draft
**Created:** 2026-05-21

## 1. Context and Scope
The discovery of "Multi-modal Trace Injection" (Context Smuggling via SVG, CSS, and audio metadata) marks a shift in agent exploitation. Malicious instructions can be embedded in non-textual data that multi-modal models "see" or "hear" during re-ingestion, bypassing text-only scanners. The MIB is required to sanitize these traces in real-time.

## 2. Goals & Non-Goals
* **Goals:**
    * Perform real-time semantic sanitization of non-textual reasoning traces (SVG, CSS, Audio metadata).
    * Neutralize "Context Smuggling" attempts before they reach the agent's reasoning engine.
    * Integrate with the existing Prompt Path Protection middleware.
* **Non-Goals:**
    * Removing legitimate multi-modal data required for reasoning.
    * Decoding encrypted or obfuscated binary payloads (handled by BSH Sanitizer).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Developer
* **Primary Goal:** Ensure a multi-modal agent doesn't ingest malicious instructions from a third-party SVG tool output.
* **The Happy Path (Tasks):**
    1. A tool returns an SVG file containing a hidden "comment" with an imperative instruction.
    2. The MIB intercepts the trace fragment.
    3. The MIB performs a "Structural Scan," identifying the non-conforming metadata.
    4. The MIB redacts or sanitizes the instruction while preserving the SVG's visual integrity.
    5. The sanitized trace is handed off to the agent reasoning engine.

## 4. Design & Architecture
* **System Flow:**
    `[Multi-modal Trace] -> [MIB Sanitizer] -> [Prompt Path Protection] -> [Agent Reasoning]`
* **APIs / Interfaces:**
    * `MIB.sanitize_fragment(trace_data, mime_type)`: Sanitizes a specific reasoning fragment.
* **Data Storage/State:**
    * Sanitization rules and known attack patterns are stored in the Metadata Governance Layer.

## 5. Alternatives Considered
* **Text-Only Stripping**: Rejected as it misses hidden instructions in non-textual fields (e.g., SVG `path` data noise).
* **Binary Blocking**: Rejected as it breaks multi-modal agent functionality.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: MIB is a mandatory gate for all non-textual tool outputs in high-trust missions.
* **Observability**: Sanitization alerts and redacted fragments are visualized in the IDS Status Monitor.

## 7. Evolutionary Changelog
* **2026-05-21:** Initial Document Creation (Upgraded from Semantic Integrity Bridge).
