# Design Doc: Intent-Aware Policy Engine
**Status:** Draft
**Created:** 2026-03-04

## 1. Context and Scope
As AI agents become more autonomous, the primary attack vector has shifted from direct model exploitation to "Agent Hijacking" and "Indirect Prompt Injection." In these scenarios, an agent consumes malicious data (like a website or a poisoned config file) that instructs it to perform harmful actions using its authorized tools.

Existing MCP security is mostly "Boundary-Based" (e.g., "Can this agent access the filesystem?"). MCP Any needs a more granular, "Intent-Aware" layer that inspects *why* a tool is being called before allowing execution.

## 2. Goals & Non-Goals
* **Goals:**
    * Inspect the high-level task/intent before a tool is executed.
    * Block tool calls that deviate from the verified intent.
    * Provide a standardized way for agents to "commit" to an intent before receiving tool access.
    * Support dynamic Rego/CEL policies for intent validation.
* **Non-Goals:**
    * Replace existing RBAC/Capability-based security.
    * Decipher intent from encrypted payloads.
    * Handle model-level safety (this is an execution-level guardrail).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Enterprise DevOps
* **Primary Goal:** Prevent an agent from deleting a production S3 bucket if its stated intent was "Analyze log files for errors."
* **The Happy Path (Tasks):**
    1. The parent agent initializes a session with the Intent-Aware Middleware, stating its high-level goal: "Find and summarize error logs."
    2. The Middleware issues an "Intent Token" that is cryptographically bound to this goal.
    3. The agent attempts to call `s3:delete_bucket`.
    4. The Middleware intercepts the call, analyzes the `s3:delete_bucket` action against the "Summarize error logs" intent.
    5. The Middleware identifies a mismatch and blocks the call, logging a "Policy Violation: Intent Mismatch."

## 4. Design & Architecture
* **System Flow:**
    `Agent` -> `Intent Commit (Goal: X)` -> `MCP Any (Middleware)` -> `Issue Intent Token`
    `Agent` -> `Tool Call (Action: Y) + Intent Token` -> `Policy Engine (Match X vs Y)` -> `Allow/Deny`
* **APIs / Interfaces:**
    * `intent.commit(goal_string)`: Returns an `intent_token`.
    * `tool.execute(..., intent_token)`: Mandatory for sensitive tools.
* **Data Storage/State:**
    * Intent tokens are stored in an in-memory TTL cache, bound to the session.

## 5. Alternatives Considered
* **Model-Based Verification:** Asking another LLM if the action is safe. Rejected due to latency and the risk of the "verifier" also being tricked by prompt injection.
* **Static Tool Scoping:** Only giving the agent the tools it needs. Rejected because many tools are multi-purpose (e.g., `bash` can be used to read logs OR delete files).

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The intent token must be signed by MCP Any and tied to the agent's identity to prevent token theft.
* **Observability:** Every intent-mismatch must be logged with high-fidelity traces for forensic analysis.

## 7. Evolutionary Changelog
* **2026-03-04:** Initial Document Creation.
