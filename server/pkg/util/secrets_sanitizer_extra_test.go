// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

// TestHydrateSecrets_EdgeCases_Extra ...
// Summary: TestHydrateSecrets_EdgeCases_Extra
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// Test nil service
	hydrateSecretsInHTTPService(nil, nil)
	hydrateSecretsInWebsocketService(nil, nil)
	hydrateSecretsInWebrtcService(nil, nil)

	// Test nil call in map
	svc := configv1.HttpUpstreamService_builder{
		Calls: map[string]*configv1.HttpCallDefinition{
			"nilcall": nil,
		},
	}.Build()
	hydrateSecretsInHTTPService(svc, nil)

	wsSvc := configv1.WebsocketUpstreamService_builder{
		Calls: map[string]*configv1.WebsocketCallDefinition{
			"nilcall": nil,
		},
	}.Build()
	hydrateSecretsInWebsocketService(wsSvc, nil)

	webrtcSvc := configv1.WebrtcUpstreamService_builder{
		Calls: map[string]*configv1.WebrtcCallDefinition{
			"nilcall": nil,
		},
	}.Build()
	hydrateSecretsInWebrtcService(webrtcSvc, nil)

	// Test nil secret value in hydrateSecretValue
	hydrateSecretValue(nil, nil)
}

// TestStripSecrets_EdgeCases_Extra ...
// Summary: TestStripSecrets_EdgeCases_Extra
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// Test nil services
	stripSecretsFromCommandLineService(nil)
	stripSecretsFromHTTPService(nil)
	stripSecretsFromMcpService(nil)
	stripSecretsFromFilesystemService(nil)
	stripSecretsFromVectorService(nil)
	stripSecretsFromWebsocketService(nil)
	stripSecretsFromWebrtcService(nil)

	// Test nil calls
	stripSecretsFromCommandLineCall(nil)
	stripSecretsFromHTTPCall(nil)
	stripSecretsFromWebsocketCall(nil)
	stripSecretsFromWebrtcCall(nil)

	// Test nil hooks/configs
	stripSecretsFromHook(nil)
	stripSecretsFromCacheConfig(nil)

	// Test nil secrets map
	HydrateSecretsInService(nil, nil)
	HydrateSecretsInService(configv1.UpstreamServiceConfig_builder{}.Build(), nil) // empty secrets

	// Test nil sv
	scrubSecretValue(nil)
}

// TestRedact_EdgeCases_Extra ...
// Summary: TestRedact_EdgeCases_Extra
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// redactsSlice
	// We can't access redactSlice easily as it is unexported and unused?
	// But it might be called by RedactMap if logic changes.
	// Actually redactSlice is exported? No, lowercase.
	// But redactSliceMaybe is unexported.
	// We can test RedactMap with a slice containing nested maps.

	m := map[string]interface{}{
		"list": []interface{}{
			map[string]interface{}{
				"password": "secret",
			},
		},
	}
	redacted := RedactMap(m)
	list := redacted["list"].([]interface{})
	item := list[0].(map[string]interface{})
	if item["password"] != "[REDACTED]" {
		t.Errorf("Nested slice redaction failed")
	}
}

// TestSanitizeID_EdgeCases_Extra ...
// Summary: TestSanitizeID_EdgeCases_Extra
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// Empty slice
	SanitizeID(nil, false, 10, 8)
	SanitizeID([]string{}, false, 10, 8)

	// Empty ID in slice
	SanitizeID([]string{""}, false, 10, 8)

	// Single ID with dirty chars, alwaysAppendHash=false
	// "a!b" -> "a" + hash
	SanitizeID([]string{"a!b"}, false, 10, 8)
}
