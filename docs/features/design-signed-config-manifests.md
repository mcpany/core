# Design Doc: Signed Config Manifests (SCM)
**Status:** Draft
**Created:** 2026-03-01

## 1. Context and Scope
Recent vulnerabilities in Claude Code (CVE-2025-59536) demonstrated that an attacker can execute arbitrary commands or exfiltrate secrets by poisoning local project configuration files (like `mcp.json` or hooks). MCP Any, as a universal gateway, must ensure that the tools and hooks it orchestrates are from a verified, trusted source.

## 2. Goals & Non-Goals
* **Goals:**
    *   Require cryptographic signatures for all configuration files (`mcp.json`, `config.yaml`, and orchestration hooks).
    *   Provide a "Trust-on-First-Use" (TOFU) or "Pre-Approved Key" model for developer environments.
    *   Block any tool or hook whose configuration hash does not match a signed manifest.
* **Non-Goals:**
    *   Signing the *binary* of the upstream tool (this is handled by OS-level code signing).
    *   Managing a global PKI (Public Key Infrastructure).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-conscious Developer
* **Primary Goal:** Open a community repository and use its MCP tools without risking machine takeover.
* **The Happy Path (Tasks):**
    1.  Developer clones a repository containing an `mcp.json` and a `.mcp-manifest.sig`.
    2.  Developer runs `mcpany run`.
    3.  MCP Any detects the manifest, verifies the signature against the Developer's trusted keys.
    4.  The tools are loaded because the signature is valid.
    5.  (Unhappy Path): If the `mcp.json` was modified by a malicious commit, the signature check fails, and MCP Any refuses to load the tools.

## 4. Design & Architecture
* **System Flow:**
    - **Manifest Generation**: A CLI tool `mcpany sign <config_file>` creates a detached signature using the user's private key (SSH, GPG, or Ed25519).
    - **Verification Hook**: Before the `ServiceRegistry` loads a new service or hook, the `ManifestVerifier` middleware checks for a corresponding `.sig` file.
    - **Trust Store**: A local database in `~/.mcpany/trust_store.db` contains fingerprints of trusted public keys.

* **APIs / Interfaces:**
    ```go
    type ManifestVerifier interface {
        Verify(configPath string, signaturePath string) error
    }
    ```

## 5. Alternatives Considered
* **User-Approval (HITL) for every tool**: Rejected due to "Approval Fatigue." Developers would eventually click "Allow" without checking.
* **Sandboxing Everything**: Difficult to implement across all OSs (Windows/Mac/Linux) with the same level of security and performance.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** If the private key is compromised, the manifest system is bypassed. We recommend hardware-backed keys (YubiKey).
* **Observability:** Failed signature checks are logged as high-severity security events in the Audit Log.

## 7. Evolutionary Changelog
* **2026-03-01:** Initial Document Creation.
