# Design Doc: Ephemeral Containerized MCP Executor

**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
The rapid expansion of the MCP ecosystem has led to the proliferation of third-party MCP servers and the emergence of "Agent-as-a-Server" (AaaS) models. Traditional "Bare-Host" execution of MCP servers (via stdio) poses significant security risks, as a compromised or rogue server could gain full access to the host's filesystem and network. To mitigate this, MCP Any needs a mechanism to execute tools within isolated, ephemeral containers.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Launch MCP servers in isolated Docker/Podman containers on-demand.
    *   Provide strict filesystem and network egress policies for containerized tools.
    *   Support ephemeral lifecycles (container per tool-call or per session).
    *   Seamlessly bridge stdio/HTTP transport between the host and the container.
*   **Non-Goals:**
    *   Building a custom container runtime (we leverage Docker/Podman).
    *   Persistent storage for containerized tools (by default, they are ephemeral).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security-Conscious Developer.
*   **Primary Goal:** Run a potentially untrusted MCP server from a community marketplace without risking host system integrity.
*   **The Happy Path (Tasks):**
    1.  User adds a tool to `mcpany.yaml` with `execution_mode: ephemeral_container`.
    2.  User specifies a base image (e.g., `node:20-slim`) and the MCP start command.
    3.  When an agent calls the tool, MCP Any pulls the image (if needed) and starts a container.
    4.  MCP Any mounts only the specific directories required for the task.
    5.  The tool executes, results are returned, and the container is immediately destroyed.

## 4. Design & Architecture
*   **System Flow:**
    - **Executor Lifecycle**: The `EphemeralExecutor` middleware intercepts tool calls, manages container spin-up via the Docker API, and handles cleanup.
    - **Transport Bridging**: Uses named pipes or Unix sockets to bridge the host's stdio to the container's entrypoint.
    - **Resource Scoping**: Dynamically generates Docker labels and resource limits (CPU/Memory) based on the Policy Engine.
*   **APIs / Interfaces:**
    - New configuration fields in `mcpany.yaml`: `runtime: docker`, `image: string`, `mounts: []string`.
*   **Data Storage/State:**
    - Container IDs are tracked in the `Shared KV Store` for session-bound containers.

## 5. Alternatives Considered
*   **Wasm-based Sandboxing**: Executing tools in a Wasm runtime. *Rejected* due to limited library support for many existing MCP servers (especially those needing Node.js/Python).
*   **VM Isolation (Firecracker)**: Harder isolation but higher overhead. *Rejected* as a default in favor of containers, but remains a future "High-Security" option.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Containerized tools follow "Least Privilege" for filesystem mounts. Egress is blocked by default unless explicitly whitelisted in the Policy Firewall.
*   **Observability:** Container logs are streamed to the MCP Any diagnostic buffer and tagged with the Tool ID.

## 7. Evolutionary Changelog
*   **2026-02-27:** Initial Document Creation.
