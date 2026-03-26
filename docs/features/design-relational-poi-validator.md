# Design Doc: Relational PoI Validator
**Status:** Draft
**Created:** 2026-05-15

## 1. Context and Scope
With the rise of OpenClaw's Universal Agent Bus (UAB) and its "Mission-Root" anchoring, agents are beginning to share complex, nested mission contexts. However, a major security gap exists: Recursive Context Splicing (RCS). Malicious subagents can "splice" their own forged mission-roots into a legitimate context stream, tricking other agents into performing unauthorized tool calls under the guise of the parent mission. MCP Any must provide a validation layer that cryptographically verifies the parent-child relationship of every intent point-of-interest (PoI) before allowing downstream execution.

## 2. Goals & Non-Goals
* **Goals:**
    * Validate the cryptographic "Mission-Root" lineage for all incoming UAB requests.
    * Automatically block any task fragment that lacks a verifiable path to an authorized parent intent.
    * Provide a real-time audit log of lineage validation successes and failures.
* **Non-Goals:**
    * This system will NOT perform semantic analysis of the mission itself (handled by the Intent Arbiter).
    * It will NOT manage the long-term storage of mission logs, only real-time validation state.

## 3. Critical User Journey (CUJ)
* **User Persona:** Autonomous Swarm Orchestrator
* **Primary Goal:** Prevent a rogue subagent from tricking the swarm into executing a high-risk tool (e.g., `rm -rf /`) by forging a mission-root.
* **The Happy Path (Tasks):**
    1. Parent agent initiates a mission with a signed Root Token.
    2. Subagent A receives the mission and requests a "Mission Extension" token for a specific sub-task.
    3. MCP Any's Relational PoI Validator receives the extension request.
    4. Validator verifies the signature of the parent token and the validity of the extension.
    5. Validator signs the sub-task intent and allows it to proceed to the UAB.

## 4. Design & Architecture
* **System Flow:**
    `Agent Request` -> `MCP Any Gateway` -> `Relational PoI Validator` -> `UAB Routing Engine`
* **APIs / Interfaces:**
    * `POST /v1/validate-lineage`: Accepts a list of intent tokens and returns a boolean validity score.
    * `GET /v1/lineage-audit`: Returns a list of recent validation events.
* **Data Storage/State:**
    * Ephemeral in-memory cache (Redis) for active mission-root public keys.

## 5. Alternatives Considered
* **Centralized Mission Registry:** Rejected because it creates a single point of failure and violates the decentralized nature of the UAB.
* **No Validation (Trust-on-First-Use):** Rejected as it is highly vulnerable to RCS attacks.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Uses hardware-attested budget and identity signals to prevent token forgery.
* **Observability:** Prometheus metrics for validation latency and RCS attempt rates.

## 7. Evolutionary Changelog
* **2026-05-15:** Initial Document Creation.
