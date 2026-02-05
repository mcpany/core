// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/util"
	"golang.org/x/oauth2"
	"google.golang.org/protobuf/proto"
)

// InitiateOAuth starts the OAuth2 flow for a given service or credential.
// It returns the authorization URL and the state parameter.
func (am *Manager) InitiateOAuth(ctx context.Context, userID, serviceID, credentialID, redirectURL string) (string, string, error) {
	// Fix for unused userID:
	_ = userID

	am.mu.RLock()
	storage := am.storage
	am.mu.RUnlock()

	if storage == nil {
		return "", "", fmt.Errorf("storage not initialized")
	}

	var oauthConfig *configv1.OAuth2Auth

	switch {
	case credentialID != "":
		cred, err := storage.GetCredential(ctx, credentialID)
		if err != nil {
			return "", "", fmt.Errorf("failed to get credential %s: %w", credentialID, err)
		}
		if cred == nil {
			return "", "", fmt.Errorf("credential %s not found", credentialID)
		}
		if cred.GetAuthentication() == nil {
			return "", "", fmt.Errorf("credential %s has no authentication config", credentialID)
		}
		oauthConfig = cred.GetAuthentication().GetOauth2()
		if oauthConfig == nil {
			return "", "", fmt.Errorf("credential %s is not configured for OAuth2", credentialID)
		}
	case serviceID != "":
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

		oauthConfig = upstreamAuth.GetOauth2()
		if oauthConfig == nil {
			return "", "", fmt.Errorf("service %s is not configured for OAuth2", serviceID)
		}
	default:
		return "", "", fmt.Errorf("either service_id or credential_id must be provided")
	}

	clientID, err := util.ResolveSecret(ctx, oauthConfig.GetClientId())
	if err != nil {
		return "", "", fmt.Errorf("failed to resolve client_id: %w", err)
	}

	clientSecret, err := util.ResolveSecret(ctx, oauthConfig.GetClientSecret())
	if err != nil {
		return "", "", fmt.Errorf("failed to resolve client_secret: %w", err)
	}

	conf := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       strings.Split(oauthConfig.GetScopes(), " "),
		Endpoint: oauth2.Endpoint{
			AuthURL:  oauthConfig.GetAuthorizationUrl(),
			TokenURL: oauthConfig.GetTokenUrl(),
		},
		RedirectURL: redirectURL,
	}
	// Fallback to "scopes" field as space-delimited string if splitting fails or logic changes.
	// Actually oauthConfig.GetScopes() is a string "space-delimited list of scopes".
	// oauth2.Config.Scopes is []string. So Split is correct.

	if conf.Endpoint.AuthURL == "" {
		if oauthConfig.GetIssuerUrl() != "" {
			// TODO: Add OIDC discovery
			return "", "", fmt.Errorf("OIDC discovery not implemented")
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
func (am *Manager) HandleOAuthCallback(ctx context.Context, userID, serviceID, credentialID, code, redirectURL string) error {
	am.mu.RLock()
	storage := am.storage
	am.mu.RUnlock()

	if storage == nil {
		return fmt.Errorf("storage not initialized")
	}

	var oauthConfig *configv1.OAuth2Auth
	var cred *configv1.Credential

	switch {
	case credentialID != "":
		var err error
		cred, err = storage.GetCredential(ctx, credentialID)
		if err != nil {
			return fmt.Errorf("failed to get credential %s: %w", credentialID, err)
		}
		if cred == nil {
			return fmt.Errorf("credential %s not found", credentialID)
		}
		oauthConfig = cred.GetAuthentication().GetOauth2()
		if oauthConfig == nil {
			return fmt.Errorf("credential %s is not configured for OAuth2", credentialID)
		}
	case serviceID != "":
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
		oauthConfig = upstreamAuth.GetOauth2()
		if oauthConfig == nil {
			return fmt.Errorf("service %s is not configured for OAuth2", serviceID)
		}
	default:
		return fmt.Errorf("either service_id or credential_id must be provided")
	}

	clientID, err := util.ResolveSecret(ctx, oauthConfig.GetClientId())
	if err != nil {
		return fmt.Errorf("failed to resolve client_id: %w", err)
	}
	clientSecret, err := util.ResolveSecret(ctx, oauthConfig.GetClientSecret())
	if err != nil {
		return fmt.Errorf("failed to resolve client_secret: %w", err)
	}

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

	userToken := configv1.UserToken_builder{
		UserId:       proto.String(userID),
		ServiceId:    proto.String(serviceID),
		AccessToken:  proto.String(token.AccessToken),
		RefreshToken: proto.String(token.RefreshToken),
		TokenType:    proto.String(token.TokenType),
		Expiry:       proto.String(token.Expiry.Format(time.RFC3339)),
		Scope:        proto.String(oauthConfig.GetScopes()),
		UpdatedAt:    proto.String(time.Now().Format(time.RFC3339)),
	}.Build()

	// If extra has "scope", use it
	if sc, ok := token.Extra("scope").(string); ok && sc != "" {
		userToken.SetScope(sc)
	}

	if credentialID != "" {
		// Update Credential
		cred.SetToken(userToken)
		if err := storage.SaveCredential(ctx, cred); err != nil {
			return fmt.Errorf("failed to save credential: %w", err)
		}
	} else {
		// Save to UserTokens table
		if err := storage.SaveToken(ctx, userToken); err != nil {
			return fmt.Errorf("failed to save token: %w", err)
		}
	}

	return nil
}
