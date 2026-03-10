# Design Doc: Containerized Skill Execution (Ephemeral Sandboxing)
**Status:** Draft
**Created:** 2026-03-10

## 1. Context and Scope
Recent RCE vulnerabilities in OpenClaw and the rise of "Poisoned Skills" (malicious MCP servers) make running MCP servers as local processes extremely risky. MCP Any must provide a "Safe-by-Default" execution environment where skills are isolated from the host by default.

## 2. Goals & Non-Goals
* **Goals:**
    * Automatically wrap Stdio/HTTP MCP servers in ephemeral Docker containers.
    * Implement "No-Host-Access" default mount policy.
    * Provide a configuration schema for "Safe Port Forwarding."
* **Non-Goals:**
    * Replacing existing local execution (it remains an opt-in for trusted tools).
    * Building a new container runtime (we leverage Docker/Podman).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer
* **Primary Goal:** Run a community-contributed "SQL Optimizer" skill without worrying about it reading their `~/.ssh` directory.
* **The Happy Path (Tasks):**
    1. User adds the skill to `config.yaml` with `sandbox: true`.
    2. MCP Any pulls/builds the minimal container image for the skill.
    3. Upon tool call, MCP Any starts an ephemeral container instance.
    4. The tool executes, and the container is immediately destroyed.

## 4. Design & Architecture
* **System Flow:**
    MCP Any -> Sandbox Manager -> Docker API -> [Ephemeral Container] -> Stdio Bridge -> Agent.
* **APIs / Interfaces:**
    * `sandbox` block in `mcp.yaml`: `image`, `memory_limit`, `cpu_limit`, `allowed_paths`.
* **Data Storage/State:**
    * Container logs streamed to MCP Any's internal audit log.

## 5. Alternatives Considered
* **gVisor/Firecracker:** Considered for higher isolation, but Docker/Podman have better local developer adoption.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Rootless containers are preferred. No privileged mode allowed.
* **Observability:** Sandbox health metrics (OOM kills, etc.) reported to UI.

## 7. Evolutionary Changelog
* **2026-03-10:** Initial Document Creation.
