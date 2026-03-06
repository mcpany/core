# Design Doc: Edge-to-Cloud Local Bridge

**Status:** Draft
**Created:** 2026-03-02

## 1. Context and Scope
With the introduction of fees for self-hosted GitHub runners (March 2026), a significant portion of the agent ecosystem is migrating to cloud-managed sandboxes (e.g., Anthropic's Sandbox, GitHub Actions). These cloud agents frequently need to access local enterprise tools, databases, or filesystems that are not (and should not be) exposed to the public internet. MCP Any needs to provide a secure, low-latency "Reach-Back" bridge that allows these cloud-based agents to securely interact with local MCP servers.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Enable cloud-based AI agents to securely call local MCP tools.
    *   Provide a "Zero-Config" tunneling mechanism for authorized cloud runners.
    *   Maintain strict Zero-Trust boundaries: the cloud agent can ONLY see whitelisted local tools.
    *   Minimize latency to support real-time tool interactions.
*   **Non-Goals:**
    *   Exposing the entire local network to the cloud.
    *   Providing a general-purpose VPN or proxy.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise DevOps Engineer using GitHub Actions for Agentic CI/CD.
*   **Primary Goal:** Allow a cloud-based "CI Reviewer Agent" to run a local proprietary security scanner on the source code via MCP.
*   **The Happy Path (Tasks):**
    1.  Engineer starts MCP Any on a local workstation with the `--enable-cloud-bridge` flag.
    2.  MCP Any generates a one-time "Bridge Attestation Token."
    3.  The cloud agent (GitHub Action) uses this token to establish a secure WebSocket-based tunnel to the local MCP Any instance.
    4.  The cloud agent lists and calls the "Local Security Scanner" tool as if it were a local MCP server.
    5.  MCP Any executes the local tool and returns the results through the encrypted tunnel.

## 4. Design & Architecture
*   **System Flow:**
    - **Tunnelling**: Uses a reverse-proxy WebSocket tunnel initiated from the local environment to a central (or private) "Bridge Relay."
    - **Authentication**: Mutual TLS (mTLS) combined with the "Bridge Attestation Token."
    - **Filtering**: The `BridgeProxyMiddleware` intercepts incoming cloud requests and filters them against the `local_whitelist` configuration.
*   **APIs / Interfaces:**
    - **Cloud Interface**: A secure WebSocket endpoint for the agent to connect to.
    - **Local Interface**: The standard MCP Any tool registry.
*   **Data Storage/State:** Bridge session metadata (connected runners, active calls) is stored in the local SQLite database.

## 5. Alternatives Considered
*   **Public Exposure (Ngrok-style)**: Exposing the local MCP port to the public internet. *Rejected* due to extreme security risks (e.g., "8,000 Exposed Servers" incident).
*   **VPN/Tailscale**: Requiring the cloud runner to join a mesh network. *Rejected* as it is often too complex for ephemeral cloud runners and provides too much network access.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** All "Reach-Back" calls are subjected to the same Policy Firewall rules as local calls. Data Loss Prevention (DLP) filters are mandatory for all responses exiting the local network.
*   **Observability:** The UI provides a "Cloud Bridge" dashboard showing active tunnels, connected runners, and a real-time audit log of all "Reach-Back" tool calls.

## 7. Evolutionary Changelog
*   **2026-03-02:** Initial Document Creation.
