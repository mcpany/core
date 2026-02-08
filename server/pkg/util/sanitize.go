// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//revive:disable:var-naming
package util

import (
	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

// RedactedString is the string used to replace sensitive data.
const RedactedString = "REDACTED"

// SanitizeUser creates a sanitized copy of the user object with sensitive data redacted.
//
// Summary: creates a sanitized copy of the user object with sensitive data redacted.
//
// Parameters:
//   - u: *configv1.User. The u.
//
// Returns:
//   - *configv1.User: The *configv1.User.
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
// Summary: creates a sanitized copy of the credential object with sensitive data redacted.
//
// Parameters:
//   - c: *configv1.Credential. The c.
//
// Returns:
//   - *configv1.Credential: The *configv1.Credential.
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
//
// Summary: sanitizes the authentication object.
//
// Parameters:
//   - a: *configv1.Authentication. The a.
//
// Returns:
//   - *configv1.Authentication: The *configv1.Authentication.
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
// Summary: sanitizes the user token.
//
// Parameters:
//   - t: *configv1.UserToken. The t.
//
// Returns:
//   - *configv1.UserToken: The *configv1.UserToken.
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
// Summary: sanitizes a SecretValue.
//
// Parameters:
//   - s: *configv1.SecretValue. The s.
//
// Returns:
//   - *configv1.SecretValue: The *configv1.SecretValue.
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
