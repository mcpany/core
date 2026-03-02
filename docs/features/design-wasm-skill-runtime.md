# Design Doc: WASM Skill Runtime

**Status:** Draft
**Created:** 2026-03-02

## 1. Context and Scope
With the rise of frameworks like OpenClaw, tool definitions are evolving from static JSON schemas into "Skills as Code"—executable logic that agents can discover and run. However, executing arbitrary code from community registries (e.g., ClawHub) poses a massive security risk. MCP Any needs a secure, sandboxed execution environment to run these skills without compromising the host system.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Provide a high-performance, isolated sandbox for executing executable tool definitions (Skills).
    *   Support WebAssembly (WASM) as the primary bytecode format for skills to ensure cross-platform compatibility and security.
    *   Enforce strict capability-based access control (e.g., restricted filesystem, no network access unless explicitly granted).
    *   Enable near-instant startup times for ephemeral skill execution.
*   **Non-Goals:**
    *   Supporting arbitrary binary execution (non-WASM).
    *   Providing a full containerization solution like Docker (WASM is preferred for its lighter weight and finer-grained isolation).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Agent Developer using OpenClaw skills.
*   **Primary Goal:** Safely execute a community-contributed "Image Optimizer" skill found on ClawHub.
*   **The Happy Path (Tasks):**
    1.  The agent discovers the "Image Optimizer" skill (distributed as a `.wasm` file).
    2.  MCP Any loads the WASM module into the `WASMSkillRuntime`.
    3.  The runtime identifies that the skill requires `read` access to the `input/` directory and `write` access to `output/`.
    4.  The Policy Firewall prompts the user (or checks pre-defined rules) to grant these specific capabilities.
    5.  The skill executes within the sandbox, only accessing the permitted directories, and returns the result to the agent.

## 4. Design & Architecture
*   **System Flow:**
    - **Loading**: `SkillLoader` fetches the `.wasm` module and its manifest.
    - **Validation**: `SkillValidator` checks the cryptographic signature of the skill (using MCP Provenance Attestation).
    - **Instantiation**: `Wasmtime` (or a similar engine) creates a new isolated instance for the tool call.
    - **Capability Injection**: MCP Any "plugs in" authorized host functions (e.g., limited FS access) into the WASM environment via WASI.
*   **APIs / Interfaces:**
    - Internal `SkillRuntime` interface with `Execute(skillID, params)`.
    - WASI-compliant host exports for sandboxed I/O.
*   **Data Storage/State:** Skills are stateless by design; any persistent state must be stored via the `Shared KV Store` using authorized host calls.

## 5. Alternatives Considered
*   **Docker Containers**: Too slow for rapid tool calling and heavy on resources.
*   **gVisor/Firecracker**: Provides better isolation but significantly higher complexity and overhead for simple tool logic.
*   **Native Process Isolation (chroot/namespaces)**: Hard to maintain cross-platform and less secure than a bytecode VM like WASM.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The WASM sandbox is the ultimate fallback when a tool is "Attested" but still potentially buggy or malicious. It enforces the principle of least privilege at the instruction level.
*   **Observability:** Export WASM execution metrics (gas usage, memory consumption, execution time) to the telemetry middleware.

## 7. Evolutionary Changelog
*   **2026-03-02:** Initial Document Creation.
