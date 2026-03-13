# Market Sync: 2026-03-25

## Ecosystem Updates

### OpenClaw: Enhanced Sub-Agent Routing
- **Improved Routing Logic**: OpenClaw has redesigned its message delivery system for sub-agent coordination.
- **Defined Channels**: Communication now travels through explicitly defined channels rather than ambiguous paths, increasing reliability and preventing message loss between parent and sub-agents.
- **Heartbeat DM Restored**: The latest update restores heartbeat DM delivery for monitoring agents, transforming Slack into an alert hub for events like API failures or keyword ranking changes.

### Claude Code: Configuration Sandbox Hardening (CVE-2026-25725)
- **Privilege Escalation Flaw**: A vulnerability was identified where `settings.json` could be abused if not properly sandboxed at startup.
- **Proactive Protection**: Anthropic version 2.1.2 ensures that `settings.json` receives read-only sandbox protections regardless of whether the file exists when the agent starts.
- **"Settings-as-Execution" Trend**: This reinforces the shift where repository-level configuration files are becoming part of the active execution layer, requiring stricter governance.

### Agentic PR Security Crisis
- **High Vulnerability Rate**: A report from DryRun Security reveals that AI coding agents (Claude Code, OpenAI Codex, Google Gemini) produce security issues in 87% of their pull requests.
- **Legacy Mistakes**: Agents are frequently repeating decade-old security vulnerabilities, highlighting a critical need for automated security scanning within the agentic workflow.

## Summary of Findings
Today's research highlights a transition from "Access Control" to "Execution Governance." As agents become more autonomous and swarms more complex, the reliability of inter-agent communication and the integrity of local configuration files are the primary failure points. The high vulnerability rate in agent-generated code also suggests that MCP Any should play a role in proactive security scanning.
