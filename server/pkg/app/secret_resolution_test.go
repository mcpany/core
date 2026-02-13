package app

import (
	"context"
	"fmt"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/storage/sqlite"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/stretchr/testify/require"
    "google.golang.org/protobuf/proto"
)

func TestSecretResolutionIntegration(t *testing.T) {
	// 1. Setup InMemory DB
	db, err := sqlite.NewDB(":memory:")
	require.NoError(t, err)
	defer db.Close()
	store := sqlite.NewStore(db)

	app := NewApplication()

	// Initialize Schema
	err = app.initializeDatabase(context.Background(), store)
	require.NoError(t, err)

	app.Storage = store

	// 2. Create Secret using Builder
	secret := configv1.Secret_builder{
		Id:    proto.String("secret-123"),
		Name:  proto.String("Test Secret"),
		Key:   proto.String("MY_API_KEY"),
		Value: proto.String("super-secret-value"),
	}.Build()

	err = store.SaveSecret(context.Background(), secret)
	require.NoError(t, err)

	// 3. Setup Context
	ctx := context.Background()
	ctx = util.WithSecretProvider(ctx, func(ctx context.Context, key string) (string, error) {
		if secret, err := app.Storage.GetSecret(ctx, key); err == nil && secret != nil {
			return secret.GetValue(), nil
		}
		secrets, err := app.Storage.ListSecrets(ctx)
		if err != nil {
			return "", err
		}
		for _, s := range secrets {
			if s.GetKey() == key || s.GetName() == key {
				return s.GetValue(), nil
			}
		}
		return "", fmt.Errorf("secret not found: %s", key)
	})

	// 4. Resolve by Key
	sv := configv1.SecretValue_builder{
        PlainText: proto.String("${MY_API_KEY}"),
	}.Build()

	resolved, err := util.ResolveSecret(ctx, sv)
	require.NoError(t, err)
	require.Equal(t, "super-secret-value", resolved)

	// 5. Resolve by ID
	svID := configv1.SecretValue_builder{
        PlainText: proto.String("${secret-123}"),
	}.Build()

	resolvedID, err := util.ResolveSecret(ctx, svID)
	require.NoError(t, err)
	require.Equal(t, "super-secret-value", resolvedID)

	// 6. Verify non-existent
	svMiss := configv1.SecretValue_builder{
        PlainText: proto.String("${MISSING}"),
	}.Build()

	resolvedMiss, err := util.ResolveSecret(ctx, svMiss)
	require.NoError(t, err)
	require.Equal(t, "${MISSING}", resolvedMiss)
}
