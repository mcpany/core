# Design Doc: Alias-Bound Path Validator (ABPV)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agents increasingly use high-level messaging tools and media processing servers, a new exploit pattern has emerged: **Alias-Based Sandbox Escapes** (CVE-2026-33581). In these attacks, tools implement their own higher-level URI schemes or parameter aliases (e.g., `mediaUrl`, `fileUrl`, `fetchUri`) that are not recognized by standard filesystem allowlisting. Attackers can coerce agents into providing local file paths within these alias parameters, which the underlying tool then resolves and exposes, bypassing the gateway's `localRoots` validation.

ABPV is a security middleware that provides deep, recursive inspection of tool parameters to ensure that all resolved paths—even those masked by aliases—comply with the agent's authorized filesystem boundaries.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept and recursively deconstruct tool-specific parameter aliases and custom URI schemes.
    * Enforce `localRoots` validation on the final resolved path of any alias.
    * Provide a pluggable registry for tool-specific alias resolution logic (e.g., how the OpenClaw message tool resolves `mediaUrl`).
    * Block any tool call where an alias resolves to a path outside the hardware-attested sandbox.
* **Non-Goals:**
    * Validating the content of the files themselves (this is handled by IDS/MITS).
    * Replacing existing `localRoots` checks; ABPV extends them to the parameter layer.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer
* **Primary Goal:** Prevent an agent from reading `/etc/passwd` via the `mediaUrl` parameter of a messaging tool.
* **The Happy Path (Tasks):**
    1. A compromised subagent attempts to call the `send_message` tool with `mediaUrl: "file:///etc/passwd"`.
    2. The tool call is intercepted by the ABPV middleware.
    3. ABPV identifies `mediaUrl` as a registered alias for this tool.
    4. ABPV invokes the tool-specific resolver to extract the underlying path: `/etc/passwd`.
    5. ABPV checks the path against the mission's hardware-attested `localRoots`.
    6. Since `/etc/passwd` is outside the authorized root, ABPV blocks the call and logs a security violation.
    7. The primary agent's reasoning loop is notified of the blocked exfiltration attempt.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent Tool Call] --> B[ABPV Middleware]
        B --> C{Is Alias Parameter?}
        C -- Yes --> D[Invoke Tool Resolver]
        D --> E[Extract Resolved Path]
        E --> F[Check localRoots]
        F -- Unauthorized --> G[Block & Log]
        F -- Authorized --> H[Pass to Tool Execution]
        C -- No --> H
    ```
* **APIs / Interfaces:**
    * `abpv.RegisterAlias(toolID, paramName, resolverFunc)`: Registers a parameter as an alias.
    * `abpv.ValidateCall(request) error`: Validates all parameters in a tool call.
* **Data Storage/State:**
    * **Alias Registry:** In-memory map of tool-specific alias resolution logic.
    * **Policy Store:** Integration with the `localRoots` defined in the mission manifest.

## 5. Alternatives Considered
* **Blacklisting "file://" URIs**: Rejected because attackers can use relative paths or tool-specific schemes (e.g., `internal://`) that still resolve to local files.
* **Kernel-level File Auditing**: Rejected as it is reactive (blocks after the file is opened) and may not stop high-level data leakage from tools that read files and return their content in responses. ABPV is proactive and semantic.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Closes the "Alias Gap" in filesystem sandboxing.
* **Observability:** Logs all deconstructed alias resolutions for forensic audit.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation addressing OpenClaw CVE-2026-33581.
