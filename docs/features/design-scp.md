# Design Doc: Survivability Certification Provider (SCP)
**Status:** Draft
**Created:** 2026-07-03

## 1. Context and Scope
As AI agents move from single-task tools to autonomous swarm deployments, the industry lacks a standardized way to verify their resilience against adversarial manipulation. Traditional security focuses on the "pipe" (transport), but modern threats like prompt injection and logic bombs target the agent's "reasoning path."

MCP Any needs to solve this by becoming the authoritative **Survivability Certification Provider (SCP)**. By implementing the Agentic Survivability Standard (ASS), MCP Any can provide hardware-attested proof that a swarm is "safe to deploy," neutralizing the trust gap that currently prevents wide-scale enterprise adoption of autonomous swarms.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement the ASS v1.0 scoring system (0-1000) natively.
    * Provide hardware-attested (TPM/Secure Enclave) signing for survivability certificates.
    * Automate adversarial probing via integration with the Adversarial Intent Prober (AIP).
    * Maintain a registry of certified "Known Good" agent frameworks and versions.
* **Non-Goals:**
    * Replacing general-purpose LLM benchmarking (e.g., HELM, MMLU).
    * Providing real-time interdiction (handled by the Attention-Density Firewall).
    * Auditing non-agentic MCP tools.

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise SecOps Architect
* **Primary Goal:** Certify a heterogeneous swarm (Claude Code + OpenClaw) for mission-critical production use without manual red-teaming.
* **The Happy Path (Tasks):**
    1. The Architect registers the target swarm components in the MCP Any SCP Dashboard.
    2. The Architect initiates a "Sovereign Audit" for the specific mission root intent.
    3. MCP Any triggers the **Adversarial Intent Prober (AIP)** to execute 25 standard adversarial probes (Injection, Exfiltration, Escalation).
    4. The SCP Hub aggregates the binary results and calculates an ASS Score.
    5. The SCP Hub requests a hardware signature from the host TPM, binding the score to the current mission-root.
    6. The Architect receives a cryptographically signed "Survivability Certificate" and white-lists the swarm for production deployment.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
      A[Architect] -->|Initiate Audit| B(SCP Hub)
      B -->|Trigger Probes| C(Adversarial Intent Prober)
      C -->|Simulated Attacks| D[Target Swarm]
      D -->|Reasoning Traces| C
      C -->|Vulnerability Report| B
      B -->|Score & Manifest| E(TPM / Secure Enclave)
      E -->|Signed Certificate| B
      B -->|Verified Cert| A
    ```
* **APIs / Interfaces:**
    * `POST /v1/audit/initiate`: Starts a new audit session for a swarm.
    * `GET /v1/audit/status/{session_id}`: Polls the status of ongoing adversarial probes.
    * `GET /v1/audit/certificate/{session_id}`: Retrieves the TPM-signed ASS certificate.
* **Data Storage/State:**
    * Audit history is stored in an encrypted SQLite sidecar.
    * Hardware keys are managed exclusively via the `HAFP` (Hardware-Accelerated Fast-Path) provider.

## 5. Alternatives Considered
* **Passive Monitoring Only:** Rejected because it allows "Dormant Logic Bombs" to execute in production before detection.
* **Manual Red-Teaming:** Rejected as it creates a human bottleneck that cannot scale with machine-speed swarm deployments.
* **Cloud-Only Certification:** Rejected to maintain **Local Sovereignty** and ensure sensitive mission constraints never leave the host enclave.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The SCP Hub itself operates in a "Detached Sandbox" to ensure that even a compromised Auditor agent cannot escalate to the host.
* **Observability:** Every probe and reasoning trace generated during the audit is logged with a monotonic timestamp and cryptographically linked to the final certificate.

## 7. Evolutionary Changelog
* **2026-07-03:** Initial Document Creation.
