# Design Doc: Isolated Discovery Sandbox (IDS) Middleware
**Status:** Draft | In Review | Approved
**Created:** 2026-04-07

## 1. Context and Scope
The tool discovery phase is a critical security frontier. Recent "ClawHavoc" and Claude Code "Ghost-Execution" exploits reveal that an agent merely *discovering* a tool can lead to RCE via malicious `discoveryCommand` or project-local configuration hooks. MCP Any must evolve to treat all discovery-time execution as high-risk events. The IDS Middleware provides a secure, ephemeral execution environment for these commands, neutralizing host-level escapes.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a zero-trust, isolated environment for all `tools.discoveryCommand` and associated hooks.
    * Mandate cryptographic proof of the absolute absence of unauthorized host-level side effects during the discovery phase.
    * Support all MCP, gRPC, and UACO discovery transports with integrated sandbox isolation.
* **Non-Goals:**
    * Providing long-term persistent storage for discovery logic (environment must be ephemeral).
    * Validating the *semantic* safety of the discovered tool's *outputs* (handled by the Semantic Integrity Bridge).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Agent Orchestrator
* **Primary Goal:** Discover local and remote MCP tools without risking host-level code execution from unverified registries.
* **The Happy Path (Tasks):**
    1. Orchestrator initiates a "Tool Discovery" request via the MCP Any gateway.
    2. IDS Middleware intercepts the request and identifies all required `discoveryCommand` or configuration hooks.
    3. IDS spawns an ephemeral, zero-trust sandbox (e.g., gVisor or WebAssembly-based).
    4. Discovery logic executes within the sandbox, producing tool schemas and metadata.
    5. IDS verifies the integrity of the sandbox environment and generates a Negative Discovery Attestation.
    6. Discovered tool metadata is exposed to the primary agent reasoning loop.

## 4. Design & Architecture
* **System Flow:**
    `Orchestrator -> IDS Middleware -> [Ephemeral Sandbox -> Discovery Logic] -> Tool Metadata -> Agent`
* **APIs / Interfaces:**
    * `IDS.ExecuteDiscovery(command_spec)`: Internal API for spawning the sandbox and executing discovery logic.
    * `IDS.GetAttestation()`: Returns the cryptographic proof of sandbox integrity.
* **Data Storage/State:** Discovery state is stored in a transient memory-only buffer within the sandbox and is purged immediately after tool registration.

## 5. Alternatives Considered
* **Binary Allow-listing:** Rejected because it cannot scale with the rapid emergence of new, legitimate community-driven skills.
* **Static Analysis of Discovery Scripts:** Rejected because of the high risk of obfuscation and dynamic instruction injection.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The sandbox must have no network access and no filesystem access to the host, except for explicitly mounted, read-only mission-root paths.
* **Observability:** Detailed logging of all discovery-time execution within the sandbox, including all attempted system calls and file accesses.

## 7. Evolutionary Changelog
* **2026-04-07:** Initial Document Creation.
