# Sentinel's Journal

## 2026-01-10 - Unauthenticated LAN Access
**Vulnerability:** The server allowed unauthenticated access from any Private IP (RFC1918), not just localhost, when no API key was configured.
**Learning:** `util.IsPrivateIP` includes LAN ranges (10.x, 192.168.x), which is often too permissive for "local-only" defaults.
**Prevention:** Use `ip.IsLoopback()` for strict localhost checks instead of generic private IP checks.
