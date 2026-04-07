// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//revive:disable:var-naming
package util

import (
	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

// Summary: RedactedString is a global constant or variable.
//
// Parameters:
//   - None
//
// Returns:
//   - None
//
// Errors:
//   - None
//
// Side Effects:
//   - None
const RedactedString = "REDACTED"

// SanitizeUser creates a sanitized copy of the user object with sensitive data redacted.
//
// Summary: SanitizeUser executes the operation.
//
// Parameters:
//   - u *configv1.User: Input parameter.
//
// Returns:
//   - *configv1.User {
: Result of the operation.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
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

// SanitizeCredential creates a sanitized copy of the credential object with sensitive data redacted.
//
// Summary: Sanitizes a credential object.
//
// Parameters:
//   - c (*configv1.Credential): The credential object to sanitize.
//
// Returns:
// Summary: SanitizeCredential executes the operation.
//
// Parameters:
//   - c *configv1.Credential: Input parameter.
//
// Returns:
//   - *configv1.Credential {
: Result of the operation.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
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

// SanitizeAuthentication sanitizes the authentication object.
// It modifies the object in place (assumes it's already a clone).
//
// Summary: Sanitizes an authentication object.
//
// Parameters:
//   - a (*configv1.Authentication): The authentication object to sanitize.
//
// Summary: SanitizeAuthentication executes the operation.
//
// Parameters:
//   - a *configv1.Authentication: Input parameter.
//
// Returns:
//   - *configv1.Authentication {
: Result of the operation.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
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

// SanitizeUserToken sanitizes the user token.
//
// Summary: Sanitizes a user token.
//
// Parameters:
//   - t (*configv1.UserToken): The user token to sanitize.
//
// Returns:
//   - *configv1.UserToken: The sanitized user token, or nil if input is nil.
// Summary: SanitizeUserToken executes the operation.
//
// Parameters:
//   - t *configv1.UserToken: Input parameter.
//
// Returns:
//   - *configv1.UserToken {
: Result of the operation.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
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

// SanitizeSecretValue sanitizes a SecretValue.
//
// Summary: Sanitizes a SecretValue object.
//
// Parameters:
//   - s (*configv1.SecretValue): The secret value to sanitize.
//
// Returns:
// Summary: SanitizeSecretValue executes the operation.
//
// Parameters:
//   - s *configv1.SecretValue: Input parameter.
//
// Returns:
//   - *configv1.SecretValue {
: Result of the operation.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
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
