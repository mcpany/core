# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

# Design Doc: A2A Trust & Identity Mesh

**Status:** Draft
**Created:** 2026-03-08

## 1. Context and Scope
As agent swarms transition from single-provider to multi-vendor (e.g., an OpenClaw agent delegating to a specialized CrewAI agent), the lack of a standardized identity layer becomes a critical vulnerability. Currently, agents trust any incoming message on the A2A bus. MCP Any needs to provide a "Verifiable Agency" layer where agent identities are backed by Decentralized Identifiers (DIDs) and every message is cryptographically signed.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Integrate a DID Resolver middleware to verify agent identities.
    *   Implement JWS (JSON Web Signature) validation for all A2A messages.
    *   Establish a "Trust Registry" where users can authorize specific Agent DIDs to access certain tools.
    *   Ensure non-repudiation of agent actions across the mesh.
*   **Non-Goals:**
    *   Acting as a DID Method provider (we will resolve existing DIDs, not issue them).
    *   Managing human-to-agent identity (focus is strictly Agent-to-Agent).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Multi-Agent Swarm Orchestrator.
*   **Primary Goal:** Delegate a sensitive financial task from a "Generalist Agent" to a "Verified Accountant Agent" without risk of impersonation.
*   **The Happy Path (Tasks):**
    1.  Orchestrator configures the Accountant Agent with a DID (e.g., `did:key:z6Mkha...`).
    2.  Generalist Agent sends an A2A task request via MCP Any.
    3.  MCP Any intercepts the request, resolves the DID, and verifies the JWS signature.
    4.  The request is allowed only if the DID matches the authorized entry in the "A2A Trust Registry."

## 4. Design & Architecture
*   **System Flow:**
    - **A2A Middleware**: Intercepts `mcp_a2a_send` calls.
    - **DID Resolver Service**: Resolves the public key associated with the `sender_did`.
    - **Signature Verifier**: Validates the payload against the resolved public key.
    - **Policy Enforcement Point (PEP)**: Checks the Trust Registry for authorization.
*   **APIs / Interfaces:**
    - `POST /v1/a2a/verify`: Internal endpoint for signature validation.
    - `mcp_a2a_trust_registry`: New tool for managing authorized DIDs.
*   **Data Storage/State:**
    - `trust_registry.json`: Persistent store of authorized DIDs and their capability scopes.

## 5. Alternatives Considered
*   **Shared Secrets (API Keys)**: Easier to implement but doesn't scale for dynamic swarms and lacks non-repudiation.
*   **Centralized OAuth2**: Requires a central authority, which contradicts the decentralized nature of agent swarms.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This is the core of "Zero Trust Agency." It prevents "Shadow Agents" from joining the mesh.
*   **Observability:** All failed signature validations are logged with high severity. The UI will show a "Verified" badge next to tool calls from authenticated agents.

## 7. Evolutionary Changelog
*   **2026-03-08:** Initial Document Creation.
