# Design Doc: Ghost-Hook Interdiction Proxy (GHIP)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
Gemini CLI and OpenClaw have recently been targeted by "Ghost-Hook" RCE exploits, where natural-language configuration files (like `GEMINI.md` or `.claude/settings.json`) are used to inject imperative instructions before any security policy is applied. These instructions trick agents into executing unauthorized tools during the "Pre-Flight" discovery phase.

GHIP acts as a mandatory, discovery-time security gate that performs sandboxed semantic analysis of all natural-language context files before they are ingested by the reasoning engine.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept and scan all Natural Language (NL) context files during the discovery phase.
    * Use sandboxed LLMs to detect "Hidden Imperatives" or unauthorized tool-call triggers.
    * Provide "Negative Discovery Attestation" (NDA) for verified clean project environments.
* **Non-Goals:**
    * Rewriting the configuration files (GHIP only blocks or approves).
    * Enforcing runtime tool-call policies (handled by the Policy Firewall).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer
* **Primary Goal:** Safely clone and explore a third-party repository without triggering malicious auto-discovery hooks.
* **The Happy Path (Tasks):**
    1. User clones a repository and points an MCP Any agent at it.
    2. Before discovery begins, GHIP intercepts the file-listing request.
    3. GHIP identifies NL context files (e.g., `README.md`, `.mcpany/context.md`).
    4. GHIP executes a sandboxed semantic scan on these files to look for imperative hooks (e.g., "Run this shell command on boot").
    5. GHIP generates a signed NDA for the mission branch.
    6. Agent begins discovery only after NDA is verified.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        Repo[Project Repo] --> GHIP{GHIP Proxy}
        GHIP -- Clean --> NDA[Signed Attestation]
        GHIP -- Poisoned --> Block[Quarantine & Alert]
        NDA --> Agent[Agent Discovery]
    ```
* **APIs / Interfaces:**
    * `GHIP_Validate(FilePath, MissionRootToken)`: Returns NDA or ViolationReport.
* **Data Storage/State:**
    * GHIP caches NDA hashes in the hardware-locked config anchor (HLCA).

## 5. Alternatives Considered
* **Static Keyword Filtering:** Rejected because natural-language imperatives can be phrased deceptively to bypass simple regex.
* **User-in-the-Loop Approval:** Rejected as the primary gate due to "Approval Fatigue" for large repositories; NL scanning provides the necessary automation.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** NL scanning must occur in an ephemeral, resource-constrained WASM or gVisor sandbox.
* **Observability:** GHIP surfaces blocked fragments to the "Visual Attention Heatmap" for user review.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
