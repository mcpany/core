// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func TestValidateGRPCAuth(t *testing.T) {
	am := NewManager()

	t.Run("GlobalAPIKey_NoCredentials", func(t *testing.T) {
		am.SetAPIKey("secret")
		ctx := context.Background()
		_, err := validateGRPCAuth(ctx, am, false)
		require.Error(t, err)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("GlobalAPIKey_ValidCredentials", func(t *testing.T) {
		am.SetAPIKey("secret")
		md := metadata.Pairs("x-api-key", "secret")
		ctx := metadata.NewIncomingContext(context.Background(), md)
		_, err := validateGRPCAuth(ctx, am, false)
		require.NoError(t, err)
	})

	t.Run("GlobalAPIKey_InvalidCredentials", func(t *testing.T) {
		am.SetAPIKey("secret")
		md := metadata.Pairs("x-api-key", "wrong")
		ctx := metadata.NewIncomingContext(context.Background(), md)
		_, err := validateGRPCAuth(ctx, am, false)
		require.Error(t, err)
	})

	t.Run("NoGlobalAuth_PublicIP", func(t *testing.T) {
		am.SetAPIKey("")
		// Need peer info
		ctx := peer.NewContext(context.Background(), &peer.Peer{
			Addr: &net.TCPAddr{IP: net.ParseIP("8.8.8.8"), Port: 1234},
		})
		_, err := validateGRPCAuth(ctx, am, false)
		require.Error(t, err)
	})

	t.Run("NoGlobalAuth_PrivateIP", func(t *testing.T) {
		am.SetAPIKey("")
		ctx := peer.NewContext(context.Background(), &peer.Peer{
			Addr: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 1234},
		})
		_, err := validateGRPCAuth(ctx, am, false)
		require.NoError(t, err)
	})

	t.Run("TrustProxy_XForwardedFor", func(t *testing.T) {
		am.SetAPIKey("")
		// Real connection from public IP (LB), but XFF says private
		ctx := peer.NewContext(context.Background(), &peer.Peer{
			Addr: &net.TCPAddr{IP: net.ParseIP("8.8.8.8"), Port: 1234},
		})
		md := metadata.Pairs("x-forwarded-for", "127.0.0.1, 8.8.8.8")
		ctx = metadata.NewIncomingContext(ctx, md)

		_, err := validateGRPCAuth(ctx, am, true) // Trust proxy enabled
		require.NoError(t, err)
	})

	t.Run("NoTrustProxy_XForwardedFor_Ignored", func(t *testing.T) {
		am.SetAPIKey("")
		// Real connection from public IP (LB), but XFF says private
		ctx := peer.NewContext(context.Background(), &peer.Peer{
			Addr: &net.TCPAddr{IP: net.ParseIP("8.8.8.8"), Port: 1234},
		})
		md := metadata.Pairs("x-forwarded-for", "127.0.0.1, 8.8.8.8")
		ctx = metadata.NewIncomingContext(ctx, md)

		_, err := validateGRPCAuth(ctx, am, false) // Trust proxy disabled
		require.Error(t, err)
	})
}
