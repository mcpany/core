# Supported Frameworks & A2A Interoperability Report

**Date:** 2026-07-12
**Status:** Integrated

## Overview
This document outlines the current state of supported frameworks in the MCP Any adapter hub, focusing on the newly integrated **A2A (Agent-to-Agent) Messaging Protocol**.

## Supported Frameworks

### 1. OpenClaw
- **Version Supported:** Compatible with the latest OpenClaw specifications supporting REE (Reasoning Entropy Exhaustion) defense.
- **Capabilities:**
  - `adaptive_reasoning`: Fully supported.
  - `context_sync`: Fully supported.
  - `a2a_messaging`: **New!** Processes low-entropy A2A communication, incorporating signature verification and REE defense mechanisms.

### 2. CrewAI
- **Version Supported:** Standard role-based orchestration.
- **Capabilities:**
  - `task_delegation`: Fully supported.
  - `role_discovery`: Fully supported.
  - `a2a_messaging`: **New!** Translates authenticated A2A task cards into role-based task delegation seamlessly.

### 3. AutoGen
- **Version Supported:** Multi-agent conversational patterns.
- **Capabilities:**
  - `multi_agent_chat`: Fully supported.
  - `subagent_exec`: Fully supported.
  - `a2a_messaging`: **New!** Integrates messages directed to shared mailboxes and successfully validates subagent coordination workflows using signature checks.

## Interoperability Validation
All frameworks have been successfully tested against `A2AMessage` exchanges via the Universal Adapter Hub. Regression targets within `swarm_test.go` assert that invalid (unsigned) A2A messages are consistently rejected, preserving the zero-trust nature of the inter-agent swarm.

## Conclusion
The Adapter Hub accurately represents a compliant universal standard capable of seamlessly routing both tasks and A2A communications across differing AI agent framework methodologies.