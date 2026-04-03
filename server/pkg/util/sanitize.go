// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//revive:disable:var-naming
package util

import (
	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

// RedactedString represents the public RedactedString entity.
//
// Summary: Defines the structured data model representing a string.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
const RedactedString = "REDACTED"

// SanitizeUser serves as a public interface for interacting with SanitizeUser.
//
// Summary: Sanitize the user appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func SanitizeUser(u *configv1.User) *configv1.User {
	if u == nil {
		return nil
	}
	// Clone to avoid modifying the original
	clone := proto.Clone(u).(*configv1.User)

	if clone.GetAuthentication() != nil {
		clone.SetAuthentication(SanitizeAuthentication(clone.GetAuthentication()))
	}

	return clone
}

// SanitizeCredential serves as a public interface for interacting with SanitizeCredential.
//
// Summary: Sanitize the credential appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func SanitizeCredential(c *configv1.Credential) *configv1.Credential {
	if c == nil {
		return nil
	}
	clone := proto.Clone(c).(*configv1.Credential)

	if clone.GetAuthentication() != nil {
		clone.SetAuthentication(SanitizeAuthentication(clone.GetAuthentication()))
	}

	if clone.GetToken() != nil {
		clone.SetToken(SanitizeUserToken(clone.GetToken()))
	}

	return clone
}

// SanitizeAuthentication serves as a public interface for interacting with SanitizeAuthentication.
//
// Summary: Sanitize the authentication appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func SanitizeAuthentication(a *configv1.Authentication) *configv1.Authentication {
	if a == nil {
		return nil
	}

	switch a.WhichAuthMethod() {
	case configv1.Authentication_ApiKey_case:
		if ak := a.GetApiKey(); ak != nil {
			if v := ak.GetValue(); v != nil {
				ak.SetValue(SanitizeSecretValue(v))
			}
			if ak.GetVerificationValue() != "" {
				ak.SetVerificationValue(RedactedString)
			}
		}
	case configv1.Authentication_BearerToken_case:
		if bt := a.GetBearerToken(); bt != nil {
			bt.SetToken(SanitizeSecretValue(bt.GetToken()))
		}
	case configv1.Authentication_BasicAuth_case:
		if ba := a.GetBasicAuth(); ba != nil {
			ba.SetPassword(SanitizeSecretValue(ba.GetPassword()))
			if ba.GetPasswordHash() != "" {
				ba.SetPasswordHash(RedactedString)
			}
		}
	case configv1.Authentication_Oauth2_case:
		if oa := a.GetOauth2(); oa != nil {
			oa.SetClientSecret(SanitizeSecretValue(oa.GetClientSecret()))
			oa.SetClientId(SanitizeSecretValue(oa.GetClientId()))
		}
	case configv1.Authentication_TrustedHeader_case:
		if th := a.GetTrustedHeader(); th != nil && th.GetHeaderValue() != "" {
			th.SetHeaderValue(RedactedString)
		}
	}

	return a
}

// SanitizeUserToken serves as a public interface for interacting with SanitizeUserToken.
//
// Summary: Sanitize the user token appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func SanitizeUserToken(t *configv1.UserToken) *configv1.UserToken {
	if t == nil {
		return nil
	}
	if t.GetAccessToken() != "" {
		t.SetAccessToken(RedactedString)
	}
	if t.GetRefreshToken() != "" {
		t.SetRefreshToken(RedactedString)
	}
	return t
}

// SanitizeSecretValue serves as a public interface for interacting with SanitizeSecretValue.
//
// Summary: Sanitize the secret value appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func SanitizeSecretValue(s *configv1.SecretValue) *configv1.SecretValue {
	if s == nil {
		return nil
	}

	switch s.WhichValue() {
	case configv1.SecretValue_PlainText_case:
		s.SetPlainText(RedactedString)
	case configv1.SecretValue_RemoteContent_case:
		rc := s.GetRemoteContent()
		if rc != nil && rc.GetAuth() != nil {
			rc.SetAuth(SanitizeAuthentication(rc.GetAuth()))
		}
	case configv1.SecretValue_Vault_case:
		v := s.GetVault()
		if v != nil && v.GetToken() != nil {
			v.SetToken(SanitizeSecretValue(v.GetToken()))
		}
	}

	return s
}
