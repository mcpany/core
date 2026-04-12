# Design Doc: Chapter-Bound Mission Broker
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the stabilization of "Chapters" in Gemini CLI, agents are increasingly moving toward a model where long sessions are broken into logical, task-oriented phases. However, current implementations lack a security boundary between these chapters, allowing an agent compromised in one chapter to exfiltrate context or abuse tools authorized for a different phase.

MCP Any needs to solve this by providing a first-class "Chapter" abstraction that cryptographically binds tool capabilities and context shards to specific mission phases. This ensures that the "intent-drift" observed in deep swarms is contained within the current chapter's boundaries.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a mechanism to define and transition between mission chapters.
    * Cryptographically bind tool access to the active chapter.
    * Isolate context shards such that they are only accessible within authorized chapters.
    * Support hardware-attested chapter transitions.
* **Non-Goals:**
    * Automatically generating chapter boundaries (this is the agent's/user's responsibility).
    * Providing a chat interface for chapter management (this is a UI/UX concern).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Swarm Architect
* **Primary Goal:** Ensure that a "Database Migration" chapter cannot access "Social Media Posting" tools authorized for a later phase of the mission.
* **The Happy Path (Tasks):**
    1. The architect defines a mission manifest with two chapters: `migration` and `announcement`.
    2. The agent starts the `migration` chapter; MCP Any issues a chapter-bound token.
    3. The agent attempts to call a `post_to_twitter` tool; MCP Any interdicts the call as it's not authorized for the `migration` chapter.
    4. The agent completes migration and requests a transition to the `announcement` chapter.
    5. MCP Any verifies the transition (potentially with user attestation) and issues a new token.
    6. The agent now successfully calls `post_to_twitter`.

## 4. Design & Architecture
* **System Flow:**
    * The `Chapter-Bound Mission Broker` sits as a middleware between the Agent and the Tool Registry.
    * It maintains a `Chapter Manifest` which maps Chapter IDs to Tool Scopes and Context Shard IDs.
* **APIs / Interfaces:**
    * `POST /v1/mission/chapter/start`: Initiates a chapter and returns a scoped token.
    * `POST /v1/mission/chapter/transition`: Handles the atomic move between phases.
* **Data Storage/State:**
    * Chapter state is stored in the `Mission-Root Registry`, linked to the hardware-attested mission session.

## 5. Alternatives Considered
* **Flat Intent Chains:** Rejected because they don't provide the temporal isolation required for multi-day missions.
* **Per-Task Scoping:** Too granular for most users and leads to "Approval Fatigue." Chapters provide the right balance of security and usability.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All chapter transitions must be hardware-attested. Chapter tokens are short-lived.
* **Observability:** Every chapter start, transition, and interdiction event is logged in the `Action-Chain Sovereignty Monitor`.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
