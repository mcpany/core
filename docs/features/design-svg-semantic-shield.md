# Design Doc: SVG-Layer Semantic Shield (SLSS)
**Status:** Draft
**Created:** 2026-06-30

## 1. Context and Scope
Multi-modal AI agents often ingest SVG XML structures to reason about visual layouts and diagrams. The discovery of "SVG-Layer Semantic Poisoning" reveals that attackers can embed malicious text instructions within invisible SVG layers (e.g., `<text>` nodes with `opacity: 0`, `display: none`, or zero-width paths).

MCP Any needs to provide a sanitization layer that deconstructs SVG DOM/CSS structures and strips non-visible semantic fragments before they reach the agent's reasoning engine.

## 2. Goals & Non-Goals
* **Goals:**
    * Deconstruct SVG XML and identify hidden text-bearing nodes.
    * Strip any semantic fragments that are not visible based on standard CSS rendering rules.
    * Provide a hardware-attested sanitization receipt for the multi-modal trace.
* **Non-Goals:**
    * Performing full OCR or pixel-level analysis (handled by model-side vision).
    * Modifying visible SVG paths (utility preservation).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-modal Swarm Developer
* **Primary Goal:** Sanitize untrusted SVGs from a repository before feeding them to a specialist agent.
* **The Happy Path (Tasks):**
    1. Agent receives an untrusted SVG file from a tool call.
    2. SVG is passed through the SLSS Middleware.
    3. SLSS deconstructs the DOM and evaluates visibility (CSS/attributes).
    4. Hidden nodes containing imperative instructions are purged.
    5. A "Sanitization Receipt" is appended to the context.

## 4. Design & Architecture
* **System Flow:**
    `Untrusted SVG` -> `DOM Parser (WASM)` -> `Visibility Evaluator` -> `Semantic Stripper` -> `Sanitized SVG`
* **APIs / Interfaces:**
    * `POST /v1/sanitize/svg`: Endpoint for structural deconstruction.
    * Response includes `sanitized_svg` and `attestation_token`.
* **Data Storage/State:** State-free; processing happens in-memory within an ephemeral WASM sandbox.

## 5. Alternatives Considered
* **Rasterization:** Converting SVGs to PNGs would mitigate structural injection but would destroy the agent's ability to reason about precise SVG metadata/identifiers.
* **Schema Validation:** Rejected as it cannot distinguish between valid structural nodes and those used for instruction injection.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The parser must run in a restricted WASM environment to prevent "Billion Laughs" style XML attacks.
* **Observability:** Log the count and content of "Hidden Instruction Fragments" identified during sanitization.

## 7. Evolutionary Changelog
* **2026-06-30:** Initial Document Creation.
