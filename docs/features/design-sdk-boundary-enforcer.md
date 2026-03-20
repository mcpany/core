# Design Doc: Programmatic SDK Boundary Enforcer

**Status:** Draft
**Created:** 2026-05-07

## 1. Context and Scope
The emergence of programmatic agent control via SDKs (like the OpenCode SDK v2026.1) marks a shift from chat-mediated to code-mediated agency. While this increases flexibility, it creates a new attack surface where external scripts can bypass traditional gateway security policies by injecting commands or context directly into the agent reasoning loop. The Programmatic SDK Boundary Enforcer provides mandatory security gating for all SDK-driven agent interactions, ensuring they comply with the same Zero-Trust standards as any other agent in the Universal Agent Bus.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a mandatory security gateway for all external SDK-to-agent communications.
    * Enforce cryptographically bound "SDK Intent" tokens for every programmatic call.
    * Sanitize and validate programmatic context injections against established policy manifests.
    * Provide real-time audit logs for all SDK-initiated tool calls and state mutations.
* **Non-Goals:**
    * Replacing the native SDK's internal logic or communication protocols.
    * Providing a debugger for external SDK scripts.
    * Managing the authentication of the human developer using the SDK (handled by OIDC/OAuth).

## 3. Critical User Journey (CUJ)
* **User Persona:** Platform Security Engineer
* **Primary Goal:** Ensure that a developer's automation script (using OpenCode SDK) cannot force an agent to exfiltrate secrets via a mocked tool call.
* **The Happy Path (Tasks):**
    1. A developer initiates an SDK session with MCP Any.
    2. MCP Any issues a session-bound "SDK Intent Token" after verifying the developer's credentials.
    3. The SDK script attempts to inject a "Mocked Secret Tool" into the agent's context.
    4. The SDK Boundary Enforcer intercepts the injection, identifies the unauthorized tool schema, and blocks it based on the "Zero-Trust Discovery" policy.
    5. The script is notified of the violation, and a security alert is logged in the Local Security Audit Dashboard.

## 4. Design & Architecture
* **System Flow:**
    `[SDK Script] -> [SDK Boundary Enforcer] -> [MCP Any Core Server] -> [Agent Swarm]`
* **APIs / Interfaces:**
    * `/v1/sdk/session/init`: Authenticates an SDK session and issues an Intent Token.
    * `/v1/sdk/gate/validate`: Intercepts and validates every programmatic message/call from the SDK.
    * `PolicyManifest`: A JSON/Rego manifest defining allowed SDK behaviors and injection boundaries.
* **Data Storage/State:**
    * SDK sessions and tokens are stored in the memory-mapped "Lease Registry" for sub-millisecond validation.

## 5. Alternatives Considered
* **Client-Side SDK Validation**: Rejected because it can be easily bypassed by a malicious or compromised script.
* **Manual Approval for All SDK Calls**: Rejected due to high developer friction and inability to scale with high-frequency automation.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: The Enforcer is the "Final Gate" for all programmatic agency. It prevents "SDK-to-Host" escapes by enforcing strict sandbox boundaries.
* **Observability**: All SDK interactions are visualized in the "Programmatic SDK Monitor" in the UI.

## 7. Evolutionary Changelog
* **2026-05-07:** Initial Document Creation.
### Update: 2026-05-08 - Addressing SDK-Driven Context Poisoning
**Context:** Today's market sync revealed a new exploit pattern where SDK scripts are used to inject "Silent Poison" into an agent's context fragments.
**Architecture Adjustment:**
*   Implementing mandatory semantic sanitization for all `context_injection` calls via the SDK Enforcer.
*   Introducing "Intent-Locked Context Shards" for SDK sessions, preventing scripts from modifying context outside their authorized mission branch.
**Security Impact:** Prevents malicious SDK scripts from using legitimate agents to exfiltrate data via poisoned context.
