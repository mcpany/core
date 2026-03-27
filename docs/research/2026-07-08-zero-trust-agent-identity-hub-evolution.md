# Evolution Doc: Zero-Trust Agent Identity Hub

**Date:** 2026-07-08
**Author:** Principal Software Engineer & Core Systems Lead (L7)
**Status:** Approved for Implementation

## 1. Context and Market Signals
Recent market signals indicate a significant shift towards horizontal, multi-agent swarms. In these environments, simple transport-layer security or basic tool proxying is insufficient. Agents operate autonomously and frequently spawn sub-agents, creating deep service meshes.

The primary vulnerability in these meshes is **Identity Spoofing** and unauthorized task claiming. Without a unified, cryptographically secure identity layer, a compromised agent can impersonate another specialist or escalate privileges beyond its intended scope.

## 2. Strategic Evolution: The Zero-Trust Agent Identity Hub (ZTAIH)
To secure the agent mesh, MCP Any must evolve from a simple protocol gateway to an authoritative Identity Provider (IdP) for Non-Human Identities (NHI).

We are introducing the **Zero-Trust Agent Identity Hub**. This subsystem will be responsible for:
1.  **Minting Mesh-Resident Tokens:** Issuing short-lived, cryptographically signed JWTs to connected agents upon successful authentication.
2.  **Hardware Attestation Binding:** (Future Scope) Linking tokens to specific hardware enclave IDs (TPM/SEP).
3.  **Mission-Root Lineage:** Embedding the overarching mission intent ID into the token claims, ensuring downstream actions cannot diverge from the original goal.

## 3. Core Logic & Implementation Plan
For Phase 1, we will implement the foundational `IdentityHub` in Go. This service will generate and validate agent identity tokens.

### Mermaid Flow
```mermaid
graph TD
    A[Agent Framework (e.g., OpenClaw)] -->|Connect Request| B(MCP Any Gateway)
    B --> C{Zero-Trust Identity Hub}
    C -->|Authenticate via Secret/mTLS| D[Validate Credentials]
    D --> E[Mint Mesh-Resident Token]
    E -->|Return JWT| A
    A -->|Tool Call + Token| B
    B -->|Verify Token| C
    C -->|Allow/Deny| F[Tool Execution]
```

### Technical Details (Phase 1)
*   **Location:** `server/pkg/identity/hub.go`
*   **Interface:**
    ```go
    type TokenRequest struct {
        AgentID    string
        Framework  string
        MissionRoot string
    }

    type IdentityManager interface {
        IssueToken(ctx context.Context, req TokenRequest) (string, error)
        VerifyToken(ctx context.Context, token string) (*AgentClaims, error)
    }
    ```
*   **Dependencies:** Standard JWT libraries (e.g., `github.com/golang-jwt/jwt/v5`). We must update `BUILD.bazel` to include this if not already present.

## 4. Security Considerations
*   Keys used for signing JWTs must be securely stored and rotated. For this initial implementation, we will use dynamically generated HMAC secrets or RSA keypairs held in memory.
*   Token expiration must be strictly enforced (e.g., 1-hour TTL maximum).
