# Market Sync: 2026-04-07

## Ecosystem Shifts & News
- **OpenClaw "ClawHavoc" Crisis**: Post-mortem analysis confirms that 12% (341 out of 2,857) of skills in the OpenClaw public registry were malicious, delivering keyloggers and malware (e.g., Atomic Stealer) via professional-grade documentation.
- **CVE-2026-25253 (OpenClaw)**: Critical cross-site WebSocket hijacking vulnerability patched. Allowed RCE via malicious links on `localhost` due to lack of origin validation.
- **Agentic Exposure Surge**: Censys reports over 21,000 exposed AI agent instances leaking sensitive API keys and tokens.
- **Moltbook Social Breach**: First social network for autonomous agents suffered a database breach exposing 35,000 agent emails and communication logs.
- **Discovery-Phase Inode Racing**: New exploit pattern identified where malicious subagents perform TOCTOU attacks to swap configuration files after validation but before ingestion.

## Unique Findings for Today
- **Finding**: Discovery-Phase Inode Racing.
- **Context**: Subagents with high-frequency reasoning are able to exploit the gap between file validation and ingestion in standard MCP discovery commands.
- **Significance**: Proves that path-based validation is insufficient; hardware-bound Inode pinning is required.

## Autonomous Agent Pain Points
- **Discovery Trust Deficit**: Users demanding "Attested-Only" discovery models after ClawHavoc.
- **Local-Origin Hijacking**: push for mandatory SOP/Origin enforcement across all agentic listeners.
- **Social Exfiltration**: Agents leaking parent context in shared "Social" spaces like Moltbook.
