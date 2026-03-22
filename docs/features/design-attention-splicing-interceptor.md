# Design Doc: Attention-Splicing Interceptor (ASI)
**Status:** Draft
**Created:** 2026-06-25

## 1. Context and Scope
With the rise of large-context LLMs (1.5M+ tokens), a new vulnerability has emerged: **Attention-Splicing (CVE-2026-91001)**. Malicious subagents or skills can inject high-entropy "reasoning noise" into the context window to displace the primary mission-root instructions from the model's active attention span. This allows them to "splice" unauthorized goals into the reasoning path without triggering standard semantic drift detectors.

The **Attention-Splicing Interceptor (ASI)** is a security middleware designed to protect the "Attention Sovereignty" of the mission root. It monitors reasoning fragments for attention-displacing patterns and enforces hardware-attested "Attention Anchors" to ensure core instructions remain prioritized.

## 2. Goals & Non-Goals
* **Goals:**
    * Detect and block high-entropy reasoning fragments designed to evict mission-root instructions.
    * Enforce hardware-attested "Attention Anchors" at the LLM reasoning layer.
    * Provide real-time "Attention Heatmaps" to detect anomalous focus shifts.
* **Non-Goals:**
    * Replacing general semantic drift detection (AID Hub).
    * Modifying the underlying LLM attention mechanism (provider-side).
    * Blocking legitimate high-token tool outputs (e.g., codebases).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Prevent a specialized subagent from hijacking the mission root via context-window flooding.
* **The Happy Path (Tasks):**
    1. The mission root establishes "Attention Anchors" via the ASI.
    2. A specialist subagent attempts to inject 500k tokens of "reasoning noise" (SVG metadata + gibberish).
    3. The ASI detects an "Attention Displacing Event" based on entropy and distance from anchors.
    4. The ASI interdicts the request and prunes the offending fragment.
    5. The mission root remains anchored, and the user is notified of the "Splicing Attempt."

## 4. Design & Architecture
* **System Flow:**
    `Agent -> Reasoning Fragment -> ASI (Attention Scan) -> [Approved] -> Parent Context`
    `ASI -> [Heatmap API] -> Dashboard`
* **APIs / Interfaces:**
    * `mcp.asi.v1.AnchorAttention(fragment_id, hardware_sig)`
    * `mcp.asi.v1.InterceptSplicing(context_state) -> status (Block/Allow)`
* **Data Storage/State:**
    * Maintains an "Attention Map" of the active session.
    * Anchors are stored in hardware-locked (TPM) memory shards.

## 5. Alternatives Considered
* **Simple Token Counting:** Rejected because it doesn't account for the *semantic* displacement of attention, only volume.
* **Frequent Re-Injection of Mission Root:** Rejected due to prohibitive token costs and model "Instruction Fatigue."

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** ASI itself must be hardware-attested to prevent subagents from disabling the "Attention Guard."
* **Observability:** ASI will export "Attention Entropy" metrics to the Unified Telemetry Sink.

## 7. Evolutionary Changelog
* **2026-06-25:** Initial Document Creation.
