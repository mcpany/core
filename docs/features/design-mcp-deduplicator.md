# Design Doc: Smart MCP Server Deduplicator

**Status:** Draft
**Created:** 2026-03-08

## 1. Context and Scope
As the MCP ecosystem matures, users frequently encounter "Tool Collision." This happens when multiple MCP servers (e.g., a global Postgres server and a project-specific SQL explorer) provide identical or overlapping toolsets. When passed to an LLM, these duplicates waste context tokens and confuse the model's tool-selection logic, leading to hallucinations or errors. The Smart MCP Server Deduplicator aims to detect, fingerprint, and merge these redundant toolsets into a single, clean interface.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Automatically detect duplicate MCP servers based on connection parameters (Command/URL).
    *   Fingerprint tool definitions (Name, Description, Schema) to identify overlapping tools.
    *   Merge identical tools into a "Virtual Tool" that handles routing to the best available upstream.
    *   Provide a "Deduplication Report" in the UI to show suppressed or merged tools.
*   **Non-Goals:**
    *   Automatically resolving functional conflicts (e.g., two different "search" tools that return different data).
    *   Merging tools with the same name but fundamentally different schemas.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Agentic Developer with 20+ MCP servers configured.
*   **Primary Goal:** Reduce "Context Bloat" and tool selection errors caused by duplicate tool definitions.
*   **The Happy Path (Tasks):**
    1.  User adds a new MCP server that overlaps with an existing one.
    2.  MCP Any fingerprints the new tools and detects the collision.
    3.  The system merges the identical tools and logs the deduplication event.
    4.  The LLM receives a single, optimized toolset instead of redundant entries.
    5.  User views the "Deduplication Viewer" in the UI to verify which tools were merged.

## 4. Design & Architecture
*   **System Flow:**
    - **Discovery Layer**: Captures new MCP server registration.
    - **Fingerprinting Engine**: Generates a SHA-256 hash of (Tool Name + Parameter Schema + Description).
    - **Conflict Resolver**: If hashes match, it marks the tool as a "Duplicate" and maps it to the existing "Master Tool."
    - **Virtual Router**: When a merged tool is called, the router selects the healthiest/lowest-latency upstream server.
*   **APIs / Interfaces:**
    - `GET /api/v1/deduplication/status`: Returns a list of merged/suppressed tools.
    - `POST /api/v1/deduplication/resolve`: Manual override for specific tool collisions.
*   **Data Storage/State:** In-memory registry of tool fingerprints and their corresponding upstream mappings.

## 5. Alternatives Considered
*   **Manual Deduplication**: Requiring users to manually disable duplicate tools. *Rejected* as it doesn't scale and is prone to human error.
*   **Prefix-based Namespacing**: Forcing all tools to have a server-specific prefix. *Rejected* because it increases context tokens and breaks LLM "general knowledge" of tool names (e.g., `github_search` vs `local_github_search`).

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Ensure that merging tools doesn't bypass security policies. Policies must be evaluated against the *actual* upstream server being called.
*   **Observability:** The UI must clearly distinguish between "Deduplicated" and "Unique" tools to prevent developer confusion during debugging.

## 7. Evolutionary Changelog
*   **2026-03-08:** Initial Document Creation.
