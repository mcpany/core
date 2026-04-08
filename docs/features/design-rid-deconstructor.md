# Design Doc: Recursive Instruction Deconstructor (RID)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
The emergence of "Markdown-Pipeline" exploits (e.g., in Claude Code via `CLAUDE.md`) reveals a critical gap in how agents ingest natural-language configurations. Attackers can embed recursive subcommand chains that bypass traditional shell-injection filters and security gates. RID is required to perform high-fidelity, graph-based deconstruction of these configs before they are processed by the agent.

## 2. Goals & Non-Goals
* **Goals:**
    * Parse natural-language configurations (Markdown, JSON, YAML) into an execution graph.
    * Detect and block recursive subcommand pipelines (depth > 1).
    * Normalize command arguments to prevent shell metacharacter bypasses.
* **Non-Goals:**
    * Replacing the primary shell execution environment.
    * Performing general-purpose NLP for agent reasoning.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer
* **Primary Goal:** Open a community repository without triggering a recursive subcommand exfiltration pipeline.
* **The Happy Path (Tasks):**
    1. User opens a project containing a malicious `CLAUDE.md` with a hidden 50-step subcommand pipeline.
    2. RID intercepts the configuration ingestion event.
    3. RID parses the Markdown and identifies the recursive execution graph.
    4. RID flags the pipeline as high-risk and blocks the config load.
    5. User is notified of the blocked pipeline and provided with a graph visualization of the risk.

## 4. Design & Architecture
* **System Flow:**
    `Config Ingestion -> RID Parser -> Graph Analyzer -> Policy Engine -> Execution/Block`
* **APIs / Interfaces:**
    * `rid.DeconstructConfig(content, format) -> ExecutionGraph`
    * `rid.ValidateGraph(graph) -> PolicyResult`
* **Data Storage/State:**
    * Ephemeral analysis state only.

## 5. Alternatives Considered
* **Static Keyword Blocking:** Rejected as it is easily bypassed by aliasing or obfuscation. Graph analysis is required to detect the intent of the pipeline.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Operates within the Pre-Flight Trust Enclave.
* **Observability:** Integrated with the UI for visualization of blocked pipelines.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
