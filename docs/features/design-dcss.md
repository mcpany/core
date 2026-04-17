# Design Doc: Discovery-Command Semantic Sanitizer (DCSS)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
Autonomous agent frameworks often discover tools by executing local commands or reading configuration paths (e.g., `cliPath` in OpenClaw). Recent vulnerabilities (CVE-2026-25593) prove that these discovery-time execution hooks can be weaponized for Remote Code Execution (RCE) via shell injection. MCP Any must protect the discovery phase by semantically sanitizing commands before they reach the execution engine.

## 2. Goals & Non-Goals
* **Goals:**
    * Perform real-time semantic analysis of discovery-time commands.
    * Block shell metacharacters and unauthorized flags in command arguments.
    * Execute discovery logic in an ephemeral, zero-trust sandbox.
    * Require hardware-attested user approval for new discovery command signatures.
* **Non-Goals:**
    * Sanitizing general tool calls (handled by ALSV/Policy Firewall).
    * Providing a full static analysis suite for tool binaries.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer
* **Primary Goal:** Safely import a new agent framework configuration that includes custom discovery hooks.
* **The Happy Path (Tasks):**
    1. User loads a new configuration containing a `discoveryCommand` block.
    2. DCSS intercepts the command string and performs semantic deconstruction.
    3. DCSS detects no shell-injection patterns but flags the command as "New Signature."
    4. User receives a hardware-attested approval dialog in the UI.
    5. Upon approval, DCSS executes the command in a network-isolated sandbox to retrieve tool schemas.
    6. Verified tools are registered in the Discovery Bus.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        Config[Config Loader] --> DCSS[DCSS Middleware]
        DCSS --> Parser[Semantic Parser]
        Parser --> Approval{User Approval?}
        Approval -- Yes --> Sandbox[Discovery Sandbox]
        Sandbox --> Bus[Discovery Bus]
    ```
* **APIs / Interfaces:**
    * `dcss.SanitizeCommand(cmdString) -> SanitizedCmd`
    * `dcss.RegisterSignature(cmdHash, userSignature) -> void`
* **Data Storage/State:**
    * **Signature Registry:** TPM-protected store of user-approved discovery command hashes.

## 5. Alternatives Considered
* **Manual Allow-listing:** Rejected because it doesn't scale with dynamic, agent-generated configurations.
* **Pure Sandboxing:** Rejected because sandboxes can still be escaped if the command itself triggers host-level vulnerabilities (e.g., via syscalls). Semantic sanitization provides a critical defense-in-depth layer.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Mandatory "Deny-by-Default" for any command signature not explicitly approved by the user via TPM.
* **Observability:** All blocked commands are logged to the "Local Security Violation Monitor" with full semantic traces.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
