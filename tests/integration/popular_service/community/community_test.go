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

package community_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/mcpany/core/tests/framework"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommunityServices(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		t.Skip("Skipping test in GitHub Actions due to network restrictions")
	}

	discordBotToken := os.Getenv("DISCORD_BOT_TOKEN")
	discordChannelID := os.Getenv("DISCORD_CHANNEL_ID")
	novuAPIKey := os.Getenv("NOVU_API_KEY")
	novuSubscriberID := os.Getenv("NOVU_SUBSCRIBER_ID")
	twilioAccountSID := os.Getenv("TWILIO_ACCOUNT_SID")
	twilioAuthToken := os.Getenv("TWILIO_AUTH_TOKEN")
	twilioFromNumber := os.Getenv("TWILIO_FROM_NUMBER")
	twilioToNumber := os.Getenv("TWILIO_TO_NUMBER")

	if discordBotToken == "" || discordChannelID == "" || novuAPIKey == "" || novuSubscriberID == "" || twilioAccountSID == "" || twilioAuthToken == "" || twilioFromNumber == "" || twilioToNumber == "" {
		t.Skip("Skipping test due to missing environment variables")
	}

	testCase := &framework.E2ETestCase{
		Name: "CommunityServices",
		RegistrationMethods: []framework.RegistrationMethod{framework.FileRegistration},
		GenerateUpstreamConfig: func(upstreamEndpoint string) string {
			return fmt.Sprintf(`
upstream_services:
  - name: "discord"
    http_service:
      address: "https://discord.com/api/v10"
      authentication:
        token: "Bot %s"
      calls:
        - endpoint_path: "/channels/%s/messages"
          method: "HTTP_METHOD_POST"
          schema:
            name: "send_message"
            input_schema:
              type: "object"
              properties:
                content:
                  type: "string"
                  description: "The content of the message to send."
  - name: "novu"
    http_service:
      address: "https://api.novu.co/v1"
      authentication:
        token: "ApiKey %s"
      calls:
        - endpoint_path: "/events/trigger"
          method: "HTTP_METHOD_POST"
          schema:
            name: "trigger_event"
            input_schema:
              type: "object"
              properties:
                name:
                  type: "string"
                  description: "The name of the event to trigger."
                to:
                  type: "object"
                  properties:
                    subscriberId:
                      type: "string"
                      description: "The ID of the subscriber to send the notification to."
                payload:
                  type: "object"
                  description: "The payload to send with the event."
  - name: "twilio"
    http_service:
      address: "https://api.twilio.com/2010-04-01"
      authentication:
        basic:
          username: "%s"
          password: "%s"
      calls:
        - endpoint_path: "/Accounts/%s/Messages.json"
          method: "HTTP_METHOD_POST"
          content_type: "application/x-www-form-urlencoded"
          schema:
            name: "send_sms"
            input_schema:
              type: "object"
              properties:
                To:
                  type: "string"
                  description: "The destination phone number in E.164 format."
                From:
                  type: "string"
                  description: "A Twilio phone number in E.164 format, an alphanumeric sender ID, or a Channel Endpoint address that is enabled for the type of message you want to send."
                Body:
                  type: "string"
                  description: "The text of the message you want to send."
`, discordBotToken, discordChannelID, novuAPIKey, twilioAccountSID, twilioAuthToken, twilioAccountSID)
		},
		InvokeAIClient: func(t *testing.T, mcpanyEndpoint string) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client := mcp.NewClient(&mcp.Implementation{Name: "e2e-test-client"}, nil)
			transport := &mcp.StreamableClientTransport{
				Endpoint: mcpanyEndpoint,
			}
			session, err := client.Connect(ctx, transport, nil)
			require.NoError(t, err)
			defer session.Close()

			t.Run("DiscordSendMessage", func(t *testing.T) {
				args := map[string]interface{}{
					"content": "Hello from mcpany e2e test!",
				}
				argsJSON, err := json.Marshal(args)
				require.NoError(t, err)

				resp, err := session.CallTool(ctx, &mcp.CallToolParams{
					Name:      "discord.send_message",
					Arguments: argsJSON,
				})
				require.NoError(t, err)
				assert.NotEmpty(t, resp)
			})

			t.Run("NovuTriggerEvent", func(t *testing.T) {
				args := map[string]interface{}{
					"name": "test-event",
					"to": map[string]string{
						"subscriberId": novuSubscriberID,
					},
					"payload": map[string]string{
						"message": "Hello from mcpany e2e test!",
					},
				}
				argsJSON, err := json.Marshal(args)
				require.NoError(t, err)

				resp, err := session.CallTool(ctx, &mcp.CallToolParams{
					Name:      "novu.trigger_event",
					Arguments: argsJSON,
				})
				require.NoError(t, err)
				assert.NotEmpty(t, resp)
			})

			t.Run("TwilioSendSMS", func(t *testing.T) {
				args := map[string]interface{}{
					"To":   twilioToNumber,
					"From": twilioFromNumber,
					"Body": fmt.Sprintf("Hello from mcpany e2e test! %d", time.Now().Unix()),
				}
				argsJSON, err := json.Marshal(args)
				require.NoError(t, err)

				resp, err := session.CallTool(ctx, &mcp.CallToolParams{
					Name:      "twilio.send_sms",
					Arguments: argsJSON,
				})
				require.NoError(t, err)
				assert.NotEmpty(t, resp)
			})
		},
	}
	framework.RunE2ETest(t, testCase)
}
