# Design Doc: Hardware-Attested Environment Isolated (HAEI) Provider
**Status:** Draft
**Created:** 2026-07-12

## 1. Context and Scope
Modern AI agent swarms (OpenClaw, Claude Code) frequently spawn specialist subagents to handle narrow tasks. However, today's "Process-Environment Leakage" (PEL) patterns reveal that parent process environment variables—often containing sensitive mission-root tokens, API keys, and host credentials—are implicitly inherited by child processes. This creates a massive "Shadow Privilege" surface where a compromised or rogue subagent can exfiltrate parent secrets. MCP Any needs to provide a "Zero-Leak" environment mint that physically isolates these tokens using hardware attestation.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a mandatory, hardware-attested environment scrubbing layer for all subagent spawns.
    * Ensure mission-root tokens and sensitive parent secrets are physically inaccessible to specialist child processes.
    * Issue hardware-bound (TPM/SEP) environment manifests for every sub-session.
* **Non-Goals:**
    * Replacing OS-level containerization (Docker/gVisor). HAEI acts as the credential and environment gate *before* and *during* container orchestration.
    * Managing the lifecycle of the LLM itself.

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Orchestrator (e.g., Enterprise DevOps)
* **Primary Goal:** Delegate a "File System Audit" task to a specialist subagent without exposing the primary Anthropic/OpenAI API keys or the Mission-Root administrative token.
* **The Happy Path (Tasks):**
    1. The primary agent requests a subagent spawn via MCP Any.
    2. HAEI intercepts the request and identifies the "Mission-Root" environment.
    3. HAEI performs a "Hardware-Attested Scrub," removing all non-essential environment variables.
    4. HAEI generates a new, task-bound identity token and injects only the minimal required secrets (e.g., read-only FS access token).
    5. The subagent process is spawned in an isolated environment with a hardware-signed manifest confirming the scrub.
    6. The subagent performs the task; any attempt to "probe" parent env vars returns null or a "Sovereignty Violation" signal.

## 4. Design & Architecture
* **System Flow:**
    `Primary Agent` -> `MCP Any (HAEI Interceptor)` -> `Hardware Security Module (TPM)` -> `Isolated Subagent Environment`
* **APIs / Interfaces:**
    * `POST /v1/env/mint`: Requests a new hardware-attested environment manifest.
    * `GET /v1/env/verify`: Validates the integrity of a sub-session environment against its manifest.
* **Data Storage/State:**
    * Environment manifests are stored as ephemeral, hardware-signed blobs.
    * Secret mappings are maintained in a kernel-locked memory region.

## 5. Alternatives Considered
* **Standard OS `env -i`**: Rejected because it lacks cryptographic proof of the scrub and can be bypassed by certain shell-spawn patterns.
* **Full Virtualization (VMs)**: Rejected due to 500ms+ latency tax, which is unacceptable for high-frequency agentic coordination.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** HAEI implements "Default-Deny" for all environment variables. Only variables explicitly whitelisted in the Mission-Root Manifest are propagated.
* **Observability:** Every environment "Mint" and "Scrub" event is logged to the hardware-attested audit trail.

## 7. Evolutionary Changelog
* **2026-07-12:** Initial Document Creation.
