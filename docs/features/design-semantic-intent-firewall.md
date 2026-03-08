# Design Doc: Semantic Intent Firewall
**Status:** Draft
**Created:** 2026-03-03

## 1. Context and Scope
As AI agents move from simple tool execution to autonomous multi-agent coordination (A2A), traditional method-based filtering (e.g., "allow `list_files`") is no longer sufficient. Attackers can use "A2A Contagion" to pass malicious intent through legitimate methods. The Semantic Intent Firewall (SIF) is designed to inspect the *purpose* and *payload* of agentic requests to ensure they align with high-level safety policies.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept all JSON-RPC calls before they reach the MCP server.
    * Use a lightweight "Small Language Model" (SLM) or rule-based engine to classify the intent of the request.
    * Prevent "Semantic Overreach" where an agent tries to perform actions outside its authorized task scope.
    * Provide a non-bypassable audit log of *intended* actions.
* **Non-Goals:**
    * Full-scale LLM reasoning on every request (must be high performance).
    * Replacing existing transport-layer security (mTLS/OIDC).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Enterprise AI Architect.
* **Primary Goal:** Prevent an external-facing "Support Agent" from being tricked into "exfiltrating employee data" via a legitimate `read_file` tool call.
* **The Happy Path (Tasks):**
    1. The "Support Agent" receives a poisoned prompt.
    2. The agent attempts to call `read_file(path="/etc/passwd")`.
    3. MCP Any intercepts the call.
    4. The SIF analyzes the call: "Method: read_file, Path: /etc/passwd -> Intent: System File Access".
    5. The SIF checks the Policy Engine: "Support Agent restricted to /docs/*".
    6. The SIF blocks the call and logs a "Semantic Policy Violation".

## 4. Design & Architecture
* **System Flow:**
    `Agent -> MCP Any Gateway -> Semantic Intent Firewall -> Policy Engine -> MCP Server`
* **APIs / Interfaces:**
    * New middleware hook in the MCP Any proxy chain: `OnRequest(ctx, req)`.
    * Policy definition schema (CEL or Rego) updated to include `intent` labels.
* **Data Storage/State:**
    * Intent cache (LRU) to avoid re-classifying identical payloads.
    * Persistent audit log in SQLite/PostgreSQL.

## 5. Alternatives Considered
* **Static Path Filtering:** Rejected because it cannot handle dynamic or complex semantic attacks (e.g., "write a script that later exfiltrates data").
* **Human-in-the-loop (HITL):** Too slow for autonomous swarms, though SIF can trigger HITL for "Ambiguous" intents.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The SIF itself must be isolated and have its own restricted permissions.
* **Observability:** Prometheus metrics for "Intent Latency" and "Policy Denials".

## 7. Evolutionary Changelog
* **2026-03-03:** Initial Document Creation.
