# Design Doc: Cross-Origin Connection Guard
**Status:** Draft
**Created:** 2026-03-06

## 1. Context and Scope
Recent security disclosures in the AI agent ecosystem (specifically OpenClaw) have highlighted a critical vulnerability: local AI agents listening on localhost can be hijacked by malicious websites via cross-origin requests. Since browsers automatically include cookies or rely on IP-based trust for local requests, a rogue site can "blindly" issue JSON-RPC calls to a local agent.

MCP Any, as a universal gateway, must protect the tools and state it manages from such hijacking. The "Cross-Origin Connection Guard" ensures that only authorized local processes or verified cloud bridges can communicate with the MCP Any server.

## 2. Goals & Non-Goals
* **Goals:**
    * Reject any request with an unauthorized `Origin` header.
    * Prevent "DNS Rebinding" attacks by validating the `Host` header.
    * Implement a challenge-response (AHA) for CLI-based agents to prove local ownership.
    * Support seamless bridging for authorized remote agents (e.g., via the Environment Bridging Middleware).
* **Non-Goals:**
    * Implementing a full User Authentication system (this is a transport-layer security measure).
    * Protecting against a compromised local machine (if the OS is compromised, the agent is compromised).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer using Claude Code with MCP Any.
* **Primary Goal:** Use local tools via MCP Any while browsing the web without risking tool hijacking by malicious sites.
* **The Happy Path (Tasks):**
    1. Developer starts `mcpany` server.
    2. Developer starts a CLI agent (e.g., `claude-code`).
    3. The CLI agent performs an AHA handshake (Exchange Nonce via local file or pipe).
    4. The CLI agent connects to MCP Any with the AHA token.
    5. MCP Any verifies the token and the `Origin` (or lack thereof for CLI).
    6. Tool calls are processed securely.
    7. A malicious website tries to fetch `localhost:3000/rpc`; MCP Any detects the `Origin: https://malicious.com` and drops the connection.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        participant Client as Local CLI Agent
        participant Server as MCP Any Server
        participant Browser as Malicious Website

        Note over Client, Server: AHA Handshake
        Client->>Server: Get Challenge
        Server-->>Client: Nonce (via protected local socket/file)
        Client->>Server: Connect + Nonce Hash
        Server->>Server: Validate Origin & Nonce
        Server-->>Client: Connection Accepted

        Note over Browser, Server: Hijack Attempt
        Browser->>Server: POST /rpc (Origin: malicious.com)
        Server->>Server: Check Origin Whitelist
        Server-->>Browser: 403 Forbidden
    ```
* **APIs / Interfaces:**
    * New Middleware: `OriginGuard` - intercepts all incoming HTTP/WS/gRPC connections.
    * New Internal Tool: `mcpany-aha` - helper for CLI clients to perform the handshake.
* **Data Storage/State:**
    * Volatile storage for active AHA nonces (TTL bound).
    * Configuration-bound `allowed_origins` list.

## 5. Alternatives Considered
* **Mutual TLS (mTLS):** Rejected for local dev due to certificate management complexity for end-users.
* **API Keys:** Effective, but AHA provides better "Zero Config" security for local CLI agents by leveraging filesystem permissions.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** This is a core Zero Trust component, moving from "Trusted Network" (localhost) to "Verified Identity/Origin."
* **Observability:** Audit logs will specifically flag blocked cross-origin attempts with the offending origin and target tool.

## 7. Evolutionary Changelog
* **2026-03-06:** Initial Document Creation.
