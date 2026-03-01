# Design Doc: Supply Chain Integrity Guard
**Status:** Draft | In Review | Approved
**Created:** 2026-03-01

## 1. Context and Scope
The "ClawHavoc" and "Clinejection" attacks have demonstrated that the greatest threat to autonomous agents is no longer just the model output, but the integrity of the tools they execute. A single malicious MCP server or skill can compromise the entire host system. MCP Any must provide a "Zero-Trust" perimeter that verifies the provenance of every tool before it is loaded or executed.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement cryptographic signature verification for all MCP servers.
    * Provide a local "Attestation Service" to check for malware signatures in tool packages.
    * Establish a "Low-Trust Sandbox" for unverified tools.
    * Automate the quarantine of any tool from an unverified or "Shadow" source.
* **Non-Goals:**
    * Building a full antivirus solution.
    * Replacing the need for manual user review of high-risk actions (handled by HITL Gateway).
    * Verifying the "intent" of the LLM itself (handled by Policy Firewall).

## 3. Critical User Journey (CUJ)
* **User Persona:** Corporate Security Officer / Individual Developer.
* **Primary Goal:** Prevent the execution of malicious "Shadow" tools in their agent ecosystem.
* **The Happy Path (Tasks):**
    1. User adds a new MCP server from a community registry.
    2. MCP Any automatically intercepts the registration and checks for a valid cryptographic signature.
    3. The Attestation Service scans the server's executable and configuration for known malicious patterns (e.g., AMOS).
    4. If the tool is unverified, it is automatically assigned to the "Low-Trust Sandbox" with restricted permissions.
    5. The user is notified via the Security Dashboard and can choose to "Promote" the tool after manual review.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[MCP Server Registration] --> B{Attestation Service}
        B -- Verified Signature --> C[High-Trust Registry]
        B -- No Signature --> D[Low-Trust Sandbox]
        B -- Malware Detected --> E[Quarantine]
        C --> F[Policy Firewall]
        D --> F
        F --> G[Tool Execution]
    ```
* **APIs / Interfaces:**
    * `/api/v1/attest`: Endpoint for submitting a tool package for attestation.
    * `/api/v1/quarantine`: List and manage quarantined tools.
* **Data Storage/State:**
    * `attestation.db`: Local SQLite store for verification status and cryptographic fingerprints.

## 5. Alternatives Considered
* **Manual Whitelisting Only**: Rejected due to the high friction and inability to scale with thousands of community tools.
* **Network Isolation Only**: Rejected because some malicious tools can still exfiltrate data via the agent's own communication channel if not sandboxed at the OS level.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All tools are considered "Guilty until proven innocent." The sandbox is the default execution environment.
* **Observability:** All attestation failures and quarantine events are logged to the Security Dashboard with detailed provenance metadata.

## 7. Evolutionary Changelog
* **2026-03-01:** Initial Document Creation.
