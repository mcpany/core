# Design Doc: MCP Supply Chain Integrity Guard
**Status:** Draft
**Created:** 2026-02-26

## 1. Context and Scope
The "Clinejection" attack and recent MCP discovery-based exploits have demonstrated that agents can be tricked into installing or calling malicious tools if the discovery process is unverified. As swarms grow to 100+ agents, the surface area for "Rogue Tool Injection" increases exponentially. MCP Any must enforce cryptographic provenance for all connected tools.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement mandatory cryptographic signature verification for all MCP server manifests.
    * Maintain a "Trusted Provider" registry (e.g., GitHub, Anthropic, Google) with public key pinning.
    * Provide a "Lockfile" for MCP configurations to ensure no silent updates to tool definitions.
    * Integration with the On-Demand Discovery system to only index attested tools.
* **Non-Goals:**
    * Hosting a public MCP registry (MCP Any verifies, it doesn't host).
    * Auditing the source code of every tool (only verifying the publisher's identity).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Architect.
* **Primary Goal:** Ensure that agents only discover and use tools from verified, trusted sources.
* **The Happy Path (Tasks):**
    1. Architect configures a list of `trusted_keys` in MCP Any.
    2. A discovery provider (e.g., Local Ollama or Remote HTTP) returns a list of tools.
    3. MCP Any checks for a companion `.sig` or `manifest.jwt`.
    4. If the signature is missing or invalid, the tool is excluded from the discovery index.
    5. The UI shows a "Verified" badge for attested tools and a "Warning/Blocked" state for others.

## 4. Design & Architecture
* **System Flow:**
    - **Attestation Check**: During the discovery phase (On-Demand Discovery), the Attestation Middleware intercepts the tool metadata.
    - **Signature Verification**: Uses standard Ed25519 or RSA-PSS to verify the manifest signature against the trusted keystore.
    - **Provenance Mapping**: Each tool is tagged with its provenance (e.g., `provenance: "github.com/mcp-server-git"`).
* **APIs / Interfaces:**
    - Manifest extension:
    ```json
    {
      "mcp_version": "1.0",
      "attestation": {
        "signature": "base64...",
        "signer": "anthropic-verified-tools"
      }
    }
    ```
* **Data Storage/State:** Keystore stored in `MCPANY_KEYSTORE_PATH`.

## 5. Alternatives Considered
* **Reputation-Based Discovery**: Using a star-rating system. *Rejected* as it is easily gamed and provides no cryptographic certainty.
* **Manual Approval for Every Tool**: *Rejected* due to lack of scalability in massive swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Foundational for preventing "Tool Shadowing" and "Injection" attacks.
* **Observability:** A "Provenance Dashboard" will show the distribution of tool sources.

## 7. Evolutionary Changelog
* **2026-02-26:** Initial Document Creation.
