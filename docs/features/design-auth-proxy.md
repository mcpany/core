# Design Doc: Universal MCP Authentication Proxy

**Status:** Draft
**Created:** 2026-03-09

## 1. Context and Scope
As agents become more autonomous, they are increasingly required to interact with production-grade APIs that use complex authentication mechanisms (OAuth 2.0, AWS IAM Role Assumption, mTLS). Manually passing these credentials to every agent or subagent is insecure and technically cumbersome. Claude Code's recent updates highlight this pain point. MCP Any will solve this by acting as a secure "Auth Proxy" that decouples agent execution from upstream credential management.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Provide a centralized vault for MCP-upstream credentials (API Keys, OAuth tokens, IAM Roles).
    *   Automatically inject required authentication headers/metadata into upstream calls.
    *   Enable agents to use tools across different trust boundaries without direct access to secrets.
    *   Support ephemeral credential generation (e.g., assuming an IAM role for a specific session).
*   **Non-Goals:**
    *   Replacing enterprise identity providers (e.g., Okta, Auth0). MCP Any *integrates* with them.
    *   Storing passwords in plain text (all secrets must be encrypted at rest or referenced via env vars).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Platform Engineer at a FinTech company.
*   **Primary Goal:** Allow a support agent to fetch AWS S3 usage reports using a specific IAM role without the agent having any AWS credentials.
*   **The Happy Path (Tasks):**
    1.  Engineer configures MCP Any with a "Cloud-S3" upstream using an `aws_iam` auth type.
    2.  User triggers an agent to "get the S3 report."
    3.  Agent calls the `mcpany_s3_report` tool.
    4.  MCP Any's Auth Proxy intercepts the call, assumes the configured IAM role, and attaches the temporary `X-Amz-Security-Token` to the upstream request.
    5.  S3 returns the data; MCP Any forwards it to the agent.
    6.  The agent receives the data securely without ever "seeing" the AWS credentials.

## 4. Design & Architecture
*   **System Flow:**
    - **Auth Resolver**: Middleware that looks up the authentication configuration for the target upstream.
    - **Credential Provider**: Interface for different auth types (e.g., `OIDCProvider`, `IAMProvider`, `StaticKeyProvider`).
    - **Header Injection**: Post-processing step in the adapter layer that signs/decorates the upstream request.
*   **APIs / Interfaces:**
    ```yaml
    # Example Config
    services:
      aws_s3:
        type: http
        auth:
          type: aws_iam
          role_arn: "arn:aws:iam::123456789012:role/MCPToolRole"
          region: "us-east-1"
    ```
*   **Data Storage/State:** Secrets are stored in the local SQLite database encrypted with the instance's Ed25519 key or passed via Environment Variable injection.

## 5. Alternatives Considered
*   **Direct Agent Auth**: Let agents handle their own auth. *Rejected* due to security risks (secret leakage) and lack of centralized auditability.
*   **Env Var Pass-Through**: Just pass secrets through as variables. *Rejected* as it doesn't handle complex dynamic auth flows like OAuth refresh or IAM assumption.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The Auth Proxy is gated by the Policy Firewall. Only authorized sessions can trigger tools that use specific credentials.
*   **Observability:** Audit logs will record which agent/session used which credential for which upstream call, providing "Attributable Identity" for all tool executions.

## 7. Evolutionary Changelog
*   **2026-03-09:** Initial Document Creation.
