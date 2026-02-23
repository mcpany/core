# Design Doc: Native Containerized MCP Execution
**Status:** Draft
**Created:** 2026-02-26

## 1. Context and Scope
With the discovery of CVE-2026-0757 (RCE in MCP managers), the current model of running MCP servers as local processes with the same privileges as the user is no longer viable for secure agentic workflows. MCP Any must provide a robust, isolated execution environment to protect the host system from malicious or compromised MCP servers.

## 2. Goals & Non-Goals
* **Goals:**
    * Automatically wrap local MCP servers (Stdio-based) in ephemeral Docker containers.
    * Provide a Zero-Trust network boundary between the container and the host.
    * Support seamless mounting of specific, authorized directories (Least Privilege).
    * Maintain performance parity with native execution.
* **Non-Goals:**
    * Building a full container orchestration system (e.g., K8s) within MCP Any.
    * Supporting non-Docker runtimes (e.g., Podman) in the first iteration.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer
* **Primary Goal:** Run a third-party MCP server (e.g., from a GitHub URL) without risking host-level file access or RCE.
* **The Happy Path (Tasks):**
    1. User adds a new MCP server to `config.yaml` with `isolation: docker` enabled.
    2. MCP Any detects the isolation requirement and pulls a base MCP-runtime image.
    3. MCP Any spawns the server inside a container with restricted CPU/Memory and no network access (unless explicitly granted).
    4. The agent calls a tool on the isolated server; MCP Any proxies the JSON-RPC over a named pipe or Unix socket.
    5. The server attempts to access `/etc/passwd`; the operation fails because only the workspace directory is mounted.

## 4. Design & Architecture
* **System Flow:**
    `Agent -> MCP Any Gateway -> Docker Exec Handler -> Isolated MCP Server (Container)`
* **APIs / Interfaces:**
    * `isolation` configuration block in `ServiceConfig`.
    * `MountPath` definitions for granular FS access.
* **Data Storage/State:**
    * Container lifecycle managed by MCP Any's background worker.
    * Ephemeral state cleared on container stop.

## 5. Alternatives Considered
* **WebAssembly (Wasm)**: Rejected for now due to limited support for standard library features (network, complex FS) in many existing MCP servers.
* **gVisor**: Excellent security but higher overhead and complexity for local development.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Containers run as non-root users; `--cap-drop ALL` by default.
* **Observability**: Logs from the containerized process are streamed back to MCP Any's central log aggregator.

## 7. Evolutionary Changelog
* **2026-02-26:** Initial Document Creation.
