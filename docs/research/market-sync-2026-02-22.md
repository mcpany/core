# Market Sync: 2026-02-22

## Ecosystem Shifts

### OpenClaw Viral Surge & Security Crisis
OpenClaw (formerly Moltbot) has achieved unprecedented viral growth, surpassing 200,000 GitHub stars. It represents a shift towards "local-first" autonomous agents that integrate with messaging apps (WhatsApp, Telegram). However, its "light safety scaffolding" has led to significant security concerns, with reports of unauthorized host-level file access and lack of governance frameworks.

### Multi-Model Orchestration via MCP
Claude Code and Gemini CLI are increasingly being used together. Tools like "Zen MCP" and the "clink" command are emerging to pipe commands between different CLI-based agents. There is a clear trend towards using MCP as the "universal bus" for multi-model workflows.

### Skill Portability
The emergence of "skill-porter" indicates a demand for cross-platform skill compatibility between Claude Code and Gemini CLI. Standardizing how skills (tools + context) are defined and ported is becoming critical.

## Autonomous Agent Pain Points
* **Security vs. Capability:** Users want powerful local execution (shell, browser) but are vulnerable to rogue agent actions.
* **Context Fragmentation:** Sharing state and context between different agent CLI tools is still clunky (e.g., manual piping).
* **Vendor Lock-in:** Users are looking for ways to avoid being locked into a single provider's ecosystem by using MCP-based adapters.

## Security Vulnerabilities
* **Unauthorized Host Access:** Local agents executing shell commands without strict sandboxing.
* **Prompt Injection in Messaging Apps:** Messaging interfaces (WhatsApp/Telegram) present new vectors for indirect prompt injection.
