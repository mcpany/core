# Design Doc: Semantic Intent Validator
**Status:** Draft
**Created:** 2026-03-11

## 1. Context and Scope
As agent ecosystems evolve from simple human-to-agent interactions to complex multi-agent swarms (Agent-to-Agent or A2A), a new class of vulnerability has emerged: "A2A Contagion." This occurs when a compromised or malicious agent propagates its intent laterally to other agents through shared state, tool calls, or handoff messages. Traditional capability-based security (tokens/scopes) is insufficient because the *action* might be authorized, but the *intent* behind it is malicious.

MCP Any needs a "Semantic Intent Validator" to act as a deep-packet inspection (DPI) layer for agentic communication, verifying that the semantic meaning of a request aligns with the established mission profile.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept and analyze the "intent" of A2A messages and tool call sequences.
    * Use local, high-speed LLMs or heuristic classifiers to score intent deviation.
    * Provide a "Semantic Firewall" that can block or flag suspicious lateral movements.
    * Support "Mission-Profile" definitions that describe allowed semantic boundaries.
* **Non-Goals:**
    * Replacing traditional RBAC/ABAC (this is an additional layer).
    * General-purpose PII scrubbing (handled by DLP middleware).
    * Guaranteed prevention of all "vibe-based" attacks (it is a probabilistic defense).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-conscious MAS (Multi-Agent System) Orchestrator.
* **Primary Goal:** Prevent a compromised "Customer Support Agent" from tricking the "Accounting Agent" into issuing an unauthorized refund via a semantic payload.
* **The Happy Path (Tasks):**
    1. Orchestrator defines a Mission Profile for the swarm (e.g., "Support agents can only query invoices, never modify them").
    2. Support Agent sends a request to Accounting Agent via MCP Any A2A Bridge.
    3. Semantic Intent Validator intercepts the payload.
    4. Validator compares the intent against the Mission Profile and historical session context.
    5. Request is approved and forwarded.

## 4. Design & Architecture
* **System Flow:**
    `Agent A -> A2A Bridge -> [Intent Validator] -> Policy Engine -> Agent B`
* **APIs / Interfaces:**
    * `IntentService.Validate(context, payload, profile) -> Score`
    * `MissionProfile` schema: JSON-based description of semantic constraints.
* **Data Storage/State:**
    * Uses the `Shared KV Store (Blackboard)` to maintain a "Semantic Trace" of the current session to detect multi-step intent shifts.

## 5. Alternatives Considered
* **Strict Schema Enforcement:** Rejected because malicious intent can be hidden in valid schemas.
* **Manual HITL for every A2A call:** Rejected as it breaks autonomous scaling.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The validator itself runs in an isolated sandbox.
* **Observability:** Logs "Intent Deviation Scores" to the Audit Log for forensic analysis.
* **Performance:** Uses quantized local models (e.g., TinyLlama or specialized BERT) to minimize latency impact.

## 7. Evolutionary Changelog
* **2026-03-11:** Initial Document Creation.
