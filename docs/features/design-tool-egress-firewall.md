# Design Doc: Tool-Specific Egress Firewall
**Status:** Draft
**Created:** 2026-03-10

## 1. Context and Scope
The "ClawHavoc" incident and the widespread discovery of SSRF vulnerabilities (36.7% of surveyed servers) in MCP implementations highlight a major architectural flaw: MCP servers typically run with the same network privileges as the host agent. A malicious or compromised MCP server can exfiltrate sensitive data or attack internal services. MCP Any must implement a mandatory network isolation layer for every tool call.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept all network requests made by an MCP server process.
    * Match requests against a declarative "Network Capability" manifest provided by the tool.
    * Block all non-whitelisted egress by default.
    * Prevent access to internal "Magic IPs" (e.g., 169.254.169.254) and localhost.
* **Non-Goals:**
    * Encrypting tool-to-target traffic (assumed to be TLS).
    * Deep Packet Inspection (DPI) of encrypted payloads (though possible if SSL-terminated, it's out of scope for V1).

## 3. Critical User Journey (CUJ)
* **User Persona:** Developer using an untrusted "Web Search" MCP server from a public registry.
* **Primary Goal:** Ensure the search tool only talks to `google.com` or `bing.com` and doesn't scan the local network.
* **The Happy Path (Tasks):**
    1. User installs the tool; MCP Any detects no `network_capabilities` in its manifest.
    2. MCP Any assigns a "No-Egress" sandbox to the tool process.
    3. Tool attempts to fetch `http://internal-db.local/secrets`.
    4. MCP Any's network namespace / eBPF filter blocks the request and logs a "Security Violation."
    5. User is notified that the tool attempted unauthorized access.

## 4. Design & Architecture
* **System Flow:**
    `MCP Server (Process)` -> `Net-Namespace / eBPF Probe` -> `Egress Proxy` -> `Internet`
    1. **Process Isolation**: Each tool process is launched in a restricted Linux network namespace.
    2. **Capability Matching**: The `Network Guard` compares requested domains/IPs against the tool's `manifest.json`.
    3. **SSRF Protection**: Hardcoded blocks for loopback (127.0.0.1), private CIDRs (10.0.0.0/8, etc.), and cloud metadata IPs.
* **APIs / Interfaces:**
    * `manifest.json` extensions: `network_capabilities: { "allowed_domains": ["api.github.com"], "allow_dns": true }`.
* **Data Storage/State:**
    * Ephemeral firewall rules managed in kernel space via `iptables` or `nftables` within the namespace.

## 5. Alternatives Considered
* **Application-level HTTP Proxy**: Rejected because it doesn't catch raw TCP/UDP exfiltration or non-HTTP protocols.
* **Cloud-based Firewalls**: Not applicable for local execution environments.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Default-Deny-All.
* **Observability**: Real-time logging of blocked requests for tool debugging.

## 7. Evolutionary Changelog
* **2026-03-10:** Initial Document Creation.
