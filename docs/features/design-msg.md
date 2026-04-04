# Design Doc: Metadata Sanitization Gateway (MSG)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
The "Tool Poisoning" exploit in natural-language metadata reveals that agents are vulnerable to instructions hidden in non-executable text. MSG acts as a semantic filter that "scrubs" external metadata before it is presented to the agent.

## 2. Goals & Non-Goals
* **Goals:**
    * Detect and redact imperative language from tool schemas and external context.
    * Normalize metadata to remove instruction fragments.
* **Non-Goals:**
    * Blocking entire tools based on safe descriptions.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Developer
* **Primary Goal:** Safely use community MCP servers.
* **The Happy Path (Tasks):**
    1. MSG intercepts tool registration.
    2. MSG redacts imperative fragments from descriptions.
    3. Agent uses the tool safely.

## 4. Design & Architecture
* **System Flow:** External Source -> MSG Middleware (Semantic Scanner) -> Sanitized Metadata -> Agent.
* **APIs / Interfaces:** Internal `pkg/msg` Go package.
* **Data Storage/State:** Cache of sanitized signatures.

## 5. Alternatives Considered
* **Static Keyword Blocking:** Too easily bypassed.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Uses air-gapped SLM for scanning.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
