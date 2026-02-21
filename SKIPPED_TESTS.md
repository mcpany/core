# Skipped Tests Manifest

## UI Tests
- `ui/tests/audit-logs.spec.ts`: Skips if `CAPTURE_SCREENSHOTS` is not set.

## Server Tests
### Integration (Upstream)
- `server/tests/integration/upstream/mcp_bundle_test.go`: Skips due to Docker dependency.
- `server/tests/integration/upstream/mcp_stdio_docker_test.go`: Skips due to Docker dependency.
- `server/tests/integration/upstream/openapi_test.go`: Skips if `GEMINI_API_KEY` is missing.
- `server/tests/integration/upstream/websocket_test.go`: Skips if `GEMINI_API_KEY` is missing.
- `server/tests/integration/upstream/copilot_cli_e2e_test.go`: Skips if `GITHUB_COPILOT_TOKEN` is missing.
- `server/tests/integration/upstream/claude_cli_e2e_test.go`: Skips if `ANTHROPIC_API_KEY` is missing.
- `server/tests/integration/upstream/gemini_cli_e2e_test.go`: Skips if `GEMINI_API_KEY` is missing.
- `server/tests/integration/upstream/webrtc_test.go`: Skips if `GEMINI_API_KEY` is missing.

### Public API Tests (Rate Limited / Flaky)
- `server/tests/public_api/agify_test.go`
- `server/tests/public_api/bored_test.go`
- `server/tests/public_api/cat_facts_test.go`
- `server/tests/public_api/deck_of_cards_test.go`
- `server/tests/public_api/dog_facts_test.go`
- `server/tests/public_api/fun_translations_test.go`
- `server/tests/public_api/genderize_test.go`
- `server/tests/public_api/nationalize_test.go`
- `server/tests/public_api/open_brewery_db_test.go`
- `server/tests/public_api/open_notify_test.go`
- `server/tests/public_api/pokeapi_test.go`
- `server/tests/public_api/public_api_ipinfo_test.go`
- `server/tests/public_api/the_cocktail_db_test.go`
- `server/tests/public_api/the_meal_db_test.go`
- `server/tests/public_api/universities_list_test.go`
- `server/tests/public_api/zippopotam_test.go`
