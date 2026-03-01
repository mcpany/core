# Design Doc: Dynamic Capability Bridge
**Status:** Draft
**Created:** 2026-03-01

## 1. Context and Scope
Modern agent environments like Claude Code STS (Secure Tool Sandbox) isolate tool execution for safety. However, this isolation often breaks access to essential local resources (e.g., a local Postgres DB, a project-specific filesystem, or a proprietary CLI). Simple port-forwarding is insecure and cumbersome. The Dynamic Capability Bridge allows these sandboxed agents to "request" and "mount" local capabilities on-demand through a cryptographically signed channel.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Provide a secure bridge between remote/sandboxed agent environments and local MCP tools.
    *   Implement "On-Demand Mounting" of capabilities (e.g., a sandbox "asks" for file access, and the user approves it).
    *   Use cryptographic attestation to ensure the sandbox is authorized to talk to the local MCP Any instance.
*   **Non-Goals:**
    *   Providing a general-purpose VPN or tunnel.
    *   Managing the sandbox lifecycle (MCP Any only manages the capability transport).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Developer using Claude Code in a remote sandbox.
*   **Primary Goal:** Allow the remote agent to read a local configuration file without manually copying it or exposing the entire filesystem.
*   **The Happy Path (Tasks):**
    1.  The remote agent identifies a need for a local resource (e.g., `config.json`).
    2.  The agent calls the `request_capability` tool via the bridge.
    3.  MCP Any on the local machine intercepts the request and prompts the user via HITL (Human-In-The-Loop) or a pre-defined policy.
    4.  Once approved, MCP Any creates a temporary, scoped "virtual tool" or "resource mount" accessible only to that sandbox session.
    5.  The remote agent accesses the resource through the bridge.

## 4. Design & Architecture
*   **System Flow:**
    - **Discovery**: The sandbox agent discovers the bridge via an initial handshake (likely via an environment variable `MCP_ANY_BRIDGE_URL`).
    - **Negotiation**: Capabilities are negotiated using a JSON-RPC based protocol over a secure WebSocket or mTLS tunnel.
    - **Execution**: Tool calls are proxied locally, with CoC tokens (see Chain of Custody) enforced.
*   **APIs / Interfaces:**
    - `bridge/v1/request`: Negotiation endpoint.
    - `bridge/v1/call`: Proxy endpoint for tool execution.
*   **Data Storage/State:** Session-bound capability mappings are stored in memory and cleared upon sandbox disconnection.

## 5. Alternatives Considered
*   **SSH Tunneling**: Manual port forwarding. *Rejected* as it is not agent-aware and lacks granular policy control.
*   **Custom WASM Extensions**: Writing custom sandbox extensions for every tool. *Rejected* as it is not scalable across different agent platforms.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The bridge is a high-risk surface. It must require explicit user attestation for the first connection and follow the "Principle of Least Privilege" for all mounted capabilities.
*   **Observability:** The UI should show "Active Bridges" and the specific capabilities currently shared with each.

## 7. Evolutionary Changelog
*   **2026-03-01:** Initial Document Creation. (Merged and evolved from Environment Bridging Middleware).
