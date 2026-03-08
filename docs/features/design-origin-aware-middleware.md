# Design Doc: Origin-Aware Middleware
**Status:** Draft
**Created:** 2026-03-06

## 1. Context and Scope
Recent vulnerabilities in the agent ecosystem (e.g., OpenClaw origin hijacking) have shown that local agent services are susceptible to cross-origin attacks. Malicious websites running in a user's browser can make requests to a local MCP Any instance if it is listening on a predictable port without strict origin validation. This design doc proposes a middleware that enforces "Origin Integrity" for all incoming RPC and HTTP requests.

## 2. Goals & Non-Goals
* **Goals:**
    * Strictly validate the origin of all incoming requests to MCP Any.
    * Prevent CSRF and cross-origin hijacking of local agent sessions.
    * Implement a "Local Trust" mechanism using ephemeral session tokens.
    * Support signed headers for inter-agent communication.
* **Non-Goals:**
    * Implementing a full Identity and Access Management (IAM) system.
    * Securing the transport layer itself (assumed to be TLS or local loopback).

## 3. Critical User Journey (CUJ)
* **User Persona:** Developer using local AI agents and web-based IDEs.
* **Primary Goal:** Use local tools securely without worrying about malicious websites hijacking the agent.
* **The Happy Path (Tasks):**
    1. The user starts MCP Any.
    2. MCP Any generates an ephemeral `Local-Trust-Token` and stores it in a secure local file (e.g., `~/.mcpany/session.token`).
    3. The trusted agent/client reads this token and includes it in the `X-MCP-Origin-Token` header for all requests.
    4. MCP Any's Origin-Aware Middleware intercepts the request, validates the token, and checks the `Origin`/`Referer` headers.
    5. The request is allowed to proceed to the tool execution layer.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Client->>Middleware: Request + X-MCP-Origin-Token
        Middleware->>TokenStore: Validate Token
        alt Token Valid & Origin Trusted
            TokenStore-->>Middleware: Success
            Middleware->>ToolEngine: Execute Tool
            ToolEngine-->>Client: Response
        else Token Invalid or Browser Origin detected
            TokenStore-->>Middleware: Failure
            Middleware-->>Client: 403 Forbidden (Origin Mismatch)
        end
    ```
* **APIs / Interfaces:**
    * Middleware implementation in `server/pkg/middleware/origin_integrity.go`.
    * Configuration flag: `security.strict_origin_validation: true`.
* **Data Storage/State:**
    * Ephemeral tokens stored in memory and synchronized to a local-only restricted file.

## 5. Alternatives Considered
* **IP Whitelisting:** Rejected because it doesn't prevent browser-based attacks from the same host (localhost).
* **Static API Keys:** Rejected because they are often committed to source control; ephemeral tokens are more secure for local-first workflows.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** This is a core component of the Zero Trust initiative. It ensures that the "Deputy" (MCP Any) only acts on behalf of a verified "Principal" (the trusted local agent).
* **Observability:** Failed origin checks will be logged with high severity and surfaced in the Security Dashboard.

## 7. Evolutionary Changelog
* **2026-03-06:** Initial Document Creation.
