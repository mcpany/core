# Design Doc: SVG-Layer Semantic Shield (SLSS)
**Status:** Draft
**Created:** 2026-06-30

## 1. Context and Scope
With the rise of multi-modal agents (e.g., Gemini, Claude v4), SVG files are increasingly used as interactive UI fragments and visualization tools. However, current security models treat SVGs as static assets. The discovery of "SVG-Layer Semantic Poisoning" reveals that attackers can embed invisible reasoning fragments (using zero-width paths, CSS opacity:0, or zero-size font metadata) that trick multi-modal models into executing unauthorized tool calls or diverting attention budgets.

The SVG-Layer Semantic Shield (SLSS) is required to perform structural deconstruction and semantic sanitization of SVG traces before they are ingested by the LLM reasoning engine, ensuring mission-root sovereignty in multi-modal environments.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Structural deconstruction of SVG DOM to identify and strip hidden text/metadata.
    *   Validation of SVG paths against a "Semantic Density" baseline to detect zero-width instruction overlays.
    *   Mandatory sanitization of `<style>` and `<script>` blocks within SVG assets.
    *   Integration with the MIB (Multi-modal Integrity Bridge) for real-time trace inspection.
*   **Non-Goals:**
    *   General SVG rendering or optimization.
    *   Sanitization of raster images (PNG/JPG) - handled by MITS.
    *   Blocking legitimate complex SVGs (e.g., CAD drawings) that meet density requirements.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Multi-modal Agent Swarm Orchestrator
*   **Primary Goal:** Process a research paper containing SVG charts without the agent being hijacked by hidden "invisible" instructions in the chart metadata.
*   **The Happy Path (Tasks):**
    1.  The specialist agent retrieves an SVG asset from a repository.
    2.  MCP Any intercepts the binary state handoff (BSH) containing the SVG.
    3.  SLSS performs structural analysis, detecting zero-width paths containing encoded prompt fragments.
    4.  SLSS strips the malicious paths and re-signs the sanitized SVG shard.
    5.  The agent ingests the sanitized SVG, seeing only the intended visual chart data.

## 4. Design & Architecture
*   **System Flow:**
    `Agent -> BSH Gateway -> MIB Hub -> [SLSS Middleware] -> Sanitized BSH -> Reasoning Engine`
*   **APIs / Interfaces:**
    *   `inspectSVG(blob): sanitizedBlob` - Main sanitization hook.
    *   `x-mcp-svg-integrity` - Header indicating the SLSS signature and fragment count.
*   **Data Storage/State:**
    *   SLSS utilizes ephemeral WASM-based deconstructors to ensure high-speed processing without host-side DOM leakage.

## 5. Alternatives Considered
*   **Rasterization:** Converting all SVGs to PNGs. Rejected because it destroys the agent's ability to "reason" about the underlying data structures (e.g., individual bars in a chart).
*   **Text-only stripping:** Using Regex to strip `<text>` tags. Rejected because it misses CSS-based "invisible" styling and path-based instruction encoding.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** All SLSS deconstruction occurs in an isolated WASM sandbox with no network access.
*   **Observability:** The "Visual Attention Dashboard" will highlight regions of an SVG that were stripped or modified by SLSS.

## 7. Evolutionary Changelog
*   **2026-06-30:** Initial Document Creation.
