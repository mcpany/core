# Design Doc: Reasoning Telemetry Exporter
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
With the transition to "Attention Sovereignty" and "Reasoning budgets," there is a critical need for standardized observability of agentic reasoning effort. Currently, metrics like token consumption and reasoning intensity are scattered across framework-specific logs, making it impossible for centralized gateways like MCP Any to enforce global policies effectively.

The Reasoning Telemetry Exporter acts as a standardized metrics sink. it collects real-time data on reasoning effort, attention utilization, and budget consumption, exporting them to standard observability stacks (Prometheus/Grafana) for dashboarding and automated budget enforcement.

## 2. Goals & Non-Goals
* **Goals:**
    * Export real-time token utilization and reasoning intensity (ARE) metrics.
    * Provide "Attention Footprint" tracking per subagent.
    * Standardize reasoning effort metrics across disparate frameworks (Claude, OpenClaw, Gemini).
    * Support Prometheus-compatible scraping endpoints.
* **Non-Goals:**
    * Storing raw reasoning monologues (handled by SRM).
    * Performing real-time enforcement (handled by ALS and PBRB).

## 3. Critical User Journey (CUJ)
* **User Persona:** SRE / Swarm Platform Engineer
* **Primary Goal:** Visualize reasoning effort and attention capture across 50+ parallel subagents in a production swarm.
* **The Happy Path (Tasks):**
    1. Engineer enables the Telemetry Exporter in MCP Any configuration.
    2. Swarm begins executing complex reasoning tasks.
    3. Exporter collects ARE headers and attention scores from the ALS and PBRB middlewares.
    4. Data is aggregated and exposed via `/metrics`.
    5. Engineer views a real-time Grafana dashboard showing "Top Attention Capturers" and "Budget Burn Rate."

## 4. Design & Architecture
* **System Flow:**
    [ALS/PBRB] -> [Telemetry Aggregator] -> [Prometheus Sink] -> [Grafana]
* **APIs / Interfaces:**
    * `ExportReasoningMetric(metricName string, value float64, labels []Label)`
    * `/metrics` scrape endpoint.
* **Data Storage/State:**
    Uses in-memory cumulative counters and gauges, compatible with standard Prometheus exposition formats.

## 5. Alternatives Considered
* **Framework-specific Logging:** Rejected due to lack of standardization and high parsing overhead.
* **Direct Cloud Logging:** Rejected because it introduces latency and doesn't allow for local-first policy enforcement.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Telemetry labels are anonymized to prevent PII leakage while maintaining enough metadata for per-agent budget tracking.
* **Observability:** This *is* the observability layer for agent reasoning.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
