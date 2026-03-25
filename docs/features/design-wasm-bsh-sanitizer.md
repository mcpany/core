# Design Doc: WASM-BSH State Sanitizer
**Status:** Draft
**Created:** 2026-03-25

## 1. Context and Scope
With the transition to high-speed Binary State Handoffs (BSH), the risk of "Binary Context Poisoning" has become a critical security frontier. Traditional JSON validation is too slow for sub-millisecond handoffs. The WASM-BSH State Sanitizer provides a high-performance, sandboxed environment for executing active sanitization logic on binary context buffers before they are mapped into an agent's memory.

## 2. Goals & Non-Goals
* **Goals:**
    * Execute sanitization logic within an isolated WASM sandbox (e.g., Wasmtime or Wazero).
    * Validate binary buffers against signed Protobuf/FlatBuffers schemas.
    * Neutralize "Dormant Fragments" and malicious state transformations.
    * Maintain Zero-Copy performance using shared memory regions (`memfd`).
* **Non-Goals:**
    * Modifying the business logic of the state being transferred.
    * Replacing the base BSH Gateway transport.

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Frequency Swarm Developer
* **Primary Goal:** Ensure that a binary context object passed from an untrusted "Crawler" agent to a high-trust "Analysis" agent doesn't contain smuggled instructions.
* **The Happy Path (Tasks):**
    1. Agent A pushes a BSH buffer to the Gateway.
    2. The Gateway maps the buffer into the WASM Sanitizer's memory.
    3. The Sanitizer runs a schema-validation script that checks for unauthorized metadata keys.
    4. The Sanitizer approves the buffer or redacts the malicious fragments.
    5. The sanitized buffer is mapped into Agent B's memory.

## 4. Design & Architecture
* **System Flow:**
    `[Binary Buffer] -> [Shared Memory] -> [WASM Sandbox] -> [Schema Validation] -> [Sanitized Buffer] -> [Target Agent]`
* **APIs / Interfaces:**
    * `Sanitize(buffer []byte, schemaID string) ([]byte, error)`
* **Data Storage/State:**
    * WASM modules for specific schemas are stored in the "Verified Skill Registry."

## 5. Alternatives Considered
* **Native Go Sanitizers**: Rejected due to the risk of a single vulnerability in the sanitizer compromising the entire gateway process.
* **JSON-Intermediary**: Rejected due to the extreme performance tax of binary-to-JSON-to-binary conversion.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The WASM sandbox provides defense-in-depth, ensuring that even a malicious sanitizer module cannot access the host.
* **Observability:** Sanitization latency and rejection counts are tracked in the "Binary Handoff Performance Monitor."

## 7. Evolutionary Changelog
* **2026-03-25:** Initial Document Creation.
