# Design Doc: Repository Configuration Attestation (RCA) Gateway
**Status:** Draft
**Created:** 2026-07-18

## 1. Context and Scope
The disclosure of CVE-2026-21852 highlights a critical vulnerability in AI agent environments where repository-controlled configuration files (e.g., `.claude/settings.json`) can manipulate security safeguards and exfiltrate sensitive data. Current systems often ingest these settings automatically during the "Project Discovery" phase. The RCA Gateway moves security to a "Mandatory Attestation" model, ensuring that no project-local configuration is ingested without explicit, hardware-bound user approval.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept and sandbox the ingestion of project-local configuration files.
    * Mandate hardware-attested (TPM/Secure Enclave) user signatures for any configuration overrides or hooks.
    * Provide a "Verified Baseline" for environment variables and tool routing.
    * Neutralize "Configuration-as-Execution" exploits.
* **Non-Goals:**
    * Managing tool-specific credentials (handled by Secret Store).
    * Providing a full project IDE.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer
* **Primary Goal:** Open an untrusted open-source repository and prevent it from silently hijacking their agent's API keys via `.mcpany/settings.json`.
* **The Happy Path (Tasks):**
    1. The user navigates into a new repository directory.
    2. MCP Any detects a `.mcpany/settings.json` file.
    3. The RCA Gateway intercepts the load request and puts the project in "Quarantine Mode."
    4. The user is presented with a diff of the requested configuration changes via the RCA UI.
    5. The user signs the configuration using a hardware-bound key.
    6. RCA Gateway releases the configuration to the agent runtime.

## 4. Design & Architecture
* **System Flow:**
  [Project Open] -> [RCA Interceptor] -> [Config Sandbox] -> [Attestation Dialog] -> [Runtime Ingestion]
* **APIs / Interfaces:**
    * `mcpany.rca.v1.ConfigurationService`
    * `rpc AttestConfig(AttestConfigRequest) returns (AttestConfigResponse)`
* **Data Storage/State:**
    * Persistent store of hardware-attested SHA-256 hashes of approved configuration blocks.

## 5. Alternatives Considered
* **Passive Warning Prompts**: Rejected as they are easily ignored and do not provide deterministic gating.
* **Static Analysis of Configs**: Rejected because semantic "intent" in natural-language settings is hard to verify without user context.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** RCA is the primary gatekeeper for the project-local trust boundary.
* **Observability:** Attestation events and blocked config attempts are logged in the mesh-resident audit log.

## 7. Evolutionary Changelog
* **2026-07-18:** Initial Document Creation.
