# Design Doc: Marketplace Skill Verification Engine (MSVE)
**Status:** Draft
**Created:** 2026-06-28

## 1. Context and Scope
The rapid expansion of the agentic ecosystem has led to a supply chain crisis. The discovery of over 800 malicious "skills" in the ClawHub marketplace (CVE-2026-25253) proves that "one-click" tool discovery is a critical attack vector. Malicious tools can exfiltrate API keys, inject prompts, or perform unauthenticated local command execution.

The Marketplace Skill Verification Engine (MSVE) addresses this by transforming the discovery bus from a passive schema repository into a verified capability hub. Every tool sourced from a third-party marketplace must pass mandatory behavioral profiling and static analysis in an isolated "Burn-In" sandbox before being exposed to an agent's reasoning loop.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Perform automated static analysis (SEMGREP-style) on all incoming tool definitions and configuration hooks.
    *   Execute skills in an air-gapped "Ghost Shell" to profile network and filesystem side-effects.
    *   Issue hardware-attested (TPM/Secure Enclave) "Audit Manifests" that cryptographically bind a tool's SHA-256 hash to its verified security posture.
    *   Prevent the loading of any marketplace tool that lacks a valid signature from a trusted MSVE node.
*   **Non-Goals:**
    *   Verifying tools developed in-house or explicitly side-loaded by the user (these fall under local security policy).
    *   Replacing the Policy Firewall (MSVE verifies the *tool*, Policy Firewall governs the *call*).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Security Administrator
*   **Primary Goal:** Allow developers to use community MCP servers while ensuring no malicious "Rug-Pull" tools enter the engineering environment.
*   **The Happy Path (Tasks):**
    1.  A developer attempts to add a new "GitLab Helper" MCP server from a public marketplace.
    2.  MCP Any intercepts the installation and routes the tool metadata to the MSVE.
    3.  The MSVE clones the tool into a "Burn-In" sandbox and executes its `discoveryCommand`.
    4.  Static analysis detects a hidden `curl` request to an untrusted domain in a WASM-based hook.
    5.  The MSVE flags the tool as "High Risk" and prevents its registration in the Discovery Bus.
    6.  The developer receives a "Verification Failed" report, and the organization is protected from a potential exfiltration vector.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    graph TD
        A[Marketplace Tool] --> B[MSVE Interceptor]
        B --> C{Burn-In Sandbox}
        C --> D[Static Analysis - Semgrep]
        C --> E[Behavioral Profiling - Ghost Shell]
        D --> F[Audit Reporter]
        E --> F
        F --> G{Risk Score < Threshold?}
        G -- Yes --> H[Sign Audit Manifest & Register]
        G -- No --> I[Quarantine & Notify]
    ```
*   **APIs / Interfaces:**
    *   `POST /v1/msve/verify`: Submit a tool manifest for audit. Returns a tracking ID.
    *   `GET /v1/msve/report/[id]`: Retrieve the detailed behavioral and static analysis report.
    *   `POST /v1/msve/sign`: Request a hardware-attested signature for a verified manifest.
*   **Data Storage/State:**
    *   Verified SHA-256 hashes are stored in the "Attested Discovery Authority" registry.
    *   Quarantined tools and their behavioral traces are stored in an encrypted forensic vault.

## 5. Alternatives Considered
*   **Community Reputation Only:** Rejected because malicious tools can easily fake "stars" or positive reviews in open marketplaces.
*   **Manual Code Review:** Rejected due to the scale of the agentic ecosystem (800+ malicious tools were found in a single marketplace).

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The MSVE node itself must be isolated to prevent "Sandbox Escapes" during the profiling phase.
*   **Observability:** Integrated with the "Verified Skill Safety Report" in the UI to provide transparent risk scores to end-users.

## 7. Evolutionary Changelog
*   **2026-06-28:** Initial Document Creation.
