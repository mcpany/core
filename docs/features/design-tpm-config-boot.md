# Design Doc: TPM-Bound Configuration Boot
**Status:** Draft
**Created:** 2026-04-15

## 1. Context and Scope
With the rise of "Clone-and-Execute" attacks (CVE-2026-25725), where malicious repositories inject hidden hooks into project-local configuration files, standard software-based signatures are no longer sufficient. The TPM-Bound Configuration Boot (TBCB) service ensures that any agent execution environment is cryptographically bound to the host hardware. This prevents malicious configurations from being "smuggled" across machines without explicit user re-attestation on the target hardware.

## 2. Goals & Non-Goals
* **Goals:**
    * Cryptographically bind project-local `.claude/settings.json` and similar files to the local Trusted Platform Module (TPM 2.0) or Secure Enclave.
    * Generate signed "Environment Integrity Manifests" that are required for agent boot.
    * Provide a "Hardware Lock" that prevents a project configuration from being loaded if it was signed on a different machine.
* **Non-Goals:**
    * Protecting against physical access to the machine.
    * Managing non-configuration project files (e.g., source code) unless they contain executable hooks.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer
* **Primary Goal:** Safely open a new, potentially untrusted repository without risking a configuration-injection RCE.
* **The Happy Path (Tasks):**
    1. User clones a repository and opens it in their agent-enabled IDE.
    2. MCP Any detects a project-local configuration and flags it as "Hardware-Unattested."
    3. The agent is prevented from booting.
    4. User reviews the configuration hooks via the "Settings Integrity Monitor."
    5. User approves the configuration.
    6. MCP Any requests the local TPM to sign a hash of the configuration, binding it to the local hardware identity.
    7. The "Hardware-Attested Boot Manifest" is generated.
    8. The agent boots securely.

## 4. Design & Architecture
* **System Flow:**
    `[Project Config] -> (Review) -> [User Approval] -> [TPM Signer] -> (Hardware Signature) -> [Boot Manifest] -> [Agent Runtime]`
* **APIs / Interfaces:**
    * `TPMSignerService`: `BindConfigToHardware(configHash []byte) (HardwareSignature, error)`
    * `IntegrityValidator`: `VerifyBootManifest(manifest BootManifest) (bool, error)`
* **Data Storage/State:**
    * Stores the hardware-bound signatures in a local, protected state directory (not within the project repo).

## 5. Alternatives Considered
* **Standard GPG Signing:** Rejected because GPG keys can be moved between machines, failing to prevent "smuggling" of malicious configs.
* **OS-Level Permissions:** Rejected because they are often bypassed by high-privilege agents or misconfigured by users.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The hardware signature must include a nonce or timestamp to prevent replay attacks.
* **Observability:** Failed boot attempts due to hardware-mismatch are logged with high-severity alerts.

## 7. Evolutionary Changelog
* **2026-04-15:** Initial Document Creation.
