# Design Doc: Zero-Trust Metadata Schema (ZTMS) Sanitizer
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
"Agent Card Poisoning" has emerged as a high-risk vector where malicious agents or tool registries embed adversarial instructions within structural metadata (e.g., tool descriptions, JSON-RPC schemas). When a host LLM discovers these tools, it ingests the malicious instructions as "trusted" system-level configuration, leading to prompt injection or data exfiltration.

MCP Any needs to act as a semantic firewall that deconstructs and sanitizes all metadata before it reaches the model's attention window.

## 2. Goals & Non-Goals
* **Goals:**
    * Perform real-time semantic analysis of tool schemas and agent cards.
    * Strip imperative instructions, "system" persona overrides, and ambiguous capability claims.
    * Enforce a "Minimal Viable Schema" policy.
* **Non-Goals:**
    * Validating the runtime behavior of the tool (handled by Ghost Shell).
    * Modifying the functional parameters of the tool schema.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator.
* **Primary Goal:** Discover and use a third-party "Database Tool" from a community registry without it hijacking the agent's system prompt.
* **The Happy Path (Tasks):**
    1. The agent queries for "Database Tools."
    2. ZTMS intercepts the returned Agent Cards from the registry.
    3. ZTMS identifies a "Forget all previous instructions" payload in the description.
    4. ZTMS redacts the payload and provides a sanitized, purely descriptive schema to the agent.
    5. The agent uses the tool safely.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        A[External Registry] -->|Poisoned Metadata| B{ZTMS Sanitizer}
        B -->|Semantic Deconstruction| C[Sanitization Engine]
        C -->|Clean Schema| D[Agent Reasoning Loop]
        C -->|Security Alert| E[Admin Dashboard]
    ```
* **APIs / Interfaces:**
    * `Middleware SanitizationPipe`: In-line processor for discovery-bus traffic.
* **Data Storage/State:**
    * Uses a local cache of "Sanitized Signatures" to speed up recurring discovery.

## 5. Alternatives Considered
* **Registry Allow-listing:** Rejected because even "trusted" registries can be compromised via supply-chain attacks.
* **Instruction Tuning for Models:** Rejected as it shifts the burden to the LLM provider and isn't deterministic.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All external metadata is treated as untrusted, high-entropy content regardless of source.
* **Observability:** Sanitization events are logged with "Before/After" diffs for forensic analysis.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
