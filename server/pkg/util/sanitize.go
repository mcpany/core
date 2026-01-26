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
func SanitizeAuthentication(a *configv1.Authentication) *configv1.Authentication {
	if a == nil {
		return nil
	}

	// Since we can't easily modify oneof fields in-place if they are private/hidden in Opaque API without using Setters,
	// checking the type and then setting the field is cleaner.
	// However, we are operating on a clone, so we can use Setters.

	switch {
	case a.HasApiKey():
		m := a.GetApiKey() // Returns *APIKeyAuth
		if m != nil {
			if m.HasValue() {
				m.SetValue(SanitizeSecretValue(m.GetValue()))
			}
			if m.GetVerificationValue() != "" {
				m.SetVerificationValue(RedactedString)
			}
		}
	case a.HasBearerToken():
		m := a.GetBearerToken()
		if m != nil {
			m.SetToken(SanitizeSecretValue(m.GetToken()))
		}
	case a.HasBasicAuth():
		m := a.GetBasicAuth()
		if m != nil {
			m.SetPassword(SanitizeSecretValue(m.GetPassword()))
			if m.GetPasswordHash() != "" {
				m.SetPasswordHash(RedactedString)
			}
		}
	case a.HasOauth2():
		m := a.GetOauth2()
		if m != nil {
			m.SetClientSecret(SanitizeSecretValue(m.GetClientSecret()))
			m.SetClientId(SanitizeSecretValue(m.GetClientId()))
		}
	case a.HasTrustedHeader():
		m := a.GetTrustedHeader()
		if m != nil && m.GetHeaderValue() != "" {
			m.SetHeaderValue(RedactedString)
		}
	}

	return a
}

// SanitizeUserToken sanitizes the user token.
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
		if v != nil {
			v.SetToken(SanitizeSecretValue(v.GetToken()))
		}
	}

	return s
}
