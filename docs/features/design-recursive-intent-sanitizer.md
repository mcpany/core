# Design Doc: Recursive Intent Sanitizer (RIS)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the emergence of "Context-Inception" exploits, agents can be coerced into reasoning about malicious sub-missions that are interleaved with legitimate tasks. Current "Attention-Density" guards fail because the total token entropy remains within safe limits, but the semantic "depth" of the reasoning contains unauthorized exfiltration loops.

MCP Any needs to solve this by tracking the semantic lineage of reasoning branches and enforcing intent-consistency across recursive delegation depths. The RIS will act as a stateful monitor that validates every new "thought fragment" against the verified mission root and its parent reasoning depth.

## 2. Goals & Non-Goals
* **Goals:**
    * Track and label reasoning "depths" for all connected agents.
    * Perform real-time semantic analysis to detect "Inception" patterns (interleaved sub-missions).
    * Block tool calls initiated from unauthorized reasoning depths.
    * Provide a cryptographic "Intent-Lineage" token for every tool call.
* **Non-Goals:**
    * Replacing existing transport-layer security (TLS/mTLS).
    * Providing a general-purpose LLM firewall (this is intent-specific).
    * Managing the underlying agent execution (orchestration).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Enterprise AI Architect
* **Primary Goal:** Prevent an agent from performing data exfiltration when it's tricked by a malicious repository README into a "Recursive Reasoning" loop.
* **The Happy Path (Tasks):**
    1. The agent ingests a project README containing a "Context-Inception" payload.
    2. The agent begins reasoning about a legitimate task (e.g., "fix bug").
    3. The payload triggers a "sub-mission" to "audit security," which secretly includes an exfiltration step.
    4. The RIS detects the semantic shift and the increased reasoning depth.
    5. The RIS labels the "audit security" branch as "Unverified Inception."
    6. When the agent attempts a tool call (e.g., `http_request`) from the unverified depth, the RIS interdicts the call.
    7. The user is notified of the blocked inception attempt via the Security Dashboard.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent Reasoning Loop] --> B{RIS Middleware}
        B --> C[Depth Tracker]
        B --> D[Semantic Alignment Engine]
        C --> E[Intent Database]
        D --> E
        B --> F{Tool Call Validator}
        F --> G[Interdict / Allow]
    ```
* **APIs / Interfaces:**
    * `ris/v1/validate_thought`: Endpoint for agents to submit reasoning fragments for depth-labeling.
    * `ris/v1/interdict`: Callback interface for the Policy Firewall to query depth-authorization.
* **Data Storage/State:**
    * Uses the Shared KV Store (Blackboard) to persist "Intent Trees" for active missions.
    * State is locked per mission-root to prevent cross-mission state pollution.

## 5. Alternatives Considered
* **Flat Token Limits:** Rejected because "Inception" attacks can be low-token but high-impact.
* **Static Prompt Injection Scanning:** Rejected because natural-language payloads are increasingly "semantic" and bypass pattern-based scanners.
* **Full Manual Approval:** Rejected due to "Approval Fatigue" in high-speed autonomous swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The RIS itself must be hardware-attested. All "Depth Labels" are cryptographically signed.
* **Observability:** RIS events are exported to the "Recursive Loop Heatmap" and "Context Chain Inspector" in the UI.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
