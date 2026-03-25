# Design Doc: WASM-BSH State Sanitizer
**Status:** Draft
**Created:** 2026-03-25

## 1. Context and Scope
Inter-agent coordination in deep swarms increasingly relies on Binary State Handoff (BSH) to mitigate "Token Storms" and reduce latency. However, binary state is a high-trust input that can contain "State Injections" designed to exploit the recipient agent's logic.

MCP Any must provide an isolated environment to execute active sanitization logic on binary context fragments *before* they are ingested by the target agent. WASM provides the ideal sandbox for this, ensuring that sanitization code itself cannot compromise the host.

## 2. Goals & Non-Goals
* **Goals:**
    * Execute active sanitization logic in a secure WASM sandbox during BSH.
    * Validate BSH fragments against signed schemas.
    * Perform semantic sanitization (e.g., stripping imperative commands from state).
    * Support pluggable sanitization modules for different agent frameworks.
* **Non-Goals:**
    * Directly modifying the recipient agent's memory.
    * Translating between different binary formats (handled by the BSH Gateway).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Ensure that a binary state handoff from an untrusted specialist agent (e.g., a "Web Search" agent) is free of malicious instructions before being ingested by the "Mission Root" agent.
* **The Happy Path (Tasks):**
    1. Specialist agent prepares a BSH packet.
    2. Packet is sent to the BSH Gateway.
    3. Gateway triggers the WASM-BSH State Sanitizer.
    4. Sanitizer loads the appropriate WASM module for the target schema.
    5. WASM module validates and scrubs the binary buffer.
    6. Sanitized buffer is handed off to the recipient agent.

## 4. Design & Architecture
* **System Flow:**
    ```
    [BSH Gateway] -> [WASM Runtime] -> [Sanitized Buffer]
                           |
                    [Module Registry]
    ```
* **APIs / Interfaces:**
    * `Sanitize(buffer []byte, schemaID string) ([]byte, error)`: Core sanitization interface.
    * WASM Export: `wasm_sanitize(ptr, len) -> (ptr, len)`.
* **Data Storage/State:**
    * WASM modules are stored in the Verified Skill Registry.

## 5. Alternatives Considered
* **Native Go Sanitization:** Rejected because it doesn't allow for framework-specific, third-party sanitization logic without recompiling MCP Any.
* **JSON-based Pre-filtering:** Rejected because it re-introduces the performance overhead that BSH was designed to eliminate.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** WASM modules are executed with zero host capability access.
* **Observability:** Sanitization latency and rejection rates are exported as Prometheus metrics.

## 7. Evolutionary Changelog
* **2026-03-25:** Initial Document Creation.
