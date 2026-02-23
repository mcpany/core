# Design Doc: Isolated MCP Sandbox Execution

**Status:** Draft
**Created:** 2026-02-26

## 1. Context and Scope
The discovery of CVE-2026-0757 (RCE in MCP Manager) highlights a fundamental risk: running MCP servers natively on the host system gives them broad access to local resources. If an LLM is tricked into executing a malicious tool, or if a tool itself is compromised, the host system is at risk. MCP Any needs a way to isolate MCP server execution to prevent "agentic breakout."

## 2. Goals & Non-Goals
*   **Goals:**
    *   Execute MCP servers in a restricted environment (Sandbox) with no host access by default.
    *   Support WASM (WebAssembly) as the primary isolation layer for performance and portability.
    *   Provide granular filesystem and network capability-based access to sandboxed servers.
    *   Enable "Attested Execution" where only signed binaries can be run in the sandbox.
*   **Non-Goals:**
    *   Replacing native execution entirely (native will remain an option for trusted tools).
    *   Providing a full virtual machine (too high overhead).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security-conscious developer running third-party MCP tools.
*   **Primary Goal:** Run a community-contributed "Web Scraper" tool without risking local SSH keys or sensitive files.
*   **The Happy Path (Tasks):**
    1.  User downloads a WASM-based MCP server.
    2.  User configures MCP Any to run the server in the "Isolated Sandbox" mode.
    3.  User grants the sandbox read-only access to a specific `./scratch` directory.
    4.  The LLM calls a tool on the sandboxed server.
    5.  MCP Any spawns the WASM runtime, executes the tool, and returns the result.
    6.  If the tool attempts to read `/etc/passwd`, the WASM runtime blocks the call and MCP Any logs a security violation.

## 4. Design & Architecture
*   **System Flow:**
    - **Launcher**: A new `SandboxLauncher` component in MCP Any.
    - **Runtime**: Integration with `Wasmtime` or `Wasmer` for executing WASM-compiled MCP servers.
    - **Bridge**: A JSON-RPC bridge that translates MCP protocol messages between the host and the sandbox.
    - **Capabilities**: A configuration-defined manifest that maps host resources to the sandbox (e.g., `fs:read:/tmp/mcp-scratch`).
*   **APIs / Interfaces:**
    ```yaml
    services:
      unsafe-tool:
        type: wasm
        path: ./bin/scraper.wasm
        sandbox:
          enabled: true
          allow_net: ["*.github.com"]
          allow_fs: ["./data"]
    ```
*   **Data Storage/State:** Sandbox definitions are part of the main `config.yaml`.

## 5. Alternatives Considered
*   **Docker Containers**: *Rejected* for local use cases due to startup latency and dependency on a Docker daemon.
*   **gVisor/Firecracker**: *Rejected* for most desktop users due to complexity and platform limitations, but considered as a "high-security" backend option for MCP Any server deployments.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The sandbox is the primary enforcement point. Even if the Policy Engine fails, the runtime environment blocks unauthorized access.
*   **Observability:** Sandbox violations (blocked syscalls/net calls) are logged as P0 security events in the Audit Log.

## 7. Evolutionary Changelog
*   **2026-02-26:** Initial Document Creation.
