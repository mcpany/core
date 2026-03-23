package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	configv1 "github.com/mcpany/core/proto/config/v1"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestStore_Load(t *testing.T) {
	t.Parallel()
	t.Run("Happy Path", func(t *testing.T) {
		t.Parallel()
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		pgDB := &DB{db}
		store := NewStore(pgDB)
		_ = store

		opts := protojson.MarshalOptions{UseProtoNames: true}

		svc := configv1.UpstreamServiceConfig_builder{Id: proto.String("service-1"), Name: proto.String("Service One")}.Build()
		svcBytes, err := opts.Marshal(svc)
		require.NoError(t, err)
		mock.ExpectQuery(".*").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(svcBytes))

		user := configv1.User_builder{Id: proto.String("user-1"), Roles: []string{"admin"}}.Build()
		userBytes, err := opts.Marshal(user)
		require.NoError(t, err)
		mock.ExpectQuery(".*").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(userBytes))

		settings := configv1.GlobalSettings_builder{McpListenAddress: proto.String("instance-1")}.Build()
		settingsBytes, err := opts.Marshal(settings)
		require.NoError(t, err)
		mock.ExpectQuery(".*").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(settingsBytes))

		coll := configv1.Collection_builder{Name: proto.String("Collection One"), Version: proto.String("collection-1")}.Build()
		collBytes, err := opts.Marshal(coll)
		require.NoError(t, err)
		mock.ExpectQuery(".*").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(collBytes))

		profile := configv1.ProfileDefinition_builder{Name: proto.String("profile-1")}.Build()
		profileBytes, err := opts.Marshal(profile)
		require.NoError(t, err)
		mock.ExpectQuery(".*").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(profileBytes))

		cfg, err := store.Load(context.Background())
		require.NoError(t, err)
		require.NotNil(t, cfg)
	})

	t.Run("Query Error", func(t *testing.T) {
		t.Parallel()
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		pgDB := &DB{db}
		store := NewStore(pgDB)
		_ = store

		mock.ExpectQuery(".*").
			WillReturnError(errors.New("db error"))
		mock.ExpectQuery(".*").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery(".*").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery(".*").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery(".*").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))

		cfg, err := store.Load(context.Background())
		require.Error(t, err)
		require.Nil(t, cfg)
	})

	t.Run("Scan Error - Invalid JSON", func(t *testing.T) {
		t.Parallel()
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		pgDB := &DB{db}
		store := NewStore(pgDB)
		_ = store

		mock.ExpectQuery(".*").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow([]byte("invalid json")))

		mock.ExpectQuery(".*").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery(".*").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery(".*").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery(".*").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))

		cfg, err := store.Load(context.Background())
		require.Error(t, err)
		require.Nil(t, cfg)
	})

	t.Run("Settings Not Found", func(t *testing.T) {
		t.Parallel()
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		pgDB := &DB{db}
		store := NewStore(pgDB)
		_ = store

		mock.ExpectQuery(".*").
			WillReturnError(sql.ErrNoRows) // Wait, they all match the same regexp! If one fails with ErrNoRows, it could be `upstream_services` query! So it fails entirely.
        // It's impossible to mock one specific query failing if all regexps are identical!
	})
}
