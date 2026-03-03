# Design Doc: Sidecar Execution Engine
**Status:** Draft
**Created:** 2026-03-03

## 1. Context and Scope
CVE-2026-25253 demonstrated that running AI agents with direct host access (local execution) allows attackers to leak sensitive gateway tokens and gain RCE via the browser. To make "Local Execution" truly safe, MCP Any must isolate the tool runtime from the host gateway and the user's sensitive environment variables.

## 2. Goals & Non-Goals
* **Goals:**
    * Isolate tool execution into transient, resource-constrained environments.
    * Prevent tools from accessing the host's filesystem or network without explicit mounts/policies.
    * Sanitize or completely strip gateway-level tokens from the tool's environment.
    * Support multiple runtimes (e.g., WASM for low latency, Docker-lite for full compatibility).
* **Non-Goals:**
    * Replacing the tool's logic (the tool still does what it was designed to do).
    * Persistent sidecars (environments should be "one-and-done" or session-bound).

## 3. Critical User Journey (CUJ)
* **User Persona:** Developer using OpenClaw with MCP Any
* **Primary Goal:** Use a third-party tool from ClawHub without risking host-level compromise.
* **The Happy Path (Tasks):**
    1. User adds a tool named `unverified-web-scraper`.
    2. Agent calls the `scrape` tool.
    3. MCP Any identifies the tool as "Unattested" or "Requires Isolation."
    4. The Sidecar Execution Engine spins up a WASM sandbox.
    5. The tool's binary/script is injected into the sandbox with limited network access to the target URL only.
    6. The tool executes, returns the data to MCP Any, and the sandbox is immediately destroyed.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        Gateway[MCP Any Gateway] --> Router{Isolation Policy?}
        Router -- No --> Direct[Direct Execution - Deprecated]
        Router -- Yes --> SEE[Sidecar Execution Engine]
        SEE --> Sandbox[Transient Sandbox - WASM/Docker]
        Sandbox --> Tool[Tool Logic]
        Tool -- Result --> SEE
        SEE -- Result --> Gateway
    ```
* **APIs / Interfaces:**
    * `ExecuteIsolated(ToolDef, Args, Policy) (Result, error)`
* **Data Storage/State:**
    * No persistent state in the sandbox. Input/Output piped via standard streams or secure shared memory.

## 5. Alternatives Considered
* **OS-Level Permissions (chmod/chroot)**: Rejected as insufficient for modern multi-stage attacks and hard to manage across platforms (Windows/macOS/Linux).
* **Virtual Machines (Firecracker/QEMU)**: Rejected for tool execution due to high startup latency (seconds vs milliseconds).

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The "Communication Channel" between the host and the sidecar must be a hardened named pipe or Unix socket, not a local HTTP port (mitigating DNS rebinding).
* **Observability:** Resource usage (CPU/Mem) per tool call will be logged to detect "Crypto-jacking" or "Infinite Loop" tools.

## 7. Evolutionary Changelog
* **2026-03-03:** Initial Document Creation.
