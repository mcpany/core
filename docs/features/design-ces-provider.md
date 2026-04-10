# Design Doc: Contextual Entanglement Scorer (CES)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
In high-density Agent Teams, subagents often inject context that is semantically similar to mission-root instructions but originates from unverified external sources (e.g., GitHub issue comments). This "Entanglement Ghosting" can trick the model into prioritizing subagent drift over parent goals. CES provides the infrastructure to score this risk and gate attention.

## 2. Goals & Non-Goals
* **Goals:**
    * Perform real-time semantic correlation analysis between subagent state and mission-root anchors.
    * Issue "Entanglement Risk Scores" for sharded context fragments.
    * Trigger automated "Attention Gating" when scores exceed a safety threshold.
* **Non-Goals:**
    * Replacing the base LLM reasoning engine.
    * Providing general-purpose text summarization.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local Agent Team Lead
* **Primary Goal:** Prevent a specialist "Issue Reader" agent from hijacking the mission intent via a malicious MD file.
* **The Happy Path (Tasks):**
    1. Specialist agent reads an external file and writes its content to the shared scratchpad.
    2. CES interceptor analyzes the write request.
    3. CES compares the new fragment against the "Mission-Root Gravity" anchors.
    4. CES detects high-confidence instruction mimicry (Score: 0.92).
    5. CES triggers the Attention-Density Guard to mask the fragment.
    6. System notifies the user of a blocked "Entanglement Attack."

## 4. Design & Architecture
* **System Flow:**
    `[Subagent Write] -> [Semantic Vectorizer] -> [Entropy Comparator] -> [CES Score] -> [Attention Gater]`
* **APIs / Interfaces:**
    * `GET /ces/score?fragment_id=...`: Retrieve the risk score for a specific shard.
    * `POST /ces/re-align`: Manually adjust the correlation baseline for a mission branch.
* **Data Storage/State:**
    * Persistent vector embeddings for mission-root fragments.

## 5. Alternatives Considered
* **Static Keyword Filtering**: Rejected as it cannot detect nuanced instruction injection or stylistic mimicry.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Uses hardware-attested vector signatures to prevent "Scorer Spoofing."
* **Observability:** Risk scores are visualized in the "Attention Entropy Heatmap."

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
