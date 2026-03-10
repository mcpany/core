# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

# Design Doc: Agent Cost-Control Middleware
**Status:** Draft
**Created:** 2026-03-09

## 1. Context and Scope
As agentic workflows become more complex and autonomous, the frequency of tool calls and recursive reasoning loops has led to unpredictable token costs. Frameworks like Claude Code and OpenClaw can quickly consume significant budgets if an agent enters an inefficient loop. MCP Any, as the central gateway for tool execution, is uniquely positioned to enforce cost-aware policies at the infrastructure level, protecting users from "runaway" agent expenses.

## 2. Goals & Non-Goals
* **Goals:**
    * Real-time monitoring of token usage per agent session.
    * Configurable soft and hard budget limits.
    * "Economical Routing" suggestions (e.g., suggesting local models for repetitive tasks).
    * Standardized cost telemetry injection into tool responses.
* **Non-Goals:**
    * Replacing the billing systems of LLM providers.
    * Guaranteeing exact penny-accurate billing (due to varying model pricing).
    * Enforcing limits on non-MCP tool calls.

## 3. Critical User Journey (CUJ)
* **User Persona:** Platform Engineer / AI Architect
* **Primary Goal:** Prevent an autonomous agent swarm from exceeding a $50/day budget.
* **The Happy Path (Tasks):**
    1. The user defines a `cost_policy` in `mcpany.yaml` with a daily limit and an alert threshold.
    2. An agent initiates a session and starts making tool calls.
    3. MCP Any intercepts each call, calculates estimated token usage, and decrements the session budget.
    4. When the budget reaches 80% (soft limit), MCP Any injects a warning into the tool response.
    5. When the budget is exhausted (hard limit), MCP Any blocks subsequent tool calls and returns a `429 Cost Limit Exceeded` error.

## 4. Design & Architecture
* **System Flow:**
    `Agent -> [MCP Any Gateway] -> [Cost Middleware] -> [Tool Executor] -> [Provider API]`
    The Middleware wraps the Tool Executor, calculating costs based on request/response size and model metadata.
* **APIs / Interfaces:**
    * `GET /v1/sessions/{id}/budget`: Retrieve current budget status.
    * `POST /v1/policies/cost`: Update cost enforcement rules.
* **Data Storage/State:**
    * Usage metrics are stored in the local SQLite "Resident State" for low-latency lookups and persistence.

## 5. Alternatives Considered
* **Model-Level Limits**: Rejected because many providers do not offer granular, real-time budget hooks for specific sessions.
* **Framework-Level Limits (e.g., LangChain)**: Rejected because it requires modifying every agent script; MCP Any provides a universal, non-invasive enforcement layer.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Cost policies are signed and immutable once a session starts to prevent "budget poisoning" by compromised agents.
* **Observability**: Integration with Prometheus/Grafana for real-time cost visualization.

## 7. Evolutionary Changelog
* **2026-03-09:** Initial Document Creation.
