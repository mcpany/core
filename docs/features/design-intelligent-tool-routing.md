# Design Doc: Intelligent Tool Routing (Intent-Aware Discovery)
**Status:** Draft
**Created:** 2026-03-09

## 1. Context and Scope
As the number of tools available to agents grows, even "Lazy Loading" via keyword search can lead to irrelevant tool selection and context bloat. Modern agents (Claude Code, Gemini CLI) are shifting towards a more semantic, intent-based approach to tool discovery. MCP Any needs an "Intelligent Tool Router" that acts as a cognitive pre-filter for tool availability.

## 2. Goals & Non-Goals
* **Goals:**
    * Minimize context usage by only activating tools that align with a high-level intent.
    * Improve tool selection accuracy by reducing the search space for the primary LLM.
    * Support dynamic activation/deactivation of tools based on session progress.
* **Non-Goals:**
    * Executing tools (handled by the core gateway).
    * Training new models (will use existing LLMs for intent analysis).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Orchestrator
* **Primary Goal:** Provide a specialized "Research Agent" with 50+ possible tools, but only load the 3-5 relevant to "biomedical data extraction."
* **The Happy Path (Tasks):**
    1. The agent sends a query to the MCP Any Gateway.
    2. The Intelligent Tool Router intercepts the query.
    3. The Router performs a fast intent analysis (using a small, local model or cached results).
    4. The Router identifies the "Biomedical" intent.
    5. The Router activates only tools tagged or indexed for "Biomedical" research.
    6. The agent receives a truncated tool list, improving its reasoning efficiency.

## 4. Design & Architecture
* **System Flow:**
    `[Agent] -> [MCP Any Gateway] -> [Intent Router Middleware] -> [Semantic Index] -> [Filtered Tool List] -> [Primary LLM]`
* **APIs / Interfaces:**
    * `POST /intent/analyze`: Internal endpoint for mapping queries to tool categories.
    * `GET /tools?intent=...`: Filtered tool retrieval.
* **Data Storage/State:**
    * Vector database (e.g., ChromaDB or local FAISS) for tool description embeddings.
    * Intent-to-Tool mapping table in the core SQLite database.

## 5. Alternatives Considered
* **Keyword-only Search**: Rejected because it fails on semantic nuance (e.g., "list" could mean files, users, or database entries).
* **Upfront Categorization**: Rejected because user intents are dynamic and often cross categories.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The router must respect the "Policy Firewall." An intent-based match does not override permission-based denial.
* **Observability:** Log intent matching scores and "misses" where an agent requested a tool that was filtered out.

## 7. Evolutionary Changelog
* **2026-03-09:** Initial Document Creation.
