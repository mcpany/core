# Design Doc: Memfd-Bound BSH Sanitizer
**Status:** Draft
**Created:** 2026-03-25

## 1. Context and Scope
The "Token Storm" crisis in high-density agent swarms has proven that JSON-based state transfer is no longer viable. Furthermore, passive binary validation is susceptible to TOCTOU (Time-of-Check to Time-of-Use) vulnerabilities.

The **Memfd-Bound BSH Sanitizer** addresses these challenges by integrating WASM-based binary scanning directly into zero-copy shared memory segments created via Linux `memfd_create`. This allows for sub-millisecond, hardware-accelerated state sanitization while neutralizing byte-level "Context Smearing" payloads.

## 2. Goals & Non-Goals
* **Goals:**
    * Achieve an 80% reduction in state-transfer latency for multi-gigabyte context handoffs.
    * Implement mandatory WASM-based byte-level scanning for all binary state handoffs (BSH).
    * Eliminate TOCTOU vulnerabilities using read-only memfd mappings for sanitization.
    * Support Protobuf-based schema validation within the WASM sandbox.
* **Non-Goals:**
    * Directly managing the inter-agent transport (handled by Named Pipes/WebSockets).
    * Encrypting state at rest (handled by the SRM provider).

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Density Swarm Orchestrator
* **Primary Goal:** Hand off a 500MB codebase context from a "Researcher" agent to a "Coder" agent in under 5ms without security risk.
* **The Happy Path (Tasks):**
    1. Researcher agent writes context to a `memfd` segment.
    2. Researcher agent passes the file descriptor (FD) to the MCP Any gateway.
    3. Memfd-Bound Sanitizer mounts the segment as a read-only mapping.
    4. Sanitizer executes a WASM-based security scan (e.g., detecting prompt injection in code comments).
    5. Sanitizer approves the segment and passes the FD to the Coder agent.
    6. Coder agent maps the memory and resumes reasoning instantly.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        A[Source Agent] -->|memfd FD| B[MCP Any Gateway]
        B -->|RO Mapping| C[WASM Sanitizer]
        C -->|Security Signal| D{Approved?}
        D -->|Yes| E[Target Agent]
        D -->|No| F[Quarantine & Alert]
    ```
* **APIs / Interfaces:**
    * `bsh.SanitizeMemfd(fd int, schema string) -> result`
    * `bsh.MapSegment(token string) -> fd`
* **Data Storage/State:**
    * Anonymous shared memory (memfd) segments, ephemeral and session-bound.

## 5. Alternatives Considered
* **JSON Serialization:** Rejected due to O(n) overhead and "Token Storm" performance degradation.
* **Shared Filesystem:** Rejected due to I/O latency and permission complexity in containerized swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Sanitizer operates in a strictly resource-bounded WASM sandbox.
* **Observability:** Performance metrics (sanitization time, memory usage) are exported to the "Zero-Copy Transport Monitor."

## 7. Evolutionary Changelog
* **2026-03-25:** Initial Document Creation.
