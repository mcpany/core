# Design Doc: WASM-BSH State Sanitizer
**Status:** Draft
**Created:** 2026-03-25

## 1. Context and Scope
Binary State Handoff (BSH) is essential for high-performance agent swarms, but passing raw memory or binary blobs between agents introduces significant "State Injection" risks. A compromised agent could "poison" the context of its successor by injecting malicious payloads that the successor's model might misinterpret as legitimate instructions.

MCP Any will implement a **WASM-BSH State Sanitizer** that executes state transformation and validation logic within a secure WASM sandbox during handoffs. This ensures that all binary state conforms to predefined schemas and security profiles before reaching the target agent.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide "Active State Sanitization" for all BSH state transfers.
    * Enforce schema compliance for binary context objects.
    * Detect and strip "State Injection" patterns at the byte level.
    * Maintain sub-millisecond latency for state handoffs.
* **Non-Goals:**
    * Sanitizing textual prompts (handled by Prompt Path Protection).
    * Managing the underlying shared memory buffers (handled by Zero-Copy Transport).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Safely pass a 2GB database context from a "Data Scientist" agent to a "Reporting" agent without risk of context poisoning.
* **The Happy Path (Tasks):**
    1. "Data Scientist" agent writes context to a shared memory buffer.
    2. Handoff request is sent to the BSH Gateway.
    3. BSH Gateway invokes the WASM-BSH Sanitizer.
    4. WASM Sanitizer performs schema validation and byte-level scanning.
    5. Sanitized state is mapped to the "Reporting" agent's address space.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Agent A->>BSH Gateway: Handoff(mem_id, schema_id)
        BSH Gateway->>WASM Sandbox: Execute Sanitizer(mem_id, schema)
        WASM Sandbox-->>BSH Gateway: Sanitized Result
        BSH Gateway->>Agent B: State Mapped(mem_id)
    ```
* **APIs / Interfaces:**
    * `sanitizeState(buffer: SharedBuffer, schema: BinarySchema): SanitizedBuffer`
* **Data Storage/State:**
    * Sanitization rules and schemas are stored in the MCP Any Registry.

## 5. Alternatives Considered
* **Host-side Sanitization:** Rejected due to performance overhead and the risk of the sanitizer itself being compromised.
* **JSON-only Handoffs:** Rejected due to the "Token Storm" and serialization latency in deep swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The WASM runtime must have zero host access and limited resource quotas.
* **Observability:** Metrics on sanitization latency and rejection rates are exported to the telemetry hub.

## 7. Evolutionary Changelog
* **2026-03-25:** Initial Document Creation.

### Update: 2026-03-25 - OpenClaw v2.5 WASM-BSH & Zero-Copy Alignment
**Context:** OpenClaw v2.5 moves toward "Active State Sanitization" to prevent binary context poisoning in deep swarms.
**Architecture Adjustment:**
* Integrating the WASM-BSH Sanitizer directly into the high-speed `memfd_create` shared memory transport.
* Mandating Protobuf-level schema validation within the WASM sandbox before state is mapped into target agent memory.
* Implementing "Point-in-Time" binary scanning to detect and strip "Context Smearing" payloads at the byte level.
**Security Impact:** Prevents "Binary State Injection" while maintaining sub-millisecond, zero-copy performance for local agent coordination.
