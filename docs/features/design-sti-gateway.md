# Design Doc: Binary STI Gateway
**Status:** Draft
**Created:** 2026-07-02

## 1. Context and Scope
In heterogeneous agent swarms, tool invocation often relies on JSON-RPC or natural language schemas. Today's research reveals that these formats are increasingly vulnerable to **Schema-Injection** attacks, where malicious subagents or tools inject imperative instructions disguised as metadata. Furthermore, the serialization overhead of large JSON schemas in deep swarms causes "Cognitive Stall."

The Binary STI Gateway implements OpenClaw's Sovereign Tool Invocation (STI) standard, utilizing binary Protobuf for tool invocation to neutralize injection attacks and provide sub-millisecond serialization.

## 2. Goals & Non-Goals
* **Goals:**
    * Neutralize "Schema-Injection" attacks via strict binary Protobuf validation.
    * Reduce tool-call serialization latency in deep agent chains.
    * Provide a hardware-locked registry for STI-compliant tool schemas.
* **Non-Goals:**
    * Deprecating legacy JSON-RPC (STI will act as a high-trust secondary path).
    * Auto-generating Protobuf definitions from natural language (Schemas must be pre-attested).

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Trust Swarm Auditor
* **Primary Goal:** Execute a sensitive database write operation without the subagent being able to inject a "DROP TABLE" instruction via schema metadata.
* **The Happy Path (Tasks):**
    1. The Auditor registers a Protobuf-defined tool in the STI registry.
    2. The subagent attempts to call the tool using the STI binary path.
    3. The Binary STI Gateway validates the Protobuf payload against the attested schema.
    4. Any attempt to inject non-schema instructions is blocked at the binary level.
    5. The tool is executed securely and the result is returned as a binary fragment.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        A[Subagent] --> B[STI Client]
        B --> C[Binary STI Gateway]
        C --> D[Protobuf Validator]
        D --> E[Attested Registry]
        E --> F[Secure Tool Execution]
    ```
* **APIs / Interfaces:**
    * `gRPC /v1/sti/invoke`: Invokes a tool using binary Protobuf.
    * `POST /v1/sti/register`: Registers a TPM-signed Protobuf schema.
* **Data Storage/State:**
    * STI schemas are stored in the hardware-attested segment of the Shared KV Store (Blackboard).

## 5. Alternatives Considered
* **Enhanced JSON Sanitization:** Rejected as it remains vulnerable to sophisticated natural-language injection patterns.
* **WASM-based Validation:** Considered but binary STI provides better performance for high-frequency tool calls.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** STI ensures that the structure of the tool call is immutable and verified before reaching the execution layer.
* **Observability:** Tracks STI invocation frequency and schema-mismatch errors in the security dashboard.

## 7. Evolutionary Changelog
* **2026-07-02:** Initial Document Creation.
