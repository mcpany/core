# Design Doc: Differential Context Guarding (DCG)
**Status:** Draft
**Created:** 2026-05-24

## 1. Context and Scope
With the rise of multi-agent swarms and shared teammate mailboxes (as seen in Claude Code and OpenClaw), a new vulnerability has emerged: **Context-Dump Exfiltration (CVE-2026-39102)**. A compromised subagent can trigger a massive exfiltration of the shared mailbox history by injecting malicious reasoning traces. MCP Any needs to solve this by moving from "Whole-State Sharing" to "Differential Context Sovereignty," where only the minimum necessary state is shared for any given tool call.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement real-time semantic analysis of tool outputs.
    * Ensure tool outputs only contain state fragments relevant to the specific call.
    * Block mass exfiltration attempts of the teammate mailbox/shared task list.
    * Provide cryptographic proof of context relevance for every handoff.
* **Non-Goals:**
    * Replacing the underlying transport layer (A2A/UAB).
    * Modifying the core reasoning engine of the connected agents.

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Swarm Architect
* **Primary Goal:** Securely share context between a Claude orchestrator and 5 OpenClaw specialists without risking a "Context-Dump" if one specialist is compromised.
* **The Happy Path (Tasks):**
    1. The Orchestrator initiates a task delegation via MCP Any.
    2. MCP Any's DCG layer intercepts the outgoing context fragment.
    3. DCG performs a semantic mapping of the fragment against the "Mission Root" intent.
    4. DCG redacts any mailbox history not explicitly required for the sub-task.
    5. The Specialist receives a "Sovereign Shard" of the context.
    6. Upon task completion, DCG validates the Specialist's output to ensure no "Shadow State" is being smuggled back to the parent.

## 4. Design & Architecture
* **System Flow:**
    `[Agent] -> [Mailbox Integrity Middleware] -> [DCG Semantic Filter] -> [Target Agent/Tool]`
* **APIs / Interfaces:**
    * `x-mcp-context-shard-id`: Cryptographic identifier for the Sovereign Shard.
    * `ValidateFragment(intent, fragment)`: Core logic for semantic relevance checking.
* **Data Storage/State:**
    Uses the Shared KV Store (Blackboard) with "Intent-Sealed Shards" to track fragment lineage.

## 5. Alternatives Considered
* **Manual Redaction:** Rejected due to human-in-the-loop latency and the inability to scale to machine-speed swarms.
* **Encryption-at-Rest for Mailboxes:** Rejected because the vulnerability occurs during active reasoning, where the agent must have "Cleartext" access to the context.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** DCG acts as a "Semantic Firewall," preventing prompt-driven exfiltration.
* **Observability:** Detailed logging of "Redaction Events" where DCG blocked unauthorized state fragments.

## 7. Evolutionary Changelog
* **2026-05-24:** Initial Document Creation.
