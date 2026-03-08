# Design Doc: Supply Chain Integrity & Skill Reputation

**Status:** Draft
**Created:** 2026-03-05

## 1. Context and Scope
As the MCP ecosystem grows, agents are increasingly exposed to "Shadow Skills"—unverified or malicious MCP servers that can exfiltrate data or perform unauthorized actions. Recent incidents like "Clinejection" and "ClawJacked" demonstrate that standard tool discovery is insufficient. MCP Any needs a robust mechanism to verify the provenance of every tool and assign a reputation score based on community audits and cryptographic signatures.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement a cryptographic verification layer for all MCP server configurations.
    *   Establish a "Reputation & Verification Engine" that scores tools based on community feedback and automated security scans.
    *   Enable "Quarantine-by-Default" for any tool with a low or unknown reputation score.
    *   Support for signed tool manifests.
*   **Non-Goals:**
    *   Building a full-blown "App Store" (the focus is on security and verification, not monetization).
    *   Vetting the *logic* of every tool (focus is on provenance and known-good signatures).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Security Architect.
*   **Primary Goal:** Ensure that agents only use tools that have been signed by the internal Platform Team or highly-reputed community members.
*   **The Happy Path (Tasks):**
    1.  Architect configures MCP Any with a list of "Trusted Root Keys."
    2.  An AI agent attempts to discover a new tool from a community repository.
    3.  MCP Any intercepts the discovery request and checks the tool's cryptographic signature against the root keys.
    4.  If the signature is missing or invalid, the tool is marked as "Quarantined" in the UI and hidden from the agent.
    5.  Architect reviews the quarantined tool and can manually "Allow-list" it.

## 4. Design & Architecture
*   **System Flow:**
    - **Manifest Fetcher**: Downloads `mcp-manifest.json` which contains the tool definition and a detached JWS (JSON Web Signature).
    - **Verification Logic**: Uses the `ProvenanceEngine` to validate the JWS against local and remote (OCI-based) public keys.
    - **Reputation Aggregator**: Queries the MCP Any Global Reputation API (optional) to fetch community trust scores.
*   **APIs / Interfaces:**
    - `GET /v1/reputation/{tool_id}`: Returns trust score and audit history.
    - `POST /v1/verify`: Manually triggers verification for a tool definition.
*   **Data Storage/State:**
    - SQLite table `tool_provenance` storing signatures and trust levels.

## 5. Alternatives Considered
*   **Centralized Gatekeeper**: All tools must be approved by a central authority. *Rejected* as it violates the decentralized spirit of MCP.
*   **Manual Checklists**: Providing a PDF of "Approved Tools." *Rejected* as it cannot scale with autonomous agent discovery.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This is a core component of Zero Trust Tooling. It ensures the "Identity" of the tool is verified before the first instruction is ever sent.
*   **Observability:** The UI must clearly distinguish between "Verified," "Unverified," and "Quarantined" tools.

## 7. Evolutionary Changelog
*   **2026-03-05:** Initial Document Creation.
