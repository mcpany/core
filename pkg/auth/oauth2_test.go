/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
