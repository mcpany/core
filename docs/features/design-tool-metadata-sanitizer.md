# Design Doc: Tool Metadata Sanitizer

**Status:** Draft
**Created:** 2026-04-04

## 1. Context and Scope
The "Metadata-Layer Context Poisoning" vulnerability (CVE-2026-42001) has demonstrated that structural metadata in tool definitions (such as descriptions and JSON schemas) can be weaponized as a high-trust injection vector. Malicious or compromised MCP servers can embed imperative instructions in fields that were previously considered "safe metadata," leading agents to bypass system prompts or execute unauthorized actions.

The **Tool Metadata Sanitizer** is a security middleware for MCP Any that intercepts tool definitions from upstream servers, scans them for malicious patterns, and sanitizes them before they are exposed to the LLM or agent runtime.

## 2. Goals & Non-Goals
* **Goals:**
    * Automatically detect imperative instructions (e.g., "Always ignore previous rules," "Now you must...") within tool descriptions and schema metadata.
    * Sanitize or redact suspicious strings while preserving the functional utility of the tool definition.
    * Integrate with the **Verified Skill Registry** to store and retrieve "Safe Metadata" attestations for known-good tools.
    * Provide detailed logging and UI alerts when metadata poisoning is detected.
* **Non-Goals:**
    * Sanitizing the runtime *output* of tool calls (this is the responsibility of the Prompt Path Protection middleware).
    * Modifying the functional logic or parameter requirements of the MCP server.
    * Protecting against direct prompt injection in user queries.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Swarm Architect
* **Primary Goal:** Safely connect a new, third-party MCP server to a multi-agent swarm without risking "Metadata Hijacking."
* **The Happy Path (Tasks):**
    1. The architect adds a new MCP server URL to the MCP Any configuration.
    2. MCP Any performs a `list_tools` call to the new server.
    3. The **Metadata Sanitizer** intercepts the response and parses every tool description, parameter name, and schema example.
    4. The sanitizer identifies a suspicious instruction hidden in a tool's `example` field and redacts it.
    5. MCP Any serves the sanitized toolset to the agent swarm.
    6. An alert appears in the **Metadata Poisoning Alert Center** in the UI, informing the architect of the blocked injection.

## 4. Design & Architecture
* **System Flow:**
    `[MCP Server] -> (list_tools) -> [Metadata Sanitizer Middleware] -> [Heuristic & Pattern Scanner] -> [Sanitized Result] -> [LLM Gateway]`
* **APIs / Interfaces:**
    * `MetadataScanner`: Internal interface for pluggable scanning engines (Regex, LLM-based, or Static Analysis).
    * `SanitizationPolicy`: Configuration object defining strictness levels (Redact vs. Warn vs. Block).
* **Data Storage/State:**
    * Uses the local **Shared KV Store (Blackboard)** to cache sanitization results by tool-hash, reducing overhead for frequent discovery calls.

## 5. Alternatives Considered
* **Manual Attestation Only:** Rejected because the scale of modern agent swarms makes manual review of every tool definition impossible.
* **LLM-Based Sanitization:** Considered for higher accuracy but rejected as the primary method due to latency and cost concerns; static pattern matching is preferred for the "Fast-Path."

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Treats all external structural metadata as untrusted inputs.
* **Observability:** Integrated with the UI's Security Dashboard to provide real-time feedback on intercepted threats.
* **Performance:** Implements aggressive caching to ensure that sanitization does not bottleneck agent discovery loops.

## 7. Evolutionary Changelog
* **[2026-04-04]:** Initial Document Creation.
