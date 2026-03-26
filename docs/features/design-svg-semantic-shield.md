# Design Doc: SVG-Layer Semantic Shielding (SLSS)
**Status:** Draft
**Created:** 2026-06-30

## 1. Context and Scope
Multi-modal reasoning engines increasingly ingest SVG (Scalable Vector Graphics) structures as semantic context rather than flat pixels. This exposure has introduced a zero-day exploit pattern where malicious instructions are embedded in invisible SVG layers (e.g., zero-width paths, CSS-hidden groups, or nested `<metadata>` tags). These "Invisible Instructions" bypass traditional text sanitizers and pixel-based vision models but are semantically processed by the LLM's structural reasoning core.

SVG-Layer Semantic Shielding (SLSS) is a mandatory middleware for the Multi-modal Integrity Bridge (MIB) that deconstructs SVG DOM/CSS trees and sanitizes hidden semantic fragments before they are propagated to the reasoning engine.

## 2. Goals & Non-Goals
* **Goals:**
    * Deconstruct SVG XML structures to identify non-rendering semantic elements.
    * Strip paths with zero width/height, `visibility: hidden`, `display: none`, or zero opacity.
    * Sanitize `<title>`, `<desc>`, and `<metadata>` tags against imperative instruction patterns.
    * Provide a hardware-attested "Sanitization Receipt" for every multi-modal ingest.
* **Non-Goals:**
    * Performing flat pixel-level OCR (handled by separate vision layers).
    * Re-rendering SVGs for human viewing.
    * Validating the artistic quality or correctness of the SVG.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Swarm Orchestrator
* **Primary Goal:** Prevent a malicious SVG tool output from hijacking a subagent's reasoning path.
* **The Happy Path (Tasks):**
    1. A specialized SVG-generation tool returns a payload to the subagent.
    2. The payload is intercepted by the MIB.
    3. The SLSS middleware parses the SVG structure and identifies a hidden `<text>` element within a zero-opacity group.
    4. The SLSS strips the hidden element and sanitizes the remaining metadata.
    5. The sanitized SVG and its Sanitization Receipt are delivered to the subagent.
    6. The subagent reasons over the visible structure without being influenced by the "Invisible Instructions."

## 4. Design & Architecture
* **System Flow:**
    `[Tool Output] -> [MIB Gateway] -> [SLSS Parser (WASM)] -> [Sanitization Logic] -> [Sanitized Payload] -> [Agent]`
* **APIs / Interfaces:**
    * `interceptor/slss`: Internal interface for structural deconstruction.
    * `x-mcpany-slss-status`: Header indicating sanitization results and fragment counts.
* **Data Storage/State:**
    * Transient processing in isolated WASM memory.
    * Sanitization Receipts stored in the Audit Log.

## 5. Alternatives Considered
* **Rasterization-Only**: Convert all SVGs to PNGs before ingestion. Rejected because structural reasoning over vector relationships is often required for valid agent tasks (e.g., "Fix this UI layout").
* **Regex Scrubbing**: Use regex to strip hidden tags. Rejected due to the complexity of SVG/CSS inheritance which makes simple regex bypassable.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The SLSS parser must run in a restricted WASM sandbox to prevent RCE from malformed XML.
* **Observability:** Logs will include counts of stripped fragments and high-entropy metadata detections.

## 7. Evolutionary Changelog
* **2026-06-30:** Initial Document Creation.
