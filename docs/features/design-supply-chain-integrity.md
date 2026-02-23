# Design Doc: Supply Chain Integrity Guard

**Status:** Draft
**Created:** 2026-02-26

## 1. Context and Scope
The recent "Clinejection" supply chain attack and the OpenClaw CVE-2026-25253 vulnerability have highlighted a critical weakness in the AI agent ecosystem: the lack of verified provenance for MCP servers and tools. Agents can be tricked into installing and executing malicious code through poisoned configurations or malicious "skills." MCP Any needs a robust way to verify the integrity and origin of every connected MCP server.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement "Attested Tooling" where MCP servers must provide a cryptographic signature.
    * Establish a "Known Good" registry of verified MCP server hashes.
    * Provide a Zero Trust execution environment where unverified tools are blocked by default.
    * Support "Provenance Verification" for both local (stdio) and remote (HTTP) servers.
* **Non-Goals:**
    * Sandboxing the actual execution of third-party binaries (this is handled by the OS/Docker).
    * Becoming a central certificate authority (will support existing PKI/GPG).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious DevSecOps Engineer.
* **Primary Goal:** Ensure that only verified, signed MCP servers can be executed by the organization's agent swarm.
* **The Happy Path (Tasks):**
    1. Administrator configures MCP Any with a list of "Trusted Public Keys."
    2. A developer attempts to add a new MCP server (e.g., `github-mcp`).
    3. MCP Any fetches the server's manifest and checks its cryptographic signature against the Trusted Public Keys.
    4. If the signature is valid and the hash matches, the server is registered.
    5. If the signature is missing or invalid, registration is blocked, and an alert is logged.

## 4. Design & Architecture
* **System Flow:**
    - **Manifest Verification**: Every MCP server must provide an `mcp-manifest.json` containing hashes of its executables/components and a digital signature.
    - **Policy Engine Hook**: The existing Policy Firewall is extended to include an `IntegrityCheck` hook that runs before any tool call or server registration.
    - **Local Attestation**: For local servers, MCP Any verifies the binary hash on disk before execution.
* **APIs / Interfaces:**
    - New configuration block: `integrity_policy: { mode: "enforce", trusted_keys: [...] }`.
* **Data Storage/State:**
    - A local "Known Good" database (SQLite) stores verified hashes and metadata.

## 5. Alternatives Considered
* **Runtime Sandboxing Only**: Relying purely on Docker/gVisor. *Rejected* because it doesn't prevent "logical" prompt injection or data exfiltration if the tool is malicious by design but "safe" for the OS.
* **Centralized Registry**: Forcing all tools to be registered on a central MCP Any hub. *Rejected* to maintain the decentralized, local-first nature of the protocol.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** This is the core of the feature. It prevents "Registry Poisoning" and "RCE via Skill Injection."
* **Observability:** Detailed audit logs for every verification success/failure.

## 7. Evolutionary Changelog
* **2026-02-26:** Initial Document Creation.
