# Design Doc: Attested Discovery Authority
**Status:** Draft
**Created:** 2026-04-05

## 1. Context and Scope
As AI agents become increasingly autonomous, they rely on tool discovery to expand their capabilities dynamically. However, the discovery phase is now a primary attack vector. The disclosure of **CVE-2026-33574** in OpenClaw revealed that lexical path validation is insufficient due to Time-of-Check Time-of-Use (TOCTOU) race conditions. An attacker can rebind a validated tools-root path to a malicious location before the actual write/execution occurs.

MCP Any must act as the authoritative "Certificate Authority" for tool discovery, ensuring not just the identity of the tool, but the physical integrity of its location on the filesystem.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide cryptographic proof of tool provenance before exposure to the agent.
    * Implement **Hardware-Locked Inode Pinning (HLIP)** to neutralize TOCTOU race conditions.
    * Mandate sandboxed execution for all discovery-time commands (e.g., `discoveryCommand`).
    * Generate "Environment Integrity Manifests" for every discovery session.
* **Non-Goals:**
    * Managing the entire lifecycle of the tool post-discovery (this is handled by the execution middleware).
    * Providing a marketplace for tools (this is a separate "Registry" concern).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Securely add a new gRPC-based testing tool to an active agent session without risking host-level RCE.
* **The Happy Path (Tasks):**
    1. The user points MCP Any to a new tool configuration.
    2. The Attested Discovery Authority intercepts the discovery request.
    3. It resolves the tool's path and immediately pins the underlying filesystem Inode to the session.
    4. Any discovery-time commands are executed in an ephemeral, network-isolated sandbox.
    5. The Authority generates a SHA-256 hash of the tool's structural metadata and signs it.
    6. The tool is exposed to the agent only after the hardware-attested manifest is verified.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Discovery Request] --> B[Path Resolver]
        B --> C{HLIP Middleware}
        C -->|Pin Inode| D[Discovery Sandbox]
        D -->|Execute Hook| E[Metadata Extractor]
        E -->|Sign| F[Discovery Authority]
        F -->|Verified Manifest| G[Agent Capability Bus]
    ```
* **APIs / Interfaces:**
    * `discovery.Attest(configPath) -> AttestationReceipt`: Pins the path and returns a cryptographic receipt.
    * `discovery.Verify(receipt) -> bool`: Validates that the Inode has not changed since attestation.
* **Data Storage/State:**
    * **Inode Registry:** In-memory map of session-bound Inodes and their verified hashes.

## 5. Alternatives Considered
* **Lexical Path Normalization:** Rejected because it does not protect against symlink-switching or path-rebinding between validation and execution.
* **Full Docker Isolation:** While secure, the overhead for simple tool discovery is too high for local CLI-driven agents. HLIP provides kernel-level security with sub-millisecond latency.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Mandatory HLIP ensures that "Configuration-as-Execution" cannot be weaponized.
* **Observability:** All discovery failures and TOCTOU violation attempts are logged with high-priority alerts in the UI.

## 7. Evolutionary Changelog
* **2026-04-05:** Initial Document Creation focusing on TOCTOU mitigation.
