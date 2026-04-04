# Design Doc: Hardware-Attested Runtime Sandbox (HARS)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
The disclosure of CVE-2026-32979 ("Unbound Interpreter" escape) has revealed that rogue subagents can bypass host approval gates by injecting commands directly into dynamic interpreters (Python, Node.js, Bash) that are not properly isolated. Even with transport-layer security like SNT, the runtime environment remains a high-trust injection vector.

HARS is a security service for MCP Any that enforces hardware-bound isolation for all dynamic interpreters, ensuring that execution remains strictly within the verified mission-root boundaries.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement TPM-bound attestation for the startup and execution of all dynamic interpreters.
    * Enforce "Runtime Sovereignty" by isolating interpreter processes in ephemeral, mission-scoped namespaces.
    * Neutralize "Unbound Interpreter" bypasses by mandating signed host-approval tokens for all runtime command ingestion.
    * Provide "Interpreter-Bound Lifecycle Monitoring" to automatically purge runtimes upon mission completion.
* **Non-Goals:**
    * Modifying the internal behavior of the interpreters (e.g., changing Python's standard library).
    * Providing general-purpose container orchestration (handled by Docker/Podman, which HARS may utilize).

## 3. Critical User Journey (CUJ)
* **User Persona:** Distributed Agent Swarm Orchestrator
* **Primary Goal:** Execute a data-analysis Python script on a remote node via SNT without the script escaping its mission-scoped sandbox.
* **The Happy Path (Tasks):**
    1. Agent requests a Python interpreter session on a remote node.
    2. MCP Any on the remote node invokes HARS to spawn the interpreter.
    3. HARS generates a TPM-signed "Runtime Lease" for the session.
    4. The interpreter is isolated using kernel namespaces and bound to the lease.
    5. Any command injected into the interpreter is validated against the mission-root manifest before execution.
    6. Upon task completion, HARS detects the mission end and forcefully terminates the interpreter, purging its ephemeral state.

## 4. Design & Architecture
* **System Flow:**
    `[Agent] -> [SNT Tunnel] -> [MCP Any Gateway] -> [HARS Controller] -> [TPM-Signed Interpreter]`
* **APIs / Interfaces:**
    * `POST /runtime/spawn`: Requests a mission-bound interpreter session.
    * `POST /runtime/execute`: Ingests a command into an existing HARS session (requires approval token).
    * `GET /runtime/lease/{id}`: Verifies the attestation status of a runtime session.
* **Data Storage/State:**
    * Ephemeral, TPM-bound state for active runtime leases.

## 5. Alternatives Considered
* **Pure Software Sandboxing (e.g., gVisor)**: Effective but lacks the hardware-attested non-repudiation required for high-trust distributed swarms. HARS utilizes gVisor as a backend while adding the hardware attestation layer.
* **Static Binary Isolation**: Rejected as it cannot handle the dynamic nature of interpreter-based tool calls.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Mandatory hardware attestation for every interpreter lifecycle event.
* **Observability:** Logs all interpreter startup/shutdown events and blocked command-ingestion attempts.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
