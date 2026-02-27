# Design Doc: Universal Policy Translator
**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
With the release of Gemini CLI v0.30.0 and Claude Code's Tool Search, the AI agent ecosystem is fragmented into different security policy formats. MCP Any needs to act as the universal bridge that allows a single security policy to be enforced across disparate agent frameworks. The Universal Policy Translator will ingest various policy formats (Gemini Policy Engine, Claude Tool Search params, Rego/CEL) and provide a unified enforcement mechanism.

## 2. Goals & Non-Goals
* **Goals:**
    * Ingest Gemini CLI `.policy` and `--policy` definitions.
    * Support Claude-style "Tool Search" metadata and filtering.
    * Provide a unified Rego-based internal representation for enforcement.
    * Enable "Single Sign-On" for security policies across swarms.
* **Non-Goals:**
    * Building a new policy language (use existing Rego/CEL).
    * Enforcing policies on non-MCP tools.

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Architect
* **Primary Goal:** Define a tool access policy once and have it enforced whether the agent is Claude Code, Gemini CLI, or an OpenClaw swarm.
* **The Happy Path (Tasks):**
    1. Architect uploads a Gemini-formatted policy to MCP Any.
    2. MCP Any translates this into its internal Rego-based Policy Firewall.
    3. A Claude-based subagent attempts to call a restricted tool.
    4. MCP Any interceptor evaluates the translated policy and denies the request, returning a standardized error that the Claude agent understands.

## 4. Design & Architecture
* **System Flow:**
    `[External Policy (Gemini/Claude)] -> [Translation Engine] -> [Unified Rego Policy] -> [Policy Firewall Middleware] -> [Tool Execution]`
* **APIs / Interfaces:**
    * `/api/v1/policies/translate`: POST endpoint to convert external formats to Rego.
    * `/api/v1/policies/active`: GET/PUT endpoint to manage the live policy set.
* **Data Storage/State:** Policies are stored in the internal SQLite DB and cached in memory for zero-latency enforcement.

## 5. Alternatives Considered
* **Manual Conversion:** Rejected as too error-prone and slow for rapidly evolving swarms.
* **Vendor-Specific Plugins:** Rejected because it leads to "Policy Silos" where security is inconsistent across agents.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The translator itself must be isolated to prevent "Policy Injection" attacks.
* **Observability:** Every translation and enforcement decision must be logged to the Audit Log with a reference to the original source policy.

## 7. Evolutionary Changelog
* **2026-02-27:** Initial Document Creation.
