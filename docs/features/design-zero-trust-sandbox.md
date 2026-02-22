# Design Doc: Zero Trust Tool Sandbox
**Status:** Draft
**Created:** 2026-02-25

## 1. Context and Scope
With the rise of autonomous agents like OpenClaw, there is a significant risk of unauthorized host-level access when agents execute CLI or filesystem tools. Current MCP implementations often run the server with the same privileges as the user, allowing a rogue or hallucinating agent to perform destructive actions. MCP Any needs a way to isolate these executions without sacrificing the speed of local execution.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Isolate all command-based tool executions in ephemeral Docker containers.
    *   Use named pipes for secure, low-latency communication between the host MCP Any server and the sandboxed tool.
    *   Provide strict filesystem mapping (mounts) that are read-only by default.
*   **Non-Goals:**
    *   Supporting non-Docker container runtimes in the first iteration.
    *   Providing a full virtualized desktop environment.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Local LLM Swarm Orchestrator
*   **Primary Goal:** Execute a potentially dangerous `shell_command` tool without exposing the host's `.ssh` directory.
*   **The Happy Path (Tasks):**
    1.  User configures a `command` upstream with `sandbox: enabled`.
    2.  Agent calls the tool via MCP.
    3.  MCP Any Server spawns an ephemeral Docker container using a hardened base image.
    4.  The command is executed inside the container, with only the necessary project directory mounted.
    5.  Results are streamed back via a named pipe.
    6.  Container is destroyed immediately after execution.

## 4. Design & Architecture
*   **System Flow:**
    `MCP Client -> MCP Any Core -> Sandbox Manager -> Docker Engine -> Ephemeral Container`
*   **APIs / Interfaces:**
    *   New configuration field: `upstream.sandbox.enabled: boolean`
    *   New configuration field: `upstream.sandbox.image: string`
*   **Data Storage/State:**
    *   State is strictly ephemeral. No data persists in the container after the tool call completes.

## 5. Alternatives Considered
*   **Virtual Machines (Firecracker):** Rejected for initial version due to higher cold-start latency compared to Docker containers.
*   **gVisor:** Considered but Docker with named pipes is more accessible for the current user base (OpenClaw users).

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The sandbox follows the principle of least privilege. Containers have no network access unless explicitly configured.
*   **Observability:** All sandbox logs are captured and tagged with the Tool Execution ID.

## 7. Evolutionary Changelog
*   **2026-02-25:** Initial Document Creation.
