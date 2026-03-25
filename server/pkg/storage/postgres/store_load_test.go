package postgres

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestStore_Load(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	store := &Store{db: &DB{DB: db}}

	t.Run("Success", func(t *testing.T) {
		user := configv1.User_builder{Id: proto.String("user-1")}.Build()
		userBytes, err := protojson.MarshalOptions{}.Marshal(user)
		require.NoError(t, err)
		mock.ExpectQuery(".*").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(userBytes))

		settings := configv1.GlobalSettings_builder{LogLevel: configv1.GlobalSettings_LOG_LEVEL_INFO.Enum()}.Build()
		settingsBytes, err := protojson.MarshalOptions{}.Marshal(settings)
		require.NoError(t, err)
		mock.ExpectQuery(".*").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(settingsBytes))

		collection := configv1.Collection_builder{Name: proto.String("Collection One")}.Build()
		collectionBytes, err := protojson.MarshalOptions{}.Marshal(collection)
		require.NoError(t, err)
		mock.ExpectQuery(".*").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(collectionBytes))

		service := configv1.UpstreamServiceConfig_builder{Id: proto.String("service-1")}.Build()
		serviceBytes, err := protojson.MarshalOptions{}.Marshal(service)
		require.NoError(t, err)
		mock.ExpectQuery(".*").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(serviceBytes))

		mock.ExpectQuery(".*").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))

		mock.ExpectQuery(".*").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))

		cfg, err := store.Load(context.Background())
		require.NoError(t, err)
		require.NotNil(t, cfg)
		require.Equal(t, configv1.GlobalSettings_LOG_LEVEL_INFO, cfg.GetGlobalSettings().GetLogLevel())
		require.Equal(t, "user-1", cfg.GetUsers()[0].GetId())
		require.Equal(t, "Collection One", cfg.GetCollections()[0].GetName())
		require.Equal(t, "service-1", cfg.GetUpstreamServices()[0].GetId())
	})

	t.Run("Query Error", func(t *testing.T) {
		mock.ExpectQuery(".*").WillReturnError(context.DeadlineExceeded)

		cfg, err := store.Load(context.Background())
		require.Error(t, err)
		require.Nil(t, cfg)
		require.ErrorIs(t, err, context.DeadlineExceeded)
	})
}
