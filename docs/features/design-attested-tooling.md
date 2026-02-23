# Design Doc: Attested Tooling Pipeline

**Status:** Draft
**Created:** 2026-02-26

## 1. Context and Scope
With the rise of "Agentic Spear-Phishing" and supply chain attacks like "Clinejection," simple configuration-based tool registration is no longer sufficient. MCP Any must ensure that every tool it exposes comes from a verified source and has not been tampered with. This document outlines a pipeline for cryptographic attestation of MCP servers.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Verify the identity of MCP servers via cryptographic signatures (e.g., Sigstore/Cosign).
    *   Implement a "Tool Manifest" that includes checksums for all upstream assets.
    *   Support "Attestation-Required" profiles where unverified tools are blocked.
    *   Provide an audit log of all attestation checks.
*   **Non-Goals:**
    *   Developing a new PKI (Public Key Infrastructure).
    *   Scanning tool source code for vulnerabilities (this is a runtime integrity check).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security-Conscious CTO.
*   **Primary Goal:** Prevent unauthorized "shadow" MCP servers from being connected to the corporate agent gateway.
*   **The Happy Path (Tasks):**
    1.  Admin configures MCP Any with a list of trusted public keys/OIDC providers.
    2.  A developer tries to register a new MCP server.
    3.  MCP Any requests an attestation bundle from the new server.
    4.  The server provides a signed manifest.
    5.  MCP Any verifies the signature and manifest checksums.
    6.  The tool is successfully registered and available to agents.

## 4. Design & Architecture
*   **System Flow:**
    - **Registration Hook**: The Service Registry intercepts all registration requests.
    - **Attestation Challenge**: MCP Any issues a challenge to the upstream (if it's a dynamic registration) or checks local manifest files.
    - **Verification Engine**: Uses a pluggable verification layer to check signatures against trusted roots.
*   **APIs / Interfaces:**
    - `POST /v1/registry/attest`: Endpoint for dynamic tool attestation.
*   **Data Storage/State:** Trusted roots and attestation results stored in `MCPANY_DB_PATH`.

## 5. Alternatives Considered
*   **Manual Code Review**: Too slow and doesn't scale.
*   **IP Whitelisting**: Rejected because it doesn't prevent internal tampering or "spear-phishing" via authorized IPs.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The attestation itself must be tamper-proof and short-lived.
*   **Observability:** All failed attestations must trigger high-priority alerts in the Monitoring dashboard.

## 7. Evolutionary Changelog
*   **2026-02-26:** Initial Document Creation.
