# Design Doc: Memory-Safe Reasoning Buffers (MSRB)
**Status:** Draft
**Created:** 2026-04-14

## 1. Context and Scope
Memory Control Flow Attacks have a 90%+ success rate against current top-tier models (GPT-5, Claude 4.5). These attacks poison the agent's reasoning history to persistently hijack the workflow. MCP Any needs to provide a "Cognitive Firewall" that protects the agent's internal reasoning space.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide canary-monitored memory regions for reasoning traces.
    * Detect and interdict "Poisoned" reasoning fragments before they influence control flow.
    * Implement cryptographic isolation between different agent memory shards.
* **Non-Goals:**
    * Modifying the model's weights or internal attention mechanism.
    * General-purpose vector database storage.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Prevent an external tool output from "poisoning" the agent's reasoning to skip security checks.
* **The Happy Path (Tasks):**
    1. Agent receives output from an untrusted tool.
    2. Output is stored in a designated MSRB-protected region.
    3. MSRB's semantic canary monitor detects a "control-flow hijacking" instruction pattern in the tool output.
    4. MSRB interdicts the fragment and prevents it from being re-ingested into the model's reasoning window.
    5. User is alerted to the hijacking attempt.

## 4. Design & Architecture
* **System Flow:**
    [Tool Output] -> [MSRB Sanitizer] -> [Canary Analysis] -> [Secure Memory Commit] -> [Reasoning Window]
* **APIs / Interfaces:**
    * `PUT /v1/memory/commit`: Write to a protected reasoning buffer.
    * `GET /v1/memory/query`: Retrieve fragments for reasoning window construction.
* **Data Storage/State:**
    * Reasoning fragments are stored in cryptographically isolated memory shards.

## 5. Alternatives Considered
* **Flat Context Window:** Rejected because it allows easy manipulation of control instructions via malicious tool outputs.
* **Full Agent Virtualization:** Rejected due to prohibitive latency and compute overhead.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Zero-copy sanitization prevents TOCTOU exploits.
* **Observability:** Memory safety violation alerts are geolocated and logged.

## 7. Evolutionary Changelog
* **2026-04-14:** Initial Document Creation.
