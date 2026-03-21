# Design Doc: Content-Addressable Config (CAC) Validator
**Status:** Draft
**Created:** 2026-03-21

## 1. Context and Scope
The discovery of "Binary Smuggling" (CVE-2026-31042) highlights a critical vulnerability where AI agents ingest malicious instructions hidden in binary assets (e.g., WASM, large JSON blobs) that bypass traditional text-based security reviews. MCP Any needs a way to ensure that any configuration or executable hook loaded from a project directory has been explicitly approved by the user, regardless of its file type or location.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement hash-based validation (SHA-256) for all project-local configurations and hooks.
    * Provide a CLI tool for users to "attest" (sign-off) on specific config fragments.
    * Maintain a "Known Good" registry of attested config hashes.
    * Block the execution of any hook whose content-hash does not match the registry.
* **Non-Goals:**
    * Automatically detecting malicious intent within the config (this is handled by the Policy Engine).
    * Validating non-executable data files unless they are used as configuration sources.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer
* **Primary Goal:** Prevent an agent from executing a smuggled hook found in a recently pulled repository.
* **The Happy Path (Tasks):**
    1. The user pulls a new repository containing a `.mcpany/config.json` and a smuggled hook in a `.wasm` file.
    2. The agent attempts to call a tool that triggers the hook.
    3. The CAC Validator intercepts the request and calculates the SHA-256 hash of the hook's content.
    4. Since the hash is not in the "Attested Registry," MCP Any blocks the execution and alerts the user.
    5. The user reviews the hook, and if safe, runs `mcpany attest [hook-path]` to add the hash to the registry.

## 4. Design & Architecture
* **System Flow:**
    `Agent Request` -> `Hook Resolver` -> `CAC Validator (Hash Check)` -> `Attested Registry` -> `Execution Sandbox`
* **APIs / Interfaces:**
    * `Validator` Interface: `Validate(content []byte) (bool, error)`
    * `Registry` Interface: `Add(hash string)`, `Exists(hash string) bool`
* **Data Storage/State:**
    * Attested hashes are stored in a local, user-protected SQLite database or a signed JSON file in the user's home directory.

## 5. Alternatives Considered
* **Path-Based Allow-listing**: Rejected because files can be modified at the same path without changing the allow-list status.
* **Automated AI Review**: Rejected as it is susceptible to the same "Smuggling" and prompt injection attacks it aims to prevent.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The Attested Registry must be protected from unauthorized modifications by the agent itself.
* **Observability:** All blocked execution attempts are logged with the offending file path and hash.

## 7. Evolutionary Changelog
* **2026-03-21:** Initial Document Creation.
### Update: 2026-03-22 - Ghost Shell Behavioral Profiling
**Context:** Recent Claude Code security post-mortems confirm that project-local hooks are a primary RCE vector. Blocking them outright hinders developer velocity.
**Architecture Adjustment:**
* Introducing "Ghost Shell" mode for un-attested hooks.
* Un-attested hooks are executed in an air-gapped, instrumented container that generates a behavioral profile (file I/O, network syscalls).
**Security Impact:** Allows users to audit hook behavior without risk to the host, facilitating safer "Content-Addressable" attestation decisions.
### Update: 2026-03-21 - Adaptive Trust & Hardware-Bound Attestation
**Context:** The emergence of "Headless Handoff" friction in OpenClaw v1.6 and "Binary Smuggling" in Claude Code assets requires a transition to hardware-locked configuration integrity.
**Architecture Adjustment:**
* **Mandatory Hash Fingerprinting**: All project-local configuration blocks, including hooks and WASM-based tool definitions, now require a SHA-256 fingerprint attested in the user's local manifest.
* **RCC-Aware Discovery**: Discovery logic now verifies Resource Capability Claims (RCC) against the local hardware posture before exposing sensitive tool schemas.
**Security Impact:** Neutralizes "Binary Smuggling" exfiltration attempts and ensures that headless agents maintain a verified security posture across session boundaries.
