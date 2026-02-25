# Design Doc: WASM-MCP Runtime (Sandboxed Tools)

**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
The current MCP model relies heavily on `stdio` and `http` for local tool execution. While effective, this presents a significant security risk when running third-party or unverified MCP servers, as they often have full access to the host environment. Recent supply chain attacks (e.g., Clinejection) have highlighted the need for a secure, isolated runtime. MCP Any will provide a WebAssembly (WASM) based runtime to host local tools in a cryptographically isolated sandbox.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Execute local MCP servers within a secure WASM sandbox.
    *   Provide strict, capability-based access to host resources (filesystem, network) via WASI.
    *   Enable "One-Click Safe Install" for marketplace tools.
    *   Support near-native performance for sandboxed tools.
*   **Non-Goals:**
    *   Sandboxing remote HTTP-based MCP servers (they are already isolated by the network).
    *   Rewriting existing tools to WASM (users provide compiled `.wasm` modules).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security-conscious Developer.
*   **Primary Goal:** Run a new "Code Optimizer" MCP tool from an untrusted source without risking local file access.
*   **The Happy Path (Tasks):**
    1.  User downloads `optimizer.wasm` from the MCP Marketplace.
    2.  User configures MCP Any to run this tool using the `wasm` transport.
    3.  User grants the tool specific read-only access to `/src/project/`.
    4.  LLM calls the tool; MCP Any executes it within the Wasmtime-backed runtime.
    5.  The tool attempts to access `~/.ssh/id_rsa` and is blocked by the WASM sandbox.

## 4. Design & Architecture
*   **System Flow:**
    - **Host Environment**: MCP Any (Go) embeds a WASM runtime (e.g., `wasmer-go` or `wasmtime-go`).
    - **Plugin Lifecycle**: Tools are loaded as WASM modules.
    - **Communication**: MCP JSON-RPC is piped through WASM `stdin`/`stdout`.
    - **Resource Control**: Host resources are mapped to the sandbox using WASI (WebAssembly System Interface) with explicit permission grants.
*   **APIs / Interfaces:**
    ```yaml
    services:
      untrusted-tool:
        transport: wasm
        path: ./tools/optimizer.wasm
        permissions:
          filesystem:
            - path: /src/project/
              mode: ro
    ```
*   **Data Storage/State:** WASM modules are stored in the local `plugins/` directory.

## 5. Alternatives Considered
*   **Docker Containers**: Too heavy-weight for simple local tools; slow startup.
*   **Virtual Machines**: Even heavier; overkill for most MCP use cases.
*   **Native OS Sandboxing (jail/chroot)**: Lacks cross-platform consistency and is difficult to configure securely.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The sandbox is the primary security layer. Even if the tool is malicious, it cannot escape the WASM environment.
*   **Observability:** Track memory and CPU usage of the WASM module to prevent resource exhaustion attacks (denial of service).

## 7. Evolutionary Changelog
*   **2026-02-27:** Initial Document Creation.
