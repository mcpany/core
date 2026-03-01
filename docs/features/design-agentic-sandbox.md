# Design Doc: Unified Agentic Sandbox (UAS)
**Status:** Draft
**Created:** 2026-03-01

## 1. Context and Scope
The "OpenClaw RCE" and other recent agent-based exploits have demonstrated that granting an autonomous agent direct access to the host operating system is an unacceptable risk. MCP Any must provide a "Safe Execution Environment" (SEE) where tool calls are executed in isolation. The Unified Agentic Sandbox (UAS) provides a standardized, containerized runtime for any MCP tool.

## 2. Goals & Non-Goals
* **Goals:**
    * Automatically wrap Stdio and Command-based tool calls in a transient OCI container (Docker/Podman).
    * Restrict network, filesystem, and process access for the tool within the sandbox.
    * Provide "Clean State" execution (containers are destroyed after the tool call).
* **Non-Goals:**
    * Sandboxing remote HTTP-based MCP servers (they are already isolated by the network boundary).
    * Providing long-lived persistent storage within the sandbox (use the Shared KV Store feature instead).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Execute a potentially untrusted community MCP tool (e.g., a "Web Scraper") without risking host-level file access or data exfiltration.
* **The Happy Path (Tasks):**
    1. User installs a new MCP server from the marketplace.
    2. MCP Any marks the server as "Untrusted" by default.
    3. The agent calls a tool from this server.
    4. UAS intercepts the call, spins up a minimal Alpine-based container.
    5. The tool runs inside the container with only the necessary input arguments.
    6. The container results are piped back to MCP Any.
    7. The container is immediately destroyed.

## 4. Design & Architecture
* **System Flow:**
    Agent -> MCP Any Gateway -> UAS Middleware -> [OCI Runtime (Docker/gVisor)] -> Tool Execution -> [Cleanup] -> Result -> Agent.
* **APIs / Interfaces:**
    * Configuration: `sandbox_mode: strict | permissive | none`.
    * Resource Limits: `mem_limit`, `cpu_shares` per tool.
* **Data Storage/State:**
    * Volatile memory only. Any persistent state must be explicitly written back through authorized MCP resources.

## 5. Alternatives Considered
* **Virtual Machines (VMs):** Too much overhead for transient tool calls.
* **WASM-only execution:** Many existing MCP tools are written in Python/Node and cannot be easily ported to WASM.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Use gVisor or Kata Containers for enhanced isolation where supported.
* **Observability:** Log container exit codes and resource usage metrics.

## 7. Evolutionary Changelog
* **2026-03-01:** Initial Document Creation.
