# Design Doc: WASM-MCP Runtime Layer

**Status:** Draft
**Created:** 2026-02-25

## 1. Context and Scope
Local execution of MCP servers via `stdio` (Command adapter) presents a significant security risk. If an agent is compromised via prompt injection, it can potentially execute arbitrary commands on the host machine. To achieve Zero-Trust, we need a way to execute local tools in a perfectly isolated sandbox. WebAssembly (WASM) provides this capability by allowing code to run in a memory-safe, capability-restricted environment.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Provide a secure, isolated runtime for executing `.wasm` MCP servers.
    *   Strictly control access to host resources (filesystem, network, environment) via WASI (WebAssembly System Interface).
    *   Integrate with the existing `Upstream` interface for seamless replacement of Command adapters.
    *   Support high-performance execution using a production-grade WASM engine (e.g., Wasmtime or Wazero).
*   **Non-Goals:**
    *   Transpiling existing Go/Python MCP servers to WASM automatically.
    *   Running non-WASM binaries in the sandbox.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security-conscious Enterprise Architect.
*   **Primary Goal:** Execute a local file-processing tool without giving the agent full access to the server's filesystem.
*   **The Happy Path (Tasks):**
    1.  Architect compiles a specialized file-processing MCP server to `.wasm`.
    2.  Architect configures MCP Any to use the `wasm` adapter for this tool, granting only `fs:read` access to `/tmp/uploads`.
    3.  Agent calls the tool.
    4.  MCP Any loads the `.wasm` module, initializes the sandbox with restricted WASI permissions, and executes the call.
    5.  The tool can only see `/tmp/uploads` and cannot access network or other host files.

## 4. Design & Architecture
*   **System Flow:**
    - **Configuration**: The `Upstream` config defines the path to the `.wasm` file and the specific capabilities (WASI permissions).
    - **Initialization**: Upon first call, the `WasmAdapter` compiles/instantiates the module using a pooled engine.
    - **Execution**: MCP JSON-RPC messages are passed into the WASM module via a standardized guest interface (e.g., `mcp_handle_request`).
*   **APIs / Interfaces:**
    ```yaml
    upstream:
      type: wasm
      path: ./tools/processor.wasm
      capabilities:
        filesystem:
          - host: "/tmp/uploads"
            guest: "/data"
            readOnly: true
        network: false
    ```
*   **Data Storage/State:** WASM modules are stateless by default. Any persistence must be handled via the `Shared KV Store` (Blackboard) or explicitly mapped filesystem volumes.

## 5. Alternatives Considered
*   **Docker/Containers**: *Rejected* due to higher startup latency and resource overhead for small tool executions.
*   **gVisor/Sandboxed Processes**: *Rejected* because they are platform-dependent and harder to package as "portable tools."

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The primary driver. WASM provides hardware-level isolation and explicit capability granting.
*   **Observability:** Monitor WASM execution time, memory usage, and WASI violations (attempted unauthorized access).

## 7. Evolutionary Changelog
*   **2026-02-25:** Initial Document Creation.
