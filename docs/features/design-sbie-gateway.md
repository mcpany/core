# Design Doc: SSH-Bound Isolated Execution (SBIE) Gateway
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
Recent vulnerabilities in OpenClaw (v2.x) and the "Settings-as-Shell" exploits reveal that host-level tool execution is a critical failure point. If an agent or a malicious repository can trick the gateway into running un-sanitized shell commands, the entire host is compromised.

The SBIE Gateway evolves tool execution from "Host-Local" to "Enclave-Remote" by isolating every tool call in an ephemeral SSH-based sandbox. This aligns with the new OpenShell standard in OpenClaw v2026.3.22.

## 2. Goals & Non-Goals
* **Goals:**
    * Isolate every tool-driven command in an ephemeral, resource-constrained SSH sandbox.
    * Mandate hardware-attested identity proof before establishing the SBIE tunnel.
    * Perform real-time, argument-level semantic validation (ALSV) on all tunneled commands.
    * Ensure "Zero-Host-Access" by default for all dynamic skills.
* **Non-Goals:**
    * Providing persistent storage within the sandbox (SBIE is inherently ephemeral).
    * Isolating non-command based tools (e.g., direct HTTP tools).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local Developer
* **Primary Goal:** Safely allow an agent to run `npm install` and `go test` in a repository with untrusted third-party configuration hooks.
* **The Happy Path (Tasks):**
    1. Agent requests a `shell_command` tool call.
    2. SBIE Gateway spawns a minimal, Docker-bound SSH enclave.
    3. The Gateway mounts the project root as a read-only volume (except for specific allowed output paths).
    4. The ALSV middleware sanitizes the command string for shell-fallback patterns.
    5. Command executes inside the enclave; results are streamed back via the secure tunnel.
    6. SBIE Gateway destroys the enclave immediately post-execution.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Agent->>SBIE Gateway: shell_command("npm install")
        SBIE Gateway->>ALSV: Validate Arguments
        ALSV-->>SBIE Gateway: Sanitized
        SBIE Gateway->>Docker: Spawn OpenShell Enclave
        SBIE Gateway->>Enclave: SSH Exec (npm install)
        Enclave-->>SBIE Gateway: Output/ExitCode
        SBIE Gateway->>Docker: Prune Enclave
        SBIE Gateway->>Agent: Tool Result
    ```
* **APIs / Interfaces:**
    * `Internal Tool Bridge`: Redirects all `type: command` tools to the SBIE provider.
    * `Enclave Manager`: Handles the lifecycle of SSH-ready containers.
* **Data Storage/State:**
    * Execution traces are stored in the Regulatory Vault with SBIE-lineage tags.

## 5. Alternatives Considered
* **gVisor Isolation:** Rejected as primary due to the "OpenShell" ecosystem momentum, though SBIE can utilize gVisor as a runtime.
* **Wasmtime Sandboxing:** Considered for plugins, but rejected for shell tools due to the overhead of porting binary dependencies to WASM.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Enclaves have no network access unless explicitly granted via a mission-root capability token.
* **Observability:** Enclave startup latency is tracked in the Premium Tool Execution Timeline.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
