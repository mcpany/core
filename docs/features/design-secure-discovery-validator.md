# Design Doc: Secure Discovery CLI Validator
**Status:** Draft
**Created:** 2026-07-09

## 1. Context and Scope
The recent disclosure of CVE-2026-25593 in OpenClaw highlighted a critical vulnerability where unauthenticated local clients could set unsafe `cliPath` values during the tool discovery phase, leading to Remote Code Execution (RCE). Since MCP Any acts as the secure gateway for tool discovery across multiple frameworks, it must ensure that any executable path or command defined during discovery is rigorously validated before execution.

This feature implements a "Pre-Flight" validation layer that intercepts all discovery-time configuration changes, performing real-time semantic analysis and path allow-listing to neutralize command injection vectors.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Rigidly validate all executable paths (e.g., `cliPath`, `discoveryCommand`) against a system-level trust manifest.
    *   Perform semantic analysis on command arguments to block shell metacharacters and unauthorized flags.
    *   Provide an "Attestation Bypass" mechanism where users can explicitly sign off on known-good custom tools.
*   **Non-Goals:**
    *   Validating the runtime behavior of the tool after it has been discovered and launched (handled by other security layers).
    *   Replacing OS-level sandbox isolation.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Solo Developer using local swarms and community MCP servers.
*   **Primary Goal:** Safely discover and enable community-sourced tools without risking host-level command injection from malicious configuration files.
*   **The Happy Path (Tasks):**
    1.  The Agent framework attempts to register a new tool via a project-local configuration file containing a `cliPath`.
    2.  The Secure Discovery CLI Validator intercepts the registration request.
    3.  The Validator checks the proposed executable path against the `/etc/mcpany/trust_manifest.json`.
    4.  The Semantic Scanner parses the `discoveryCommand` for forbidden characters (e.g., `;`, `&&`, `|`).
    5.  Discovery is permitted, and the tool is added to the Discovery Bus.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    graph TD
        A[Agent/Framework] -->|Tool Discovery Request| B(Secure Discovery Validator)
        B --> C{Trust Manifest Check}
        C -->|Path Not Found| D[Block & Log Alert]
        C -->|Path Valid| E{Semantic Arg Scanner}
        E -->|Suspicious Pattern| D
        E -->|Clean| F[Expose to Discovery Bus]
    ```
*   **APIs / Interfaces:**
    *   `internal ValidateDiscoveryRequest(req DiscoveryMetadata) error`: Main entry point for the validator.
    *   `GET /security/discovery/manifest`: Returns the current list of trusted executable paths.
*   **Data Storage/State:**
    *   Read-only `trust_manifest.json` managed by the system administrator.
    *   Ephemeral cache of recently validated and signed tool definitions.

## 5. Alternatives Considered
*   **Static Path Allow-listing Only:** Rejected because it doesn't handle dynamic command arguments which can also be used for injection.
*   **Mandatory User Approval for every Tool:** Rejected due to "Approval Fatigue" which would degrade the developer experience in high-density meshes.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Follows the principle of "Implicit Distrust" for any path provided in a project-local or dynamic configuration.
*   **Observability:** All blocked discovery attempts are logged to the `Local Security Audit Dashboard` with detailed reason codes.

## 7. Evolutionary Changelog
*   **2026-07-09:** Initial Document Creation.
