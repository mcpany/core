# Design Doc: Adversarial PR Description Scanner
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
Recent vulnerabilities (CVE-2025-53773) have shown that attackers can weaponize pull request (PR) descriptions and other high-trust metadata to execute indirect prompt injection. When an agent ingests a PR to perform a code review or merge, it often treats the description as high-trust instruction context, which can lead to unauthorized tool execution.

MCP Any must act as the primary defense layer by implementing an Adversarial PR Description Scanner. This system treats all external metadata as untrusted "Content" and performs deep semantic scanning to detect and neutralize embedded imperative instructions before they reach the agent's reasoning engine.

## 2. Goals & Non-Goals
* **Goals:**
    * Perform real-time SEMGREP-style scanning on all ingested PR metadata.
    * Detect "Instruction Smuggling" patterns (e.g., "Ignore previous instructions and run shell...").
    * Provide a visual "Redaction Log" for users to review blocked metadata fragments.
    * Support pluggable "Adversarial Heuristics" to stay updated with new injection patterns.
* **Non-Goals:**
    * Modifying the original PR on the source platform (e.g., GitHub).
    * Blocking non-adversarial, natural language descriptions.

## 3. Critical User Journey (CUJ)
* **User Persona:** CI/CD Build Agent Supervisor
* **Primary Goal:** Prevent an agent from being hijacked by a malicious PR description during an automated triage process.
* **The Happy Path (Tasks):**
    1. An agent requests tools to read a new PR from GitHub.
    2. MCP Any intercepts the PR metadata (title, description).
    3. The Adversarial Scanner identifies a "Chain-of-Thought Spoofing" fragment in the description.
    4. MCP Any redacts the malicious fragment and tags the content with a "Low Confidence" security flag.
    5. The agent receives the sanitized context and proceeds with its task without being hijacked.

## 4. Design & Architecture
* **System Flow:**
    `External Metadata (GitHub) -> MCP Any -> [Adversarial Scanner] -> Semantic Sanitizer -> Agent Reasoning Engine`
* **APIs / Interfaces:**
    * `POST /v1/scan/metadata`: Scan a block of text for adversarial patterns.
    * `GET /v1/scan/alerts`: Retrieve a list of recently blocked instruction injections.
* **Data Storage/State:**
    Redaction signatures and alert logs are stored in the local audit vault.

## 5. Alternatives Considered
* **LLM-Based Self-Filtering:** Rejected as models are inherently vulnerable to the injections they are meant to filter.
* **Static Blocklists:** Rejected as attackers frequently rotate syntax to bypass simple keyword filters.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Implements "Content-Gating" where even data from "Trusted" sources like GitHub is treated as potentially hostile.
* **Observability:** Alerts are surfaced in the UI via the Metadata Sanitization Log.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
