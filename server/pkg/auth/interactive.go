// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"golang.org/x/oauth2"
)

func ptr(s string) *string {
	return &s
}

// resolveSecretValue helps extract string from SecretValue oneof
func resolveSecretValue(sv *configv1.SecretValue) string {
	if sv == nil {
		return ""
	}
	switch v := sv.Value.(type) {
	case *configv1.SecretValue_PlainText:
		return v.PlainText
	case *configv1.SecretValue_EnvironmentVariable:
		// TODO: Implement env var lookup
		return ""
	case *configv1.SecretValue_FilePath:
		// TODO: Implement file lookup
		return ""
	default:
		return ""
	}
}

// InitiateOAuth starts the OAuth2 flow for a given service.
// It returns the authorization URL and the state parameter.
func (am *Manager) InitiateOAuth(ctx context.Context, userID, serviceID, redirectURL string) (string, string, error) {
	am.mu.RLock()
	storage := am.storage
	am.mu.RUnlock()

	if storage == nil {
		return "", "", fmt.Errorf("storage not initialized")
	}

	// ... (rest of simple checks)

	// Get service config to find OAuth settings
	service, err := storage.GetService(ctx, serviceID)
	if err != nil {
		return "", "", fmt.Errorf("failed to get service %s: %w", serviceID, err)
	}
	if service == nil {
		return "", "", fmt.Errorf("service %s not found", serviceID)
	}

	upstreamAuth := service.GetUpstreamAuth()
	if upstreamAuth == nil {
		return "", "", fmt.Errorf("service %s has no upstream auth configuration", serviceID)
	}

	oauthConfig := upstreamAuth.GetOauth2()
	if oauthConfig == nil {
		return "", "", fmt.Errorf("service %s is not configured for OAuth2", serviceID)
	}

	clientID := resolveSecretValue(oauthConfig.GetClientId())
	clientSecret := resolveSecretValue(oauthConfig.GetClientSecret())

	conf := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{oauthConfig.GetScopes()},
		Endpoint: oauth2.Endpoint{
			AuthURL:  oauthConfig.GetAuthorizationUrl(),
			TokenURL: oauthConfig.GetTokenUrl(),
		},
		RedirectURL: redirectURL,
	}

	if conf.Endpoint.AuthURL == "" {
		if oauthConfig.GetIssuerUrl() != "" {
			// TODO: Add OIDC discovery
		}
		if conf.Endpoint.AuthURL == "" {
			return "", "", fmt.Errorf("authorization_url is required")
		}
	}

	// Generate random state
	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		return "", "", fmt.Errorf("failed to generate state: %w", err)
	}
	state := base64.URLEncoding.EncodeToString(stateBytes)

	url := conf.AuthCodeURL(state, oauth2.AccessTypeOffline)
	return url, state, nil
}

// HandleOAuthCallback handles the OAuth2 callback code exchange.
func (am *Manager) HandleOAuthCallback(ctx context.Context, userID, serviceID, code, redirectURL string) error {
	am.mu.RLock()
	storage := am.storage
	am.mu.RUnlock()

	if storage == nil {
		return fmt.Errorf("storage not initialized")
	}

	service, err := storage.GetService(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("failed to get service %s: %w", serviceID, err)
	}
	if service == nil {
		return fmt.Errorf("service %s not found", serviceID)
	}

	upstreamAuth := service.GetUpstreamAuth()
	if upstreamAuth == nil {
		return fmt.Errorf("service %s has no upstream auth configuration", serviceID)
	}
	oauthConfig := upstreamAuth.GetOauth2()
	if oauthConfig == nil {
		return fmt.Errorf("service %s is not configured for OAuth2", serviceID)
	}

	clientID := resolveSecretValue(oauthConfig.GetClientId())
	clientSecret := resolveSecretValue(oauthConfig.GetClientSecret())

	conf := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			TokenURL: oauthConfig.GetTokenUrl(),
		},
		RedirectURL: redirectURL,
	}

	token, err := conf.Exchange(ctx, code)
	if err != nil {
		return fmt.Errorf("failed to exchange code: %w", err)
	}

	// Save token
	userToken := &configv1.UserToken{
		UserId:       ptr(userID),
		ServiceId:    ptr(serviceID),
		AccessToken:  ptr(token.AccessToken),
		RefreshToken: ptr(token.RefreshToken),
		TokenType:    ptr(token.TokenType),
		Expiry:       ptr(token.Expiry.Format(time.RFC3339)),
		Scope:        ptr(oauthConfig.GetScopes()),
		UpdatedAt:    ptr(time.Now().Format(time.RFC3339)),
	}

	if err := storage.SaveToken(ctx, userToken); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	return nil
}
