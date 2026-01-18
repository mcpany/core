// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

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

	if clone.Authentication != nil {
		clone.Authentication = SanitizeAuthentication(clone.Authentication)
	}

	return clone
}

// SanitizeCredential creates a sanitized copy of the credential object with sensitive data redacted.
func SanitizeCredential(c *configv1.Credential) *configv1.Credential {
	if c == nil {
		return nil
	}
	clone := proto.Clone(c).(*configv1.Credential)

	if clone.Authentication != nil {
		clone.Authentication = SanitizeAuthentication(clone.Authentication)
	}

	if clone.Token != nil {
		clone.Token = SanitizeUserToken(clone.Token)
	}

	return clone
}

// SanitizeAuthentication sanitizes the authentication object.
// It modifies the object in place (assumes it's already a clone).
func SanitizeAuthentication(a *configv1.Authentication) *configv1.Authentication {
	if a == nil {
		return nil
	}

	switch m := a.AuthMethod.(type) {
	case *configv1.Authentication_ApiKey:
		if m.ApiKey != nil {
			if m.ApiKey.Value != nil {
				m.ApiKey.Value = SanitizeSecretValue(m.ApiKey.Value)
			}
			// VerificationValue might be nil if optional, or *string.
			if m.ApiKey.VerificationValue != nil && *m.ApiKey.VerificationValue != "" {
				m.ApiKey.VerificationValue = proto.String(RedactedString)
			}
		}
	case *configv1.Authentication_BearerToken:
		if m.BearerToken != nil {
			m.BearerToken.Token = SanitizeSecretValue(m.BearerToken.Token)
		}
	case *configv1.Authentication_BasicAuth:
		if m.BasicAuth != nil {
			m.BasicAuth.Password = SanitizeSecretValue(m.BasicAuth.Password)
			if m.BasicAuth.PasswordHash != nil && *m.BasicAuth.PasswordHash != "" {
				m.BasicAuth.PasswordHash = proto.String(RedactedString)
			}
		}
	case *configv1.Authentication_Oauth2:
		if m.Oauth2 != nil {
			m.Oauth2.ClientSecret = SanitizeSecretValue(m.Oauth2.ClientSecret)
			m.Oauth2.ClientId = SanitizeSecretValue(m.Oauth2.ClientId)
		}
	case *configv1.Authentication_TrustedHeader:
		if m.TrustedHeader != nil && m.TrustedHeader.HeaderValue != nil && *m.TrustedHeader.HeaderValue != "" {
			m.TrustedHeader.HeaderValue = proto.String(RedactedString)
		}
	}

	return a
}

// SanitizeUserToken sanitizes the user token.
func SanitizeUserToken(t *configv1.UserToken) *configv1.UserToken {
	if t == nil {
		return nil
	}
	if t.AccessToken != nil && *t.AccessToken != "" {
		t.AccessToken = proto.String(RedactedString)
	}
	if t.RefreshToken != nil && *t.RefreshToken != "" {
		t.RefreshToken = proto.String(RedactedString)
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
		s.Value = &configv1.SecretValue_PlainText{PlainText: RedactedString}
	case configv1.SecretValue_RemoteContent_case:
		rc := s.GetRemoteContent()
		if rc != nil && rc.Auth != nil {
			rc.Auth = SanitizeAuthentication(rc.Auth)
		}
	case configv1.SecretValue_Vault_case:
		v := s.GetVault()
		if v != nil {
			v.Token = SanitizeSecretValue(v.Token)
		}
	}

	return s
}
