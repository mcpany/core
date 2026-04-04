# Design Doc: Zero-Trust MCP Proxy (ZTMP)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
Recent ecosystem audits (Knostic, 2025) have revealed that over 1,800 Model Context Protocol (MCP) servers are currently exposed to the internet without any authentication or access control. Furthermore, critical vulnerabilities in `mcp-remote` and Anthropic's `MCP Inspector` have demonstrated that connecting to an untrusted or compromised MCP server can result in full host-level Remote Code Execution (RCE).

The ZTMP is a mandatory security gateway within MCP Any that acts as a "hard shell" around these vulnerable upstreams. It enforces Zero-Trust authentication for every client request and performs deep semantic sanitization on all JSON-RPC traffic to ensure that no malicious payloads can bridge the gap from a tool server to the agent's host environment.

## 2. Goals & Non-Goals
* **Goals:**
    * Mandate authentication for all downstream MCP server connections.
    * Provide a semantic firewall that sanitizes tool inputs and outputs.
    * Neutralize OS command execution vectors (RCE) in downstream tool responses.
    * Implement a "Safe-by-Default" proxy mode for all discovered MCP servers.
* **Non-Goals:**
    * Implementing the tool logic itself (this is a security-first proxy).
    * Restricting legitimate tool usage (it focuses on malicious instruction/payload detection).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer
* **Primary Goal:** Use a community-contributed "GitHub Search" MCP server without risking an RCE from a malicious repository.
* **The Happy Path (Tasks):**
    1. The user adds the remote "GitHub Search" server URL to MCP Any.
    2. MCP Any automatically wraps the connection in a ZTMP instance.
    3. The agent calls the `search` tool; ZTMP validates the arguments against the tool schema.
    4. The remote server returns a payload containing an `mcp-remote` style command injection sequence.
    5. ZTMP's semantic parser identifies the malicious sequence and blocks the response.
    6. The agent is notified of a security interdiction; the host remains uncompromised.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        Agent[AI Agent] -->|JSON-RPC| ZTMP[ZTMP Gateway]
        ZTMP -->|Auth Check| Auth[Identity Provider]
        ZTMP -->|Scan Input| Shield[Injection Shield]
        Shield -->|Validated Req| Remote[Remote MCP Server]
        Remote -->|Response| Shield
        Shield -->|Scan Output| ZTMP
        ZTMP -->|Sanitized Res| Agent
    ```
* **APIs / Interfaces:**
    * `MCP-Proxy-Authenticate`: Custom header for hardware-attested proxy authentication.
    * `ZTMP_Policy_Engine`: Rego-based engine for defining tool-level access rules.
* **Data Storage/State:**
    * Maintains a local cache of "Safe-Server" fingerprints and hardware-attested identity tokens.

## 5. Alternatives Considered
* **Direct Connection (Status Quo):** Rejected due to the 100% RCE risk documented in CVE-2025-6514.
* **Network-Level Firewalls (e.g., Istio):** Rejected because they lack the "Semantic Awareness" required to parse MCP JSON-RPC payloads for prompt/command injection.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** ZTMP is the enforcement point for the product's Zero Trust mission. All tool traffic is considered "Hostile by Default."
* **Observability:** Blocked RCE attempts are logged with high-fidelity traces in the "Local Security Violation Monitor."

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
