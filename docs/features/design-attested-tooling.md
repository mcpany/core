# Design Doc: Attested Tooling & Verified Registries
**Status:** Draft
**Created:** 2026-03-03

## 1. Context and Scope
The "8,000 Exposed Servers" crisis and the "ClawHub" malicious skill injection incident (1,184 malicious skills) have highlighted a catastrophic vulnerability in the MCP ecosystem. Currently, agents load MCP tools from various sources (local, remote, community registries) without verifying their integrity or origin.

MCP Any needs to implement a "Verified-by-Default" layer that ensures every tool loaded into an agent's context has been cryptographically attested by a trusted authority or the user themselves.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a middleware that intercepts tool loading and verifies a cryptographic signature.
    * Support "Attested Registries" where tools are signed by community maintainers.
    * Provide a "User-Signed" mode for local custom tools.
    * Quarantining of unverified tools until explicit user approval.
* **Non-Goals:**
    * Implementing a full certificate authority (CA) infrastructure. We will use existing standards like COSE or Sigstore.
    * Dynamic analysis/sandboxing of tool execution (this is handled by the Policy Firewall).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Enterprise AI Architect
* **Primary Goal:** Ensure that no unverified or "Shadow" tools are available to the corporate agent swarm.
* **The Happy Path (Tasks):**
    1. User configures MCP Any to "Strict Attestation Mode".
    2. User adds a new tool from a verified community registry (e.g., "Verified-GitHub-MCP").
    3. MCP Any fetches the tool and verifies its signature against the registry's public key.
    4. The tool is successfully loaded and exposed to the agent.
    5. A developer tries to add a custom, unsigned tool.
    6. MCP Any flags the tool as "Unverified" and blocks its exposure to the LLM until the developer signs it with their local key.

## 4. Design & Architecture
* **System Flow:**
    1. Discovery Service identifies a new tool.
    2. Attestation Middleware intercepts the tool definition.
    3. The middleware looks for a `manifest.json.sig` file or a JWS/COSE signature header.
    4. Signature is verified against a configured "Trust Bundle" (local keys + remote registry keys).
    5. Result (Verified/Unverified/Malicious) is appended to tool metadata.
* **APIs / Interfaces:**
    * `GET /api/v1/attestation/status`: Check attestation status of all loaded tools.
    * `POST /api/v1/attestation/sign`: Sign a local tool definition.
* **Data Storage/State:**
    * Signatures are stored alongside tool definitions in the Registry Service.
    * Trust Bundle is stored in a secure, encrypted configuration file.

## 5. Alternatives Considered
* **Runtime Sandboxing Only**: Rejected because malicious tools can still exfiltrate data via prompt injection if they are allowed to stay in the context window. Prevention at discovery is superior.
* **Centralized "App Store" Model**: Rejected in favor of a decentralized "Mesh" model to maintain the open nature of MCP.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The Attestation Middleware itself must be protected from unauthorized key modification.
* **Observability:** Detailed audit logs for every verification attempt (success and failure).

## 7. Evolutionary Changelog
* **2026-03-03:** Initial Document Creation focusing on signature verification and trusted registries.
