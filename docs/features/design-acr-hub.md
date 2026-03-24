# Design Doc: ACR Hub Controller
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
With the rise of deep agent swarms, subagent autonomy has become a security
liability. The ACR Hub implements the OpenClaw ACR protocol to provide a
centralized revocation point for all transient capabilities.

## 2. Goals & Non-Goals
* **Goals:**
  - Provide real-time capability revocation.
  - Bind all tool calls to a root mission token.
* **Non-Goals:**
  - Managing agent identity (handled by SMI Relay).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Orchestrator
* **Primary Goal:** Revoke internet access for all subagents once a search task
  is complete.
* **Happy Path:**
  1. Orchestrator issues mission-root token.
  2. ACR Hub registers transient search capability.
  3. Orchestrator signals completion.
  4. ACR Hub invalidates all related sub-tokens instantly.

## 4. Design & Architecture
The ACR Hub acts as a middleware for the A2A Bridge.

## 5. Alternatives Considered
* **Local Policy Checks:** Rejected due to lack of cross-framework propagation.

## 6. Cross-Cutting Concerns
* **Security:** Uses TPM-bound signing for revocation heartbeats.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
