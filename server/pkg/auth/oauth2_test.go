// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"net/http"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestOAuth2Auth_Authenticate(t *testing.T) {
	t.Run("InvalidTokenURL", func(t *testing.T) {
		clientID := configv1.SecretValue_builder{
			PlainText: proto.String("id"),
		}.Build()
		clientSecret := configv1.SecretValue_builder{
			PlainText: proto.String("secret"),
		}.Build()
		auth := &OAuth2Auth{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			TokenURL:     "invalid-url",
		}
		req, _ := http.NewRequest("GET", "/", nil)
		err := auth.Authenticate(req)
		assert.Error(t, err)
	})
}
