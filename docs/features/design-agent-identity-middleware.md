# Design Doc: AID (Agent Identity) Resolver

**Status:** Draft
**Created:** 2026-03-09

## 1. Context and Scope
As agent swarms become more heterogeneous (using frameworks like OpenClaw, CrewAI, and AutoGen), the lack of a standardized identity layer has become a critical bottleneck. Agents cannot securely verify the identity of other agents requesting tools or state. OpenClaw's AID (Agent Identity) specification, based on Decentralized Identifiers (DIDs), provides a solution. MCP Any will implement an AID Resolver to act as the identity bridge for all connected agents.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Verify and resolve agent DIDs based on the OpenClaw AID v1.0 spec.
    *   Map verified AID identities to MCP Any capability-based permissions.
    *   Enable cryptographically secure delegation of tool access between agents.
    *   Provide an "Identity-Aware" middleware for all tool calls.
*   **Non-Goals:**
    *   Issuing DIDs (MCP Any is a resolver/consumer, not a DID registrar).
    *   Implementing agent logic based on identity.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security Architect for an AI Swarm.
*   **Primary Goal:** Ensure that only authorized agents from a specific swarm can access the `database_admin` toolset.
*   **The Happy Path (Tasks):**
    1.  Architect configures MCP Any to require AID verification for the `database` service.
    2.  An agent (OpenClaw-based) sends a tool call with an `x-mcp-aid` header containing its DID and a signature.
    3.  The AID Resolver middleware intercepts the call, resolves the DID, and verifies the signature using the agent's public key.
    4.  If verified, the middleware attaches the agent's identity and swarm membership to the context.
    5.  The Policy Firewall checks if this specific identity/swarm is authorized for the tool call.

## 4. Design & Architecture
*   **System Flow:**
    - **Resolution**: The `AIDResolver` uses a DID resolver driver (e.g., `did:jwk` or `did:web`) to fetch the agent's public key.
    - **Verification**: Verifies the `x-mcp-aid-signature` against the payload and the resolved public key.
    - **Context Injection**: Injects verified identity metadata into the tool execution context.
*   **APIs / Interfaces:**
    - New Middleware: `pkg/middleware/aid_resolver.go`
    - Header standard: `x-mcp-aid` (DID) and `x-mcp-aid-signature` (JWS/JWE).
*   **Data Storage/State:** Cached DID documents in the local SQLite database to reduce resolution latency.

## 5. Alternatives Considered
*   **API Keys per Agent**: Generating unique API keys for every agent. *Rejected* as it doesn't scale for dynamic swarms and lacks the decentralized trust of DIDs.
*   **Mutual TLS (mTLS)**: Using certificates for agent identity. *Rejected* due to the complexity of certificate management in ephemeral agent environments.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** AID is a foundational pillar of Zero Trust. It moves security from "Where the request came from" (IP/Network) to "Who is making the request" (Identity).
*   **Observability:** All audit logs will include the verified AID of the calling agent, enabling precise tracking of agent behavior.

## 7. Evolutionary Changelog
*   **2026-03-09:** Initial Document Creation.
