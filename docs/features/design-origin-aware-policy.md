# Design Doc: Origin-Aware Policy Engine Extension
**Status:** Draft
**Created:** 2026-03-08

## 1. Context and Scope
Recent vulnerabilities in the OpenClaw AI agent ecosystem (March 2026) have demonstrated that local agents are susceptible to side-channel attacks where a malicious website can trigger tool calls by sending requests from the browser to the agent's local RPC endpoint. As MCP Any aims to be the universal bus for all AI agents, it must move beyond simple "token-based" security to "Origin-Aware" attestation. This feature extends the existing Policy Firewall to verify not just "who" is calling but "where" the call is originating from.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement strict validation of the `Origin` and `Host` headers for all incoming MCP requests.
    * Support Process-Level Attestation (verifying the PID/Executable of the calling agent) for local Stdio/Unix socket transports.
    * Provide a configurable "Safe-Origin" allowlist in the Policy Engine.
* **Non-Goals:**
    * Implementing a full OS-level sandbox (MCP Any relies on the OS for process isolation).
    * Verifying the identity of remote users (this is handled by the existing MFA/Attestation layer).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer using a Local LLM Swarm.
* **Primary Goal:** Prevent a rogue website from executing `fs:write` tools via MCP Any while the developer is browsing.
* **The Happy Path (Tasks):**
    1. The developer starts MCP Any with the `Origin-Aware` policy enabled.
    2. MCP Any automatically detects local trusted agents (e.g., Gemini CLI, Claude Code) and assigns them "Trusted Origin" status.
    3. A malicious script in the browser attempts to POST to `localhost:8080/mcp/tool/call`.
    4. MCP Any detects the `Origin: http://malicious-site.com` header (or missing local attestation) and blocks the call.
    5. The Policy Engine logs a "Suspected Origin Hijacking" event.

## 4. Design & Architecture
* **System Flow:**
    `Agent/Browser Request` -> `Gateway Middleware` -> `Origin Attestor` -> `Policy Firewall (Rego)` -> `Tool Executor`
* **APIs / Interfaces:**
    * New `origin_policy` field in `config.yaml`.
    * Extension of the Policy Context to include `origin_metadata` (Process ID, Executable Path, Origin Header).
* **Data Storage/State:**
    * Uses a local, transient cache of "Attested PIDs" to speed up lookups for local transport.

## 5. Alternatives Considered
* **Requiring MFA for every tool call:** Rejected as it destroys the "Autonomous Agent" experience.
* **Binding to random ports:** Rejected as it only provides security through obscurity and complicates agent discovery.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** This is a core Zero Trust enhancement. It implements "Location-Based Least Privilege."
* **Observability:** Blocked requests will be tagged with detailed origin metadata in the audit log.

## 7. Evolutionary Changelog
* **2026-03-08:** Initial Document Creation.
