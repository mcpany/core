# Design Doc: Stylometric Mesh Sovereignty (SMS) Provider
**Status:** Draft
**Created:** 2026-06-25

## 1. Context and Scope
The emergence of "Stylometric Splicing" poses a critical threat to swarm coordination. Malicious or compromised subagents can mimic the writing style and "cognitive signature" of a parent agent or a high-trust teammate to inject unauthorized instructions into shared mailboxes. Because traditional identity tokens are transport-layer, they don't protect the *content* of the message from being a mimicry-based hijacking attempt.

The **SMS Provider** adds a layer of behavioral security by performing real-time stylometric analysis of all inter-agent messages. It ensures that the "Voice" of the agent matches its hardware-attested identity.

## 2. Goals & Non-Goals
* **Goals:**
    * Perform real-time stylometric consistency checks on all inter-agent coordination messages.
    * Alert and/or block messages where the stylometric signature diverges from the attested agent baseline.
    * Maintain a "Behavioral Profile Registry" bound to hardware identities.
* **Non-Goals:**
    * Detecting factual hallucinations (handled by ARI).
    * Preventing style changes that are explicitly part of a "Persona Switch" authorized by the Mission Root.

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Security Auditor
* **Primary Goal:** Prevent a specialist subagent from impersonating the "Lead Architect" agent in a shared teammate mailbox.
* **The Happy Path (Tasks):**
    1. The "Lead Architect" agent establishes its behavioral baseline during the initial mission boot.
    2. A specialist subagent attempts to send a task-delegation message, mimicking the Lead Architect's professional tone and syntax.
    3. The SMS Provider analyzes the message fragments.
    4. The Provider detects a "Stylometric Mismatch" between the message and the subagent's attested profile.
    5. The message is quarantined, and the Mission Root is notified of a "Sovereignty Corruption" attempt.

## 4. Design & Architecture
* **System Flow:**
    `Agent Message -> SMS Middleware -> [Stylometric Analysis] -> [Profile Matching] -> Result`
* **APIs / Interfaces:**
    * `mcp.sms.v1.VerifySignature(message_blob, agent_id)`: Returns a confidence score.
    * `mcp.sms.v1.UpdateProfile(agent_id, trace_samples)`: Refines the behavioral baseline.
* **Data Storage/State:**
    * Behavioral profiles are stored as encrypted, hardware-bound vectors in the Blackboard.

## 5. Alternatives Considered
* **Pure Cryptographic Signing**: Rejected because a compromised agent can still sign a mimicry-based injection. SMS protects the *semantic* identity.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** SMS provides "Depth-of-Defense" against reasoning hijacking. It is particularly effective against "Stylometric Splicing" in horizontal meshes.
* **Observability:** Stylometric drift scores will be visualized in the "Behavioral Anchoring Monitor."

## 7. Evolutionary Changelog
* **2026-06-25:** Initial Document Creation.
