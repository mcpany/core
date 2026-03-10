# Design Doc: Thinking Protocol Middleware
**Status:** Draft
**Created:** 2026-03-10

## 1. Context and Scope
With the release of models like Claude 4.6 and Gemini 2.x, "Thinking Blocks" (or chain-of-thought) have become a standard part of the model's output. These blocks provide critical reasoning context but are currently unstandardized and often stripped away by simple gateways. MCP Any needs a "Thinking Protocol Middleware" to capture, standardize, and optionally stream these reasoning chains to users and subagents. This increases transparency and allows for "Intent-Based Validation" before tools are even called.

## 2. Goals & Non-Goals
* **Goals:**
    * Capture "Thinking Blocks" from supported LLM providers (Anthropic, Google, OpenAI).
    * Standardize these blocks into a common `ThinkingContext` schema.
    * Provide a real-time WebSocket streaming endpoint for active thinking chains.
    * Allow subagents to ingest the parent agent's reasoning chain as part of their context.
* **Non-Goals:**
    * Generating thinking on behalf of models that don't support it natively.
    * Modifying the model's internal reasoning process.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local AI Developer / Power User.
* **Primary Goal:** Inspect the model's reasoning in real-time to debug a complex tool-calling loop.
* **The Happy Path (Tasks):**
    1. Agent starts a task via MCP Any.
    2. Model begins generating "Thinking" tokens.
    3. MCP Any intercepts these tokens and streams them to the UI's "Reasoning Monitor."
    4. User sees the model's logic step-by-step before it makes a (potentially destructive) tool call.
    5. User feels confident in the model's path and allows it to proceed.

## 4. Design & Architecture
* **System Flow:**
    `LLM Provider` -> `Thinking Capture Middleware (MCP Any)` -> `[UI Stream | Subagent Context]`
* **APIs / Interfaces:**
    * `GET /v1/stream/thinking/{session_id}`: WebSocket endpoint for real-time thinking tokens.
    * `ThinkingContext` Header: Standardized header for passing reasoning to subagents.
* **Data Storage/State:**
    * Thinking chains are stored in the `Session Blackboard` (SQLite) for historical audit and cross-agent inheritance.

## 5. Alternatives Considered
* **Stripping Thinking Blocks**: Simple but loses valuable context and debugging information.
* **Model-Specific Parsing in UI**: Hard to maintain as every model uses different delimiters (`<thinking>`, `thought:`, etc.).

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Thinking blocks may contain sensitive plan data. Access to the stream is restricted to the session owner.
* **Performance**: Low-latency token proxying is critical to ensure the UI doesn't lag behind the model's generation.

## 7. Evolutionary Changelog
* **2026-03-10:** Initial Document Creation.
