# Design Document: Multimodal Inference-Time Sanitizer (MITS)

## 1. Objective
To design and implement the Multimodal Inference-Time Sanitizer (MITS) middleware for MCP Any. This component will provide real-time sanitization of non-textual reasoning traces (such as SVG, CSS, and audio metadata) using OpenClaw ContextEngine hooks to detect and neutralize "Context Smuggling" and "Prompt Path" injections via polyglot payloads.

## 2. Background
Current prompt protection mechanisms are largely text-centric, focusing on sanitizing raw text outputs for prompt injections. However, recent trends show attackers embedding malicious instructions in the metadata or noise of multimodal files (e.g., SVG paths, CSS styling, audio spectrograms) that agents process. These instructions are invisible to text-only scanners but are "seen" or "heard" by multimodal models during re-ingestion, causing unauthorized behavior. MITS addresses this gap.

## 3. Goals & Non-Goals
**Goals:**
*   Implement real-time scanning of multimodal payloads (images, audio, structured documents) for embedded instructions.
*   Integrate with OpenClaw ContextEngine lifecycle hooks for seamless inference-time interception.
*   Extend the Semantic Integrity Bridge to support multimodal inputs.
*   Prevent "Context Smuggling" attacks without significantly degrading system latency.

**Non-Goals:**
*   Building a novel multimodal LLM from scratch.
*   Replacing existing text-based sanitization tools.
*   Handling binary execution payloads (this is handled by other layers like the WASM-BSH State Sanitizer).

## 4. Proposed Solution
MITS will sit as an active interception layer in the agent's context pipeline. When a tool returns a multimodal payload, or when the agent retrieves such a payload from memory, MITS will process it before it reaches the reasoning engine.

### 4.1 Architecture
*   **ContextEngine Hook Integration:** MITS will register as a pre-inference hook in the ContextEngine Adapter.
*   **Payload Classification:** A fast heuristic pass to determine payload type (SVG, PNG, CSS, Audio).
*   **Semantic Deconstruction:**
    *   *SVG/CSS:* Parse the DOM/AST to identify anomalous text nodes, hidden elements, or path data resembling text patterns (e.g., OCR-based injection).
    *   *Audio/Images:* Perform high-entropy checks and steganography analysis to detect hidden data. Use a specialized lightweight model to check for "semantic intent" in noise.
*   **Sanitization Action:** Strip anomalous metadata, rasterize vectors to flatten hidden layers, or redact the payload entirely if confidence in malicious intent is high.

### 4.2 Security Considerations
*   MITS itself must be robust against parser-based vulnerabilities (e.g., XML External Entity (XXE) attacks when parsing SVGs).
*   The sanitization process should err on the side of caution (fail-secure).

## 5. Alternatives Considered
*   **Post-Inference Auditing:** Scanning outputs only after the LLM generates them. Rejected because the agent's internal state is already contaminated.
*   **Relying purely on the LLM's internal safety:** Rejected as LLMs are susceptible to sophisticated polyglot payloads and "jailbreaks".

## 6. Implementation Plan
1.  **Phase 1: Basic Metadata Scrubbing.** Implement hooks for SVG and CSS AST parsing and sanitization.
2.  **Phase 2: ContextEngine Integration.** Wire the metadata scrubbers into the OpenClaw ContextEngine pre-inference hooks.
3.  **Phase 3: Advanced Multimodal Analysis.** Integrate lightweight steganography and semantic intent models for image and audio payloads.
4.  **Phase 4: Testing & Telemetry.** Comprehensive testing with known polyglot payloads and integration with the Unified RL Feedback Telemetry Bridge.
