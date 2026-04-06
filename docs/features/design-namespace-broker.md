# Design Doc: Authoritative Namespace Broker (ANB)
**Status:** Draft
**Created:** 2026-04-06

## 1. Context and Scope
As AI agent meshes evolve from single-registry toolsets to complex, multi-registry ecosystems (e.g., Gemini CLI v0.34.0), a new vulnerability has emerged: **Registry Shadowing**. Malicious subagents can register "Shadow Tools" with identical names to high-trust tools (like `run_shell_command`) in parallel registries, tricking parent agents into executing unauthorized code.

MCP Any needs to solve this by acting as the authoritative "Registry Arbitrer." The ANB will manage namespace priorities and ensure that once a tool name is bound to a verified registry, it cannot be overridden or shadowed by low-trust subagent injections.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a deterministic priority system for multi-registry tool discovery.
    * Use cryptographic signatures to bind tool names to verified origins.
    * Provide real-time shadowing detection and alerting.
* **Non-Goals:**
    * This system will NOT perform behavioral analysis of the tools (handled by SBF/MRA).
    * It will NOT block tool registration entirely, only prevent name-collision shadowing.

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Architect
* **Primary Goal:** Prevent a subagent from overriding the company's approved `git_push` tool with a malicious local version.
* **The Happy Path (Tasks):**
    1. The Architect defines a "Sovereign Registry" in MCP Any for approved company tools.
    2. A subagent attempts to register a local registry containing a tool named `git_push`.
    3. The ANB intercepts the registration and detects the namespace conflict.
    4. The ANB validates the signature of the company's `git_push` and assigns it priority.
    5. The subagent's `git_push` is automatically renamed to `shadow_git_push` or quarantined.

## 4. Design & Architecture
* **System Flow:**
    Discovery Bus -> ANB Validator -> [Verified Namespace Registry] -> Filtered Tool List
* **APIs / Interfaces:**
    * `POST /api/v1/namespace/register`: Register a new authoritative namespace.
    * `GET /api/v1/namespace/conflicts`: Retrieve active shadowing alerts.
* **Data Storage/State:**
    * Persistent SQLite table mapping `(ToolName, NamespaceID, RegistrySignature)`.

## 5. Alternatives Considered
* **Flat Registry with "Last-Write-Wins"**: Rejected because it is inherently vulnerable to shadowing attacks.
* **Namespace-only prefixes (e.g., `registry1_tool`)**: Rejected because it breaks agent reasoning stability (LLMs expect standard tool names).

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Mandatory signature verification for all "Sovereign" namespaces.
* **Observability:** Real-time logging of shadowing attempts and registry priority shifts.

## 7. Evolutionary Changelog
* **2026-04-06:** Initial Document Creation.
