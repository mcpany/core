package auth

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockOAuth2Server_WellKnownConfiguration(t *testing.T) {
	mockServer := NewMockOAuth2Server(t)
	defer mockServer.Close()

	resp, err := http.Get(mockServer.URL + "/.well-known/openid-configuration")
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var config map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&config)
	assert.NoError(t, err)

	assert.Equal(t, mockServer.URL, config["issuer"])
	assert.Equal(t, mockServer.URL+"/jwks", config["jwks_uri"])
	assert.Equal(t, mockServer.URL+"/auth", config["authorization_endpoint"])
	assert.Equal(t, mockServer.URL+"/token", config["token_endpoint"])
}

func TestMockOAuth2Server_JWKS(t *testing.T) {
	mockServer := NewMockOAuth2Server(t)
	defer mockServer.Close()

	resp, err := http.Get(mockServer.URL + "/jwks")
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var jwks map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&jwks)
	assert.NoError(t, err)

	keys, ok := jwks["keys"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, keys, 1)

	key, ok := keys[0].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "RS256", key["alg"])
	assert.Equal(t, "sig", key["use"])
}

func TestMockOAuth2Server_AuthEndpoint(t *testing.T) {
	mockServer := NewMockOAuth2Server(t)
	defer mockServer.Close()

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// Request with missing redirect_uri
	req, _ := http.NewRequest("GET", mockServer.URL+"/auth?state=xyz", nil)
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// Request with redirect_uri
	req, _ = http.NewRequest("GET", mockServer.URL+"/auth?state=xyz&redirect_uri=http://example.com/callback", nil)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)

	location, err := resp.Location()
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(location.String(), "http://example.com/callback"))
	assert.Equal(t, "mock_code", location.Query().Get("code"))
	assert.Equal(t, "xyz", location.Query().Get("state"))
}

func TestMockOAuth2Server_TokenEndpoint(t *testing.T) {
	mockServer := NewMockOAuth2Server(t)
	mockServer.ClientID = "test_client_id"
	defer mockServer.Close()

	req, _ := http.NewRequest("POST", mockServer.URL+"/token", nil)
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var tokenResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	assert.NoError(t, err)

	assert.NotEmpty(t, tokenResp["access_token"])
	assert.NotEmpty(t, tokenResp["id_token"])
	assert.Equal(t, "Bearer", tokenResp["token_type"])

	// The id_token should be the same as the access_token in this mock implementation
	assert.Equal(t, tokenResp["access_token"], tokenResp["id_token"])
}
