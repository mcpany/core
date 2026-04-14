# Design Doc: TPM-Bound Configuration Boot
**Status:** Draft
**Created:** 2026-04-14

## 1. Context and Scope
The escalation of "Configuration-as-an-Attack-Vector" (CVE-2026-25725) has demonstrated that project-local settings can be weaponized to achieve RCE or credential exfiltration. MCP Any needs a way to guarantee that an agent's environment is "Deterministic" and has not been tampered with by malicious repository commits. TPM-Bound Configuration Boot utilizes hardware security modules to sign and verify environment manifests before agent execution.

## 2. Goals & Non-Goals
* **Goals:**
    * Generate hardware-attested manifests of project-local configurations.
    * Mandate TPM-bound signatures for all executable configuration hooks.
    * Provide a "Deterministic Boot" signal to agent frameworks.
    * Neutralize "Shadow Sandbox" escapes via non-existence proofs.
* **Non-Goals:**
    * Managing the entire OS boot sequence (scope is restricted to the agent's project environment).
    * Providing encryption for all project files (focus is on configuration integrity).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer
* **Primary Goal:** Open a new repository and run an autonomous agent without risking RCE from malicious `.claude/settings.json` hooks.
* **The Happy Path (Tasks):**
    1. User opens a repository in their IDE.
    2. MCP Any intercepts the project configuration loading.
    3. MCP Any generates a manifest of all settings and hooks.
    4. User performs a one-time hardware-attested (TPM) approval of this manifest.
    5. On subsequent boots, MCP Any verifies the current environment against the TPM-signed manifest.
    6. If a mismatch or new hook is detected (e.g., after a `git pull`), the agent is blocked until re-attestation.

## 4. Design & Architecture
* **System Flow:**
    `[Local Filesystem] -> [Manifest Generator] -> [TPM Signing Service] -> [Verification Gate] -> [Agent Runtime]`
* **APIs / Interfaces:**
    * `BootValidator`: `VerifyEnvironment(manifest HardwareManifest) error`
    * `ManifestGenerator`: `CreateManifest(projectRoot string) HardwareManifest`
* **Data Storage/State:**
    * Stores signed manifests in the local TPM NVRAM or a hardware-protected SQLite sidecar.

## 5. Alternatives Considered
* **Pure Software Signing:** Rejected because software keys can be exfiltrated from the user's environment.
* **Manual Code Review:** Rejected as it doesn't scale and is prone to human error (the "44% manual review" problem).

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The signing service must have exclusive access to the hardware module.
* **Observability:** Verification failures are logged as "High" severity events in the security center.

## 7. Evolutionary Changelog
* **2026-04-14:** Initial Document Creation.
