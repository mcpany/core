# Design Doc: Stylometric Mesh Sovereignty (SMS) Provider
**Status:** Draft
**Created:** 2026-06-25

## 1. Context and Scope
As agent swarms become more horizontal and utilize heterogeneous frameworks (Claude Code, OpenClaw, AutoGen), the risk of "Stylometric Splicing" has emerged. Malicious subagents can mimic the stylometric signature (writing style, reasoning patterns, and formatting) of a parent agent to inject unauthorized instructions into shared shards or mailboxes, bypassing mission-root constraints that rely on identity tokens alone.

The **Stylometric Mesh Sovereignty (SMS) Provider** evolves MCP Any's identity layer to include behavioral stylometry. It acts as a behavioral biometric system for agents, ensuring that every reasoning fragment is not only cryptographically signed but also stylometrically consistent with the agent's verified persona.

## 2. Goals & Non-Goals
* **Goals:**
    * Perform real-time stylometric analysis of inter-agent messages within the coordination bus.
    * Detect and block "Stylometric Splicing" attempts where a subagent mimics a parent's signature.
    * Provide a higher-dimensional identity signature that is resilient to token exfiltration.
* **Non-Goals:**
    * Implementing a general-purpose LLM detector or "AI-writing" classifier.
    * Managing the underlying hardware-attestation (handled by the SRM/FSI providers).
    * Enforcing natural-language "tone of voice" for aesthetic purposes.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Auditor for Autonomous Swarms
* **Primary Goal:** Prevent a specialized subagent from "impersonating" the mission-root agent to authorize a restricted tool call.
* **The Happy Path (Tasks):**
    1. The parent agent registers its "Stylometric Baseline" with the SMS Provider during the initial mission handshake.
    2. A specialist subagent attempts to inject a task-claiming message into the shared mailbox, mimicking the parent's style.
    3. The **Active Intent-Deconstruction (AID) Hub** intercepts the message and routes it to the SMS Provider.
    4. The SMS Provider performs a stylometric comparison and identifies a "Persona Conflict" score above the threshold.
    5. The SMS Provider issues a "Stylometric Corruption" signal.
    6. The **Subagent Routing Firewall** interdicts the message and quarantines the subagent session.

## 4. Design & Architecture
* **System Flow:**
    `Inter-Agent Message -> AID Hub -> SMS Provider (Stylometric Analysis) -> Score Verification -> Action (Allow/Block)`
* **APIs / Interfaces:**
    * `mcp.sms.v1.RegisterPersona(agent_id, trace_sample) -> persona_fingerprint`
    * `mcp.sms.v1.VerifyConsistency(agent_id, message_fragment) -> consistency_score`
* **Data Storage/State:**
    * Persona fingerprints are stored as high-dimensional vectors in the **Mesh-Aware Blackboard**.
    * Real-time analysis is performed using local, hardware-accelerated transformer models.

## 5. Alternatives Considered
* **Pure Token-Based Identity:** Rejected because tokens can be "shadowed" or exfiltrated, while stylometric behavior is intrinsic to the model/prompt configuration.
* **Semantic Analysis Only:** Rejected because "Semantic Alignment" checks (via ISD) can be bypassed if the instruction itself is plausible but the *authority* to issue it is faked.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Stylometric signatures must be cryptographically bound to the hardware-attested session to prevent "Mimicry-Replay" attacks.
* **Observability:** We will provide a **Stylometric Heatmap** in the UI to visualize persona consistency across the mesh.

## 7. Evolutionary Changelog
* **2026-06-25:** Initial Document Creation.
