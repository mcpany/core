package util

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestSanitizeUser_New(t *testing.T) {
	t.Parallel()

	t.Run("nil user", func(t *testing.T) {
		assert.Nil(t, SanitizeUser(nil))
	})

	t.Run("user with no authentication", func(t *testing.T) {
		u := &configv1.User{
			Id: proto.String("test"),
		}
		sanitized := SanitizeUser(u)
		assert.Equal(t, u.Id, sanitized.Id)
	})

	t.Run("user with authentication", func(t *testing.T) {
		u := &configv1.User{
			Id: proto.String("test"),
			Authentication: &configv1.Authentication{
				AuthMethod: &configv1.Authentication_ApiKey{
					ApiKey: &configv1.APIKeyAuth{
						Value: &configv1.SecretValue{
							Value: &configv1.SecretValue_PlainText{PlainText: "secret"},
						},
					},
				},
			},
		}
		sanitized := SanitizeUser(u)
		assert.Equal(t, "REDACTED", sanitized.Authentication.GetApiKey().GetValue().GetPlainText())
	})
}

func TestSanitizeCredential_New(t *testing.T) {
	t.Parallel()

	t.Run("nil credential", func(t *testing.T) {
		assert.Nil(t, SanitizeCredential(nil))
	})

	t.Run("credential with all fields", func(t *testing.T) {
		c := &configv1.Credential{
			Authentication: &configv1.Authentication{
				AuthMethod: &configv1.Authentication_BearerToken{
					BearerToken: &configv1.BearerTokenAuth{
						Token: &configv1.SecretValue{
							Value: &configv1.SecretValue_PlainText{PlainText: "secret"},
						},
					},
				},
			},
			Token: &configv1.UserToken{
				AccessToken: proto.String("access"),
				RefreshToken: proto.String("refresh"),
			},
		}
		sanitized := SanitizeCredential(c)
		assert.Equal(t, "REDACTED", sanitized.Authentication.GetBearerToken().GetToken().GetPlainText())
		assert.Equal(t, "REDACTED", sanitized.Token.GetAccessToken())
		assert.Equal(t, "REDACTED", sanitized.Token.GetRefreshToken())
	})
}

func TestSanitizeAuthentication_New(t *testing.T) {
	t.Parallel()

	t.Run("nil authentication", func(t *testing.T) {
		assert.Nil(t, SanitizeAuthentication(nil))
	})

	t.Run("api key with verification value", func(t *testing.T) {
		a := &configv1.Authentication{
			AuthMethod: &configv1.Authentication_ApiKey{
				ApiKey: &configv1.APIKeyAuth{
					Value: &configv1.SecretValue{
						Value: &configv1.SecretValue_PlainText{PlainText: "secret"},
					},
					VerificationValue: proto.String("verification"),
				},
			},
		}
		sanitized := SanitizeAuthentication(a)
		assert.Equal(t, "REDACTED", sanitized.GetApiKey().GetValue().GetPlainText())
		assert.Equal(t, "REDACTED", sanitized.GetApiKey().GetVerificationValue())
	})

	t.Run("basic auth with password hash", func(t *testing.T) {
		a := &configv1.Authentication{
			AuthMethod: &configv1.Authentication_BasicAuth{
				BasicAuth: &configv1.BasicAuth{
					Password: &configv1.SecretValue{
						Value: &configv1.SecretValue_PlainText{PlainText: "secret"},
					},
					PasswordHash: proto.String("hash"),
				},
			},
		}
		sanitized := SanitizeAuthentication(a)
		assert.Equal(t, "REDACTED", sanitized.GetBasicAuth().GetPassword().GetPlainText())
		assert.Equal(t, "REDACTED", sanitized.GetBasicAuth().GetPasswordHash())
	})

	t.Run("oauth2", func(t *testing.T) {
		a := &configv1.Authentication{
			AuthMethod: &configv1.Authentication_Oauth2{
				Oauth2: &configv1.OAuth2Auth{
					ClientId: &configv1.SecretValue{
						Value: &configv1.SecretValue_PlainText{PlainText: "id"},
					},
					ClientSecret: &configv1.SecretValue{
						Value: &configv1.SecretValue_PlainText{PlainText: "secret"},
					},
				},
			},
		}
		sanitized := SanitizeAuthentication(a)
		assert.Equal(t, "REDACTED", sanitized.GetOauth2().GetClientId().GetPlainText())
		assert.Equal(t, "REDACTED", sanitized.GetOauth2().GetClientSecret().GetPlainText())
	})

	t.Run("trusted header", func(t *testing.T) {
		a := &configv1.Authentication{
			AuthMethod: &configv1.Authentication_TrustedHeader{
				TrustedHeader: &configv1.TrustedHeaderAuth{
					HeaderValue: proto.String("secret"),
				},
			},
		}
		sanitized := SanitizeAuthentication(a)
		assert.Equal(t, "REDACTED", sanitized.GetTrustedHeader().GetHeaderValue())
	})
}

func TestSanitizeSecretValue_New(t *testing.T) {
	t.Parallel()

	t.Run("remote content", func(t *testing.T) {
		s := &configv1.SecretValue{
			Value: &configv1.SecretValue_RemoteContent{
				RemoteContent: &configv1.RemoteContent{
					Auth: &configv1.Authentication{
						AuthMethod: &configv1.Authentication_ApiKey{
							ApiKey: &configv1.APIKeyAuth{
								Value: &configv1.SecretValue{
									Value: &configv1.SecretValue_PlainText{PlainText: "secret"},
								},
							},
						},
					},
				},
			},
		}
		sanitized := SanitizeSecretValue(s)
		assert.Equal(t, "REDACTED", sanitized.GetRemoteContent().GetAuth().GetApiKey().GetValue().GetPlainText())
	})

    t.Run("remote content without auth", func(t *testing.T) {
		s := &configv1.SecretValue{
			Value: &configv1.SecretValue_RemoteContent{
				RemoteContent: &configv1.RemoteContent{
                    HttpUrl: proto.String("http://example.com"),
				},
			},
		}
		sanitized := SanitizeSecretValue(s)
		assert.Nil(t, sanitized.GetRemoteContent().GetAuth())
	})

	t.Run("vault", func(t *testing.T) {
		s := &configv1.SecretValue{
			Value: &configv1.SecretValue_Vault{
				Vault: &configv1.VaultSecret{
					Token: &configv1.SecretValue{
						Value: &configv1.SecretValue_PlainText{PlainText: "secret"},
					},
				},
			},
		}
		sanitized := SanitizeSecretValue(s)
		assert.Equal(t, "REDACTED", sanitized.GetVault().GetToken().GetPlainText())
	})
}

func TestResolveSecretRecursive_Depth_New(t *testing.T) {
	t.Parallel()

	depth := 12
	root := &configv1.SecretValue{
		Value: &configv1.SecretValue_RemoteContent{
			RemoteContent: &configv1.RemoteContent{
				HttpUrl: proto.String("http://example.com"),
				Auth: &configv1.Authentication{
					AuthMethod: &configv1.Authentication_BearerToken{
						BearerToken: &configv1.BearerTokenAuth{
							Token: nil, // will be set in loop
						},
					},
				},
			},
		},
	}

	current := root
	for i := 0; i < depth; i++ {
		next := &configv1.SecretValue{
			Value: &configv1.SecretValue_RemoteContent{
				RemoteContent: &configv1.RemoteContent{
					HttpUrl: proto.String("http://example.com"),
					Auth: &configv1.Authentication{
						AuthMethod: &configv1.Authentication_BearerToken{
							BearerToken: &configv1.BearerTokenAuth{
								Token: nil,
							},
						},
					},
				},
			},
		}

		// Set the token of current to next
		current.GetRemoteContent().GetAuth().GetBearerToken().Token = next
		current = next
	}
	// Terminate with plain text
	current.Value = &configv1.SecretValue_PlainText{PlainText: "secret"}

	_, err := ResolveSecret(context.Background(), root)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "max recursion depth")
}

func TestResolveSecret_Errors_New(t *testing.T) {
	t.Parallel()

	t.Run("env var not set", func(t *testing.T) {
		s := &configv1.SecretValue{
			Value: &configv1.SecretValue_EnvironmentVariable{
				EnvironmentVariable: "NON_EXISTENT_ENV_VAR_12345",
			},
		}
		_, err := ResolveSecret(context.Background(), s)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "environment variable")
	})

	t.Run("file path invalid", func(t *testing.T) {
		s := &configv1.SecretValue{
			Value: &configv1.SecretValue_FilePath{
				FilePath: "non_existent_file_12345.txt",
			},
		}
		_, err := ResolveSecret(context.Background(), s)
		assert.Error(t, err)
	})
}

func TestResolveSecretMap_New(t *testing.T) {
	t.Parallel()

	secretMap := map[string]*configv1.SecretValue{
		"KEY1": {
			Value: &configv1.SecretValue_PlainText{PlainText: "value1"},
		},
	}
	plainMap := map[string]string{
		"KEY2": "value2",
	}

	resolved, err := ResolveSecretMap(context.Background(), secretMap, plainMap)
	assert.NoError(t, err)
	assert.Equal(t, "value1", resolved["KEY1"])
	assert.Equal(t, "value2", resolved["KEY2"])
}

func TestRedactSlice_New(t *testing.T) {
    t.Parallel()

    t.Run("slice with sensitive map", func(t *testing.T) {
        input := []interface{}{
            map[string]interface{}{"api_key": "secret"},
            "plain",
        }
        output := redactSlice(input)

        assert.Equal(t, "[REDACTED]", output[0].(map[string]interface{})["api_key"])
        assert.Equal(t, "plain", output[1])
    })

    t.Run("nested slice", func(t *testing.T) {
        input := []interface{}{
            []interface{}{
                map[string]interface{}{"password": "secret"},
            },
        }
        output := redactSlice(input)

        nested := output[0].([]interface{})
        assert.Equal(t, "[REDACTED]", nested[0].(map[string]interface{})["password"])
    })

     t.Run("no change", func(t *testing.T) {
        input := []interface{}{"a", "b"}
        output := redactSlice(input)

        assert.Equal(t, "a", output[0])
        assert.Equal(t, "b", output[1])
        // Slice is deep copied, so output is not input
        assert.NotEqual(t, &input, &output)
    })

    t.Run("deep copy nested", func(t *testing.T) {
        // Nested structures that are NOT sensitive.
        // This forces redactSlice to iterate and deep copy.
         input := []interface{}{
            map[string]interface{}{"safe": "val"},
            []interface{}{"safe_slice"},
            "plain",
        }
        output := redactSlice(input)

        // Ensure values are preserved
        assert.Equal(t, "val", output[0].(map[string]interface{})["safe"])
        assert.Equal(t, "safe_slice", output[1].([]interface{})[0])
        assert.Equal(t, "plain", output[2])

        // Output slice is different
        assert.NotEqual(t, &input, &output)

        // Inner map is same reference (RedactMap optimization)
        assert.Equal(t, input[0], output[0])

        // Inner slice is different reference (redactSlice deep copy logic)
        assert.NotEqual(t, &input[1], &output[1])
    })
}

func TestIsKeyColon_New(t *testing.T) {
    t.Parallel()

    t.Run("basic", func(t *testing.T) {
        input := []byte(`"key": "value"`)
        // end of "key" is at index 5 (after quote)
        assert.True(t, isKeyColon(input, 5))
    })

    t.Run("with whitespace", func(t *testing.T) {
        input := []byte(`"key"   : "value"`)
        assert.True(t, isKeyColon(input, 5))
    })

    t.Run("no colon", func(t *testing.T) {
        input := []byte(`"key" "value"`)
        assert.False(t, isKeyColon(input, 5))
    })

    t.Run("end of string", func(t *testing.T) {
        input := []byte(`"key"`)
        assert.False(t, isKeyColon(input, 5))
    })
}

func TestSanitizeUserToken_Partial(t *testing.T) {
    t.Parallel()

    t.Run("only access token", func(t *testing.T) {
        token := &configv1.UserToken{
            AccessToken: proto.String("access"),
        }
        sanitized := SanitizeUserToken(token)
        assert.Equal(t, "REDACTED", sanitized.GetAccessToken())
        assert.Empty(t, sanitized.GetRefreshToken())
    })

     t.Run("only refresh token", func(t *testing.T) {
        token := &configv1.UserToken{
            RefreshToken: proto.String("refresh"),
        }
        sanitized := SanitizeUserToken(token)
        assert.Empty(t, sanitized.GetAccessToken())
        assert.Equal(t, "REDACTED", sanitized.GetRefreshToken())
    })
}
