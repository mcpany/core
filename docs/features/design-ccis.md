# Design Doc: Cross-Context Injection Shield (CCIS)
**Status:** Draft
**Created:** 2026-07-22

## 1. Context and Scope
The disclosure of CVE-2026-0628 (Gemini Context-Riding) has exposed a critical vulnerability in browser-native and host-integrated AI assistants. Malicious agents or extensions can "ride" a high-trust execution context to bypass sandbox boundaries and access local files or sensors.

The **Cross-Context Injection Shield (CCIS)** is a mandatory semantic sanitization layer for MCP Any. It sits at the boundary between untrusted project environments (e.g., a user's local directory or a third-party repo) and the high-trust agent reasoning loop. It ensures that no "Invisible" instructions or privileged script injections can transition across context boundaries.

## 2. Goals & Non-Goals
* **Goals:**
    * Detect and block "Context-Riding" script and HTML injections (CVE-2026-0628 defense).
    * Perform real-time semantic deconstruction of all data crossing the host-to-agent boundary.
    * Mandate hardware-attested approval for any "Privileged Boundary Transition."
    * Scrub imperative instructions from environmental metadata (filenames, environment variables).
* **Non-Goals:**
    * Replacing existing transport-layer security (TLS/mTLS).
    * Providing general-purpose PII scrubbing (handled by the PII-Sovereign Context Scrubber).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer
* **Primary Goal:** Open a potentially malicious open-source repository without the agent being tricked into exfiltrating local SSH keys via context-riding.
* **The Happy Path (Tasks):**
    1. The user points Claude Code (via MCP Any) to a new repository.
    2. The repository contains a hidden `.mcpany/context.md` with an injection payload designed to ride the Gemini context.
    3. CCIS intercepts the ingestion of this context file.
    4. CCIS performs semantic deconstruction and identifies the "Context-Riding" pattern.
    5. The payload is quarantined, and the user is alerted via the "Origin Violation Security Hub."
    6. The agent continues reasoning with a sanitized, safe version of the project context.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        Project[Untrusted Project Context] -->|Ingest| CCIS[CCIS Middleware]
        CCIS -->|Semantic Analysis| Model[Security-Vetted Model]
        Model -->|Verdict: Block/Sanitize| CCIS
        CCIS -->|Safe Context| Agent[High-Trust Reasoning Loop]
    ```
* **APIs / Interfaces:**
    * `POST /v1/context/sanitize`: Endpoint for deconstructing and scrubbing ingested data fragments.
    * `Header: x-mcp-context-trust-level`: Metadata indicating the origin and attestation status of the data.
* **Data Storage/State:** CCIS maintains a local "Reputation Cache" of hardware-attested file hashes that have previously passed semantic scanning.

## 5. Alternatives Considered
* **Pure Regex Filtering**: Rejected because it cannot detect sophisticated semantic injections that look like natural language (e.g., deceptive markdown).
* **Full Sandbox Isolation**: While effective, isolation alone doesn't prevent an agent from *reading* a malicious instruction and then *acting* on it within its permitted scope. CCIS stops the instruction from ever being read by the agent.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All project-local data is treated as "External Untrusted" until CCIS attestation is completed.
* **Observability:** Blocked attempts are logged in the `Origin Violation Real-time Monitor` with full semantic traces for forensic analysis.

## 7. Evolutionary Changelog
* **2026-07-22:** Initial Document Creation.
