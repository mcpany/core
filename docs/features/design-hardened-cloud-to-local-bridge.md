# Design Doc: Hardened Cloud-to-Local Bridge
**Status:** Draft
**Created:** 2026-03-02

## 1. Context and Scope
CLI agents like Claude Code often run in local environments but increasingly rely on cloud-based reasoning or sandboxes. Bridging these environments securely is a major pain point. MCP Any needs a high-performance, secure tunneling mechanism that allows cloud agents to access local specialized tools without exposing the entire local network.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a secure, encrypted tunnel between cloud environments and local MCP Any instances.
    * Implement mandatory attestation for all incoming connections from the cloud.
    * Limit the scope of tools accessible via the bridge to a specific "Bridge-Allowed" list.
    * Optimize for low-latency command/response cycles typical of CLI agent interactions.
* **Non-Goals:**
    * General-purpose VPN or mesh networking.
    * Supporting legacy unencrypted protocols over the bridge.

## 3. Critical User Journey (CUJ)
* **User Persona:** CLI-First Developer
* **Primary Goal:** Use Claude Code (running in a cloud sandbox or locally with cloud-reasoning) to interact with a local proprietary database tool via MCP Any.
* **The Happy Path (Tasks):**
    1. Developer starts MCP Any locally and enables the `Cloud-Bridge`.
    2. Developer authenticates the bridge using a one-time cryptographic token.
    3. Developer configures the cloud agent (e.g., Claude Code) to use the MCP Any bridge endpoint.
    4. The cloud agent calls the local database tool; MCP Any validates the token and the tool's availability on the "Allow" list.
    5. The tool executes locally, and the result is returned securely to the cloud agent.

## 4. Design & Architecture
* **System Flow:**
    `Cloud Agent` -> `Encrypted Tunnel (WebSocket/gRPC)` -> `MCP Any Bridge Adapter` -> `Validation Middleware` -> `Local Tool`
* **APIs / Interfaces:**
    * `BridgeConnector`: Handles the secure handshake and attestation.
    * `BridgeProxy`: Routes tool calls between the tunnel and the internal dispatcher.
* **Data Storage/State:**
    * Session tokens and attestation keys stored in a secure local vault.

## 5. Alternatives Considered
* **SSH Tunneling:** Rejected due to complexity of setup for non-sysadmin users and lack of application-level (MCP) awareness.
* **Public Ngrok/Localtunnel:** Rejected due to security concerns and lack of integrated attestation.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Mandatory mTLS or token-based attestation.
* **Observability:** Monitor tunnel health, latency, and throughput in the UI.

## 7. Evolutionary Changelog
* **2026-03-02:** Initial Document Creation.
