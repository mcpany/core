# Design Doc: Intent-Attestation Protocol (IAP)
**Status:** Draft
**Created:** 2026-03-01

## 1. Context and Scope
With the rise of autonomous agents like OpenClaw and the discovery of vulnerabilities like OWASP ASI01 (Agent Goal Hijack), it is no longer sufficient to verify *what* an agent is doing; we must verify *why* it is doing it. Traditional API keys and capability tokens grant broad access that can be abused via indirect prompt injection (e.g., an agent reading a malicious email that tells it to delete files). MCP Any needs a way to cryptographically link every tool call back to a verified, human-initiated intent.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a mechanism to sign "intent-tokens" at the start of a user session.
    * Require all high-risk tool calls to include a valid, non-expired intent-token.
    * Allow agents to pass intent-tokens to subagents (context inheritance) while maintaining a chain of trust.
* **Non-Goals:**
    * Replacing existing authentication (API keys/OAuth). IAP is an additional layer.
    * Real-time monitoring of LLM "thoughts" (focus is on the resulting action).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Architect
* **Primary Goal:** Prevent an autonomous agent from executing unauthorized "destructive" actions triggered by external data.
* **The Happy Path (Tasks):**
    1. User initiates a session with the agent and defines a high-level goal (e.g., "Analyze these logs and summarize").
    2. MCP Any generates a session-bound Intent-Token signed by the user's local key.
    3. The agent receives the goal and the Intent-Token.
    4. The agent calls a "Read Log" tool, passing the token.
    5. MCP Any verifies the token matches the current session and allows the call.
    6. If a malicious log entry tries to trick the agent into "Delete All Logs", the subsequent tool call fails because the "Delete" action is not covered by the original "Summarize" intent-token scope.

## 4. Design & Architecture
* **System Flow:**
    User -> [Goal + Sign] -> Intent-Token Generator -> Agent -> [Tool Call + Token] -> IAP Middleware -> [Verify] -> Upstream Tool.
* **APIs / Interfaces:**
    * `POST /v1/intent/issue`: Issues a new token based on a signed intent string.
    * `MCP Call Header: x-mcp-intent-token`: The transport mechanism for tokens.
* **Data Storage/State:**
    * Intent-tokens are short-lived and stored in an in-memory TTL cache (e.g., Redis or internal Go map).

## 5. Alternatives Considered
* **Manual HITL for every call:** Too much friction for autonomous agents.
* **LLM-based intent verification:** Subject to the same injection attacks as the agent itself.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Tokens use Ed25519 signatures to ensure non-repudiation.
* **Observability:** Audit logs will record the `intent_id` alongside every tool call.

## 7. Evolutionary Changelog
* **2026-03-01:** Initial Document Creation.
