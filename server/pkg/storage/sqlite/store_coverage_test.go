package sqlite

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSqliteStore_Load_Coverage(t *testing.T) {
	t.Run("UpstreamServicesQueryError", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer db.Close()
		mock.MatchExpectationsInOrder(false)

		mock.ExpectQuery("SELECT config_json FROM upstream_services").WillReturnError(errors.New("upstream error"))
		mock.ExpectQuery("SELECT config_json FROM users").WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM global_settings WHERE id = 1").WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM service_collections").WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM profile_definitions").WillReturnRows(sqlmock.NewRows([]string{"config_json"}))

		store := NewStore(&DB{db})
		_, err = store.Load(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "upstream error")
	})

	t.Run("UsersQueryError", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer db.Close()
		mock.MatchExpectationsInOrder(false)

		mock.ExpectQuery("SELECT config_json FROM upstream_services").WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM users").WillReturnError(errors.New("users error"))
		mock.ExpectQuery("SELECT config_json FROM global_settings WHERE id = 1").WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM service_collections").WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM profile_definitions").WillReturnRows(sqlmock.NewRows([]string{"config_json"}))

		store := NewStore(&DB{db})
		_, err = store.Load(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "users error")
	})

	t.Run("ProfilesQueryError", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer db.Close()
		mock.MatchExpectationsInOrder(false)

		mock.ExpectQuery("SELECT config_json FROM upstream_services").WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM users").WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM global_settings WHERE id = 1").WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM service_collections").WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM profile_definitions").WillReturnError(errors.New("profiles error"))

		store := NewStore(&DB{db})
		_, err = store.Load(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "profiles error")
	})

	t.Run("CollectionsQueryError_Ignored", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer db.Close()
		mock.MatchExpectationsInOrder(false)

		mock.ExpectQuery("SELECT config_json FROM upstream_services").WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM users").WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM global_settings WHERE id = 1").WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM service_collections").WillReturnError(errors.New("collections error"))
		mock.ExpectQuery("SELECT config_json FROM profile_definitions").WillReturnRows(sqlmock.NewRows([]string{"config_json"}))

		store := NewStore(&DB{db})
		_, err = store.Load(context.Background())
		assert.NoError(t, err) // Collections error is ignored in Load()
	})

	t.Run("ScanErrors", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer db.Close()
		mock.MatchExpectationsInOrder(false)

		rows := sqlmock.NewRows([]string{"config_json"}).AddRow(nil) // scan error
		mock.ExpectQuery("SELECT config_json FROM upstream_services").WillReturnRows(rows)
		mock.ExpectQuery("SELECT config_json FROM users").WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM global_settings WHERE id = 1").WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM service_collections").WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM profile_definitions").WillReturnRows(sqlmock.NewRows([]string{"config_json"}))

		store := NewStore(&DB{db})
		_, err = store.Load(context.Background())
		assert.Error(t, err)
	})

	t.Run("UsersScanErrors", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer db.Close()
		mock.MatchExpectationsInOrder(false)

		mock.ExpectQuery("SELECT config_json FROM upstream_services").WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM users").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(nil))
		mock.ExpectQuery("SELECT config_json FROM global_settings WHERE id = 1").WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM service_collections").WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM profile_definitions").WillReturnRows(sqlmock.NewRows([]string{"config_json"}))

		store := NewStore(&DB{db})
		_, err = store.Load(context.Background())
		assert.Error(t, err)
	})

	t.Run("ProfilesScanErrors", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer db.Close()
		mock.MatchExpectationsInOrder(false)

		mock.ExpectQuery("SELECT config_json FROM upstream_services").WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM users").WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM global_settings WHERE id = 1").WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM service_collections").WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM profile_definitions").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(nil))

		store := NewStore(&DB{db})
		_, err = store.Load(context.Background())
		assert.Error(t, err)
	})

	t.Run("CollectionsScanErrors_Ignored", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer db.Close()
		mock.MatchExpectationsInOrder(false)

		mock.ExpectQuery("SELECT config_json FROM upstream_services").WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM users").WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM global_settings WHERE id = 1").WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM service_collections").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(nil))
		mock.ExpectQuery("SELECT config_json FROM profile_definitions").WillReturnRows(sqlmock.NewRows([]string{"config_json"}))

		store := NewStore(&DB{db})
		_, err = store.Load(context.Background())
		assert.NoError(t, err)
	})

	t.Run("UsersUnmarshalErrors", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer db.Close()
		mock.MatchExpectationsInOrder(false)

		mock.ExpectQuery("SELECT config_json FROM upstream_services").WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM users").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow([]byte("invalid json")))
		mock.ExpectQuery("SELECT config_json FROM global_settings WHERE id = 1").WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM service_collections").WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM profile_definitions").WillReturnRows(sqlmock.NewRows([]string{"config_json"}))

		store := NewStore(&DB{db})
		_, err = store.Load(context.Background())
		assert.Error(t, err)
	})

	t.Run("ProfilesUnmarshalErrors", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer db.Close()
		mock.MatchExpectationsInOrder(false)

		mock.ExpectQuery("SELECT config_json FROM upstream_services").WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM users").WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM global_settings WHERE id = 1").WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM service_collections").WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM profile_definitions").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow([]byte("invalid json")))

		store := NewStore(&DB{db})
		_, err = store.Load(context.Background())
		assert.Error(t, err)
	})

	t.Run("UpstreamServicesScanErrors", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer db.Close()
		mock.MatchExpectationsInOrder(false)

		mock.ExpectQuery("SELECT config_json FROM upstream_services").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(nil))
		mock.ExpectQuery("SELECT config_json FROM users").WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM global_settings WHERE id = 1").WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM service_collections").WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM profile_definitions").WillReturnRows(sqlmock.NewRows([]string{"config_json"}))

		store := NewStore(&DB{db})
		_, err = store.Load(context.Background())
		assert.Error(t, err)
	})

	t.Run("UpstreamServicesUnmarshalErrors", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer db.Close()
		mock.MatchExpectationsInOrder(false)

		mock.ExpectQuery("SELECT config_json FROM upstream_services").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow([]byte("invalid json")))
		mock.ExpectQuery("SELECT config_json FROM users").WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM global_settings WHERE id = 1").WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM service_collections").WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM profile_definitions").WillReturnRows(sqlmock.NewRows([]string{"config_json"}))

		store := NewStore(&DB{db})
		_, err = store.Load(context.Background())
		assert.Error(t, err)
	})

	t.Run("Success", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer db.Close()
		mock.MatchExpectationsInOrder(false)

		mock.ExpectQuery("SELECT config_json FROM upstream_services").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow([]byte(`{"name": "test-service"}`)))
		mock.ExpectQuery("SELECT config_json FROM users").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow([]byte(`{"id": "test-user"}`)))
		mock.ExpectQuery("SELECT config_json FROM global_settings WHERE id = 1").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow([]byte(`{"server_name": "test-server"}`)))
		mock.ExpectQuery("SELECT config_json FROM service_collections").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow([]byte(`{"id": "test-collection"}`)))
		mock.ExpectQuery("SELECT config_json FROM profile_definitions").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow([]byte(`{"id": "test-profile"}`)))

		store := NewStore(&DB{db})
		config, err := store.Load(context.Background())
		assert.NoError(t, err)
		assert.NotNil(t, config)
		assert.Len(t, config.GetUpstreamServices(), 1)
		assert.Len(t, config.GetUsers(), 1)
		assert.Len(t, config.GetCollections(), 1)
		assert.Len(t, config.GetGlobalSettings().GetProfileDefinitions(), 1)
	})
}
