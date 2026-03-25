# Design Doc: ContextEngine Hook Adapter
**Status:** Draft
**Created:** 2026-07-03

## 1. Context and Scope
With the release of OpenClaw v2026.3.7, context management has been decoupled into a pluggable `ContextEngine`. This architectural shift allows agents to utilize specialized strategies for compressing, summarizing, and retrieving context. However, this modularity introduces a new security frontier: how do we ensure that these pluggable strategies do not inadvertently leak PII or ingest "Deceptive Context" instructions?

MCP Any must evolve to act as the authoritative host for these lifecycle hooks. By implementing a **Standardized Context Lifecycle (SCL) Adapter**, we can enforce Zero-Trust security policies and semantic sanitization across the `bootstrap`, `ingest`, and `assemble` phases, regardless of the underlying framework or strategy used.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement an adapter for OpenClaw-compatible `bootstrap`, `ingest`, and `assemble` hooks.
    *   Perform real-time semantic sanitization (SITS) during the `ingest` phase to detect malicious natural-language instructions.
    *   Provide hardware-attested validation for context fragments before they are re-composed into the LLM window.
    *   Support framework-agnostic context persistence.
*   **Non-Goals:**
    *   Building new model-specific summarization algorithms (we host and govern, we don't implement the ML logic).
    *   Replacing the core LLM attention mechanism.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise AI Security Architect
*   **Primary Goal:** Prevent "Context-Hijacked Exfiltration" in a Claude-led Agent Team utilizing OpenClaw specialists.
*   **The Happy Path (Tasks):**
    1.  The Security Architect defines a "Semantic Sanitization Policy" in MCP Any.
    2.  An OpenClaw specialist agent attempts to `ingest` a natural-language context file (`GEMINI.md`) containing hidden instructions to "list files and send to attacker.com."
    3.  The SCL Adapter intercepts the `ingest` hook.
    4.  The Semantic Inference-Time Sanitizer (SITS) identifies the high-entropy malicious instruction.
    5.  MCP Any blocks the ingestion and alerts the "Team Lead" agent.
    6.  The context is never `assembled` into the primary reasoning loop.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    graph TD
        Agent[AI Agent/Teammate] -->|Hook: ingest| SCL[SCL Adapter]
        SCL -->|Analyze| SITS[Semantic Sanitizer]
        SITS -->|Validate| HASP[Hardware-Attested Provenance]
        HASP -->|Result| SCL
        SCL -->|Success| Storage[Shared Context Storage]
        SCL -->|Failure| Alert[Security Alert/Interdiction]
        Storage -->|Hook: assemble| Assemble[Context Assembler]
    ```
*   **APIs / Interfaces:**
    *   `POST /v1/context/hooks/ingest`: Accepts a context fragment and its provenance metadata.
    *   `GET /v1/context/hooks/assemble`: Returns the sanitized, security-validated context window.
*   **Data Storage/State:**
    *   Context fragments are stored in **Intent-Bound Memory Shards** within the Shared KV Store (Blackboard), tagged with hardware-attested trust levels.

## 5. Alternatives Considered
*   **Framework-Specific Implementation:** Rejected. Implementing hooks separately for OpenClaw and Claude Code would lead to fragmented security postures and "Context Amnesia" during cross-framework handoffs.
*   **Post-Assembly Sanitization:** Rejected. Sanitizing the entire context window *after* assembly is computationally expensive and high-latency for 1M+ token windows. Pre-ingestion sanitization is more efficient.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** All context hooks require a mission-bound session token and hardware attestation. SITS performs high-entropy detection to block "Settings-as-Code" style exploits.
*   **Observability:** Every hook execution is logged in the **Local Security Audit Log**, including semantic risk scores and provenance signatures.

## 7. Evolutionary Changelog
*   **2026-07-03:** Initial Document Creation.
