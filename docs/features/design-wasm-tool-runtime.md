# Design Doc: Wasm Runtime for Tool Execution
**Status:** Draft
**Created:** 2026-03-09

## 1. Context and Scope
The recent trend towards executing AI tools in sandboxed environments (e.g., OpenClaw's Wasm shift and the CVE-2026-25253 vulnerability) has highlighted the need for MCP Any to provide a secure execution layer. Currently, MCP tools often run with the full permissions of the host process, which is a major security risk for untrusted or "shadow" tools. This design introduces a native WebAssembly (Wasm) runtime for executing tool logic within MCP Any.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Provide a secure, isolated sandbox for executing MCP tool functions.
    *   Enable fine-grained resource limits (CPU, memory, fuel) for tool execution.
    *   Ensure zero host filesystem or network access by default.
    *   Support popular Wasm-compiled languages (Go, Rust, TypeScript via Javy).
    *   Integrate with Provenance Attestation to ensure only verified binaries run natively.
*   **Non-Goals:**
    *   Replacing all native MCP servers with Wasm-only versions.
    *   Providing a full operating system emulation or persistent filesystem within Wasm.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security-conscious Agent Orchestrator.
*   **Primary Goal:** Run a community-contributed MCP tool from ClawHub without risking host compromise.
*   **The Happy Path (Tasks):**
    1.  User configures an MCP tool pointing to a `.wasm` binary URL or local path.
    2.  MCP Any loads the binary, verifies its SHA-256 hash/signature, and instantiates the Wasm runtime.
    3.  When the agent calls the tool, MCP Any executes the specific Wasm export within a restricted sandbox.
    4.  The tool returns its output JSON via shared memory, and MCP Any destroys the instance, ensuring no persistent side effects.

## 4. Design & Architecture
*   **System Flow:**
    - **Wasm Loader**: Responsible for fetching and validating the `.wasm` binary (integrates with `Provenance Attestation`).
    - **Sandbox Manager**: Configures the `Wasmtime` or `Wasmer` engine with restricted imports (no `wasi_snapshot_preview1` by default).
    - **Tool Bridge**: Maps standard MCP JSON-RPC requests to Wasm memory exports and returns outputs.
*   **APIs / Interfaces:**
    - Config extension: `execution_type: "wasm"`
    - Resource constraints: `max_memory: "128MB"`, `max_fuel: 1000000` (execution steps).
*   **Data Storage/State:** Ephemeral memory for the Wasm instance; stateful persistence requires the `Shared KV Store`.

## 5. Alternatives Considered
*   **Docker-based isolation**: *Rejected* due to high startup latency (1-2s) and overhead for simple tool calls.
*   **gRPC Process Isolation**: *Rejected* as it still requires managing host-level processes and doesn't provide the same granular resource limits as Wasm.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The primary driver. Wasm provides a "True Zero Trust" execution environment for tool logic.
*   **Observability:** The UI Roadmap's "Wasm Sandbox Monitor" will visualize fuel consumption and memory peaks.

## 7. Evolutionary Changelog
*   **2026-03-09:** Initial Document Creation.
