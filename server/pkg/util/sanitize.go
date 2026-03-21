// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//revive:disable:var-naming
package util

import (
	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

// Summary: RedactedString is the string used to replace sensitive data. Constant string for redacted data.
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

// Summary: SanitizeUser creates a sanitized copy of the user object with sensitive data redacted. Sanitizes a user object.
//
// Parameters:
//   - u (*configv1.User): The u parameter.
//
// Returns:
//   - *configv1.User: The resulting *configv1.User.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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

// Summary: SanitizeCredential creates a sanitized copy of the credential object with sensitive data redacted. Sanitizes a credential object.
//
// Parameters:
//   - c (*configv1.Credential): The c parameter.
//
// Returns:
//   - *configv1.Credential: The resulting *configv1.Credential.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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

// Summary: SanitizeAuthentication sanitizes the authentication object. It modifies the object in place (assumes it's already a clone). Sanitizes an authentication object.
//
// Parameters:
//   - a (*configv1.Authentication): The a parameter.
//
// Returns:
//   - *configv1.Authentication: The resulting *configv1.Authentication.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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

// Summary: SanitizeUserToken sanitizes the user token. Sanitizes a user token.
//
// Parameters:
//   - t (*configv1.UserToken): The t parameter.
//
// Returns:
//   - *configv1.UserToken: The resulting *configv1.UserToken.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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

// Summary: SanitizeSecretValue sanitizes a SecretValue. Sanitizes a SecretValue object.
//
// Parameters:
//   - s (*configv1.SecretValue): The s parameter.
//
// Returns:
//   - *configv1.SecretValue: The resulting *configv1.SecretValue.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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
