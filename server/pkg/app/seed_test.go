// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/storage"
	"github.com/mcpany/core/server/pkg/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// mockStore is a proxy store to inject test errors to verify retry behavior in clearData and seedData.
type mockStore struct {
	storage.Storage
	// Injected errors
	DeleteServiceErr         error
	DeleteCredentialErr      error
	DeleteSecretErr          error
	DeleteProfileErr         error
	DeleteUserErr            error
	DeleteServiceTemplateErr error

	SaveServiceErr         error
	SaveCredentialErr      error
	SaveSecretErr          error
	SaveProfileErr         error
	CreateUserErr          error
	SaveServiceTemplateErr error

	ListServicesErr         error
	ListCredentialsErr      error
	ListSecretsErr          error
	ListProfilesErr         error
	ListUsersErr            error
	ListServiceTemplatesErr error
}

func (m *mockStore) DeleteService(ctx context.Context, name string) error {
	if m.DeleteServiceErr != nil {
		err := m.DeleteServiceErr
		m.DeleteServiceErr = nil // only fail once to allow retry to succeed
		return err
	}
	return m.Storage.DeleteService(ctx, name)
}
func (m *mockStore) DeleteCredential(ctx context.Context, id string) error {
	if m.DeleteCredentialErr != nil {
		err := m.DeleteCredentialErr
		m.DeleteCredentialErr = nil
		return err
	}
	return m.Storage.DeleteCredential(ctx, id)
}
func (m *mockStore) DeleteSecret(ctx context.Context, id string) error {
	if m.DeleteSecretErr != nil {
		err := m.DeleteSecretErr
		m.DeleteSecretErr = nil
		return err
	}
	return m.Storage.DeleteSecret(ctx, id)
}
func (m *mockStore) DeleteProfile(ctx context.Context, name string) error {
	if m.DeleteProfileErr != nil {
		err := m.DeleteProfileErr
		m.DeleteProfileErr = nil
		return err
	}
	return m.Storage.DeleteProfile(ctx, name)
}
func (m *mockStore) DeleteUser(ctx context.Context, id string) error {
	if m.DeleteUserErr != nil {
		err := m.DeleteUserErr
		m.DeleteUserErr = nil
		return err
	}
	return m.Storage.DeleteUser(ctx, id)
}
func (m *mockStore) DeleteServiceTemplate(ctx context.Context, id string) error {
	if m.DeleteServiceTemplateErr != nil {
		err := m.DeleteServiceTemplateErr
		m.DeleteServiceTemplateErr = nil
		return err
	}
	return m.Storage.DeleteServiceTemplate(ctx, id)
}

func (m *mockStore) SaveService(ctx context.Context, service *configv1.UpstreamServiceConfig) error {
	if m.SaveServiceErr != nil {
		err := m.SaveServiceErr
		m.SaveServiceErr = nil
		return err
	}
	return m.Storage.SaveService(ctx, service)
}
func (m *mockStore) SaveCredential(ctx context.Context, cred *configv1.Credential) error {
	if m.SaveCredentialErr != nil {
		err := m.SaveCredentialErr
		m.SaveCredentialErr = nil
		return err
	}
	return m.Storage.SaveCredential(ctx, cred)
}
func (m *mockStore) SaveSecret(ctx context.Context, secret *configv1.Secret) error {
	if m.SaveSecretErr != nil {
		err := m.SaveSecretErr
		m.SaveSecretErr = nil
		return err
	}
	return m.Storage.SaveSecret(ctx, secret)
}
func (m *mockStore) SaveProfile(ctx context.Context, profile *configv1.ProfileDefinition) error {
	if m.SaveProfileErr != nil {
		err := m.SaveProfileErr
		m.SaveProfileErr = nil
		return err
	}
	return m.Storage.SaveProfile(ctx, profile)
}
func (m *mockStore) CreateUser(ctx context.Context, user *configv1.User) error {
	if m.CreateUserErr != nil {
		err := m.CreateUserErr
		m.CreateUserErr = nil
		return err
	}
	return m.Storage.CreateUser(ctx, user)
}
func (m *mockStore) SaveServiceTemplate(ctx context.Context, tpl *configv1.ServiceTemplate) error {
	if m.SaveServiceTemplateErr != nil {
		err := m.SaveServiceTemplateErr
		m.SaveServiceTemplateErr = nil
		return err
	}
	return m.Storage.SaveServiceTemplate(ctx, tpl)
}

func (m *mockStore) ListServices(ctx context.Context) ([]*configv1.UpstreamServiceConfig, error) {
	if m.ListServicesErr != nil {
		return nil, m.ListServicesErr
	}
	return m.Storage.ListServices(ctx)
}
func (m *mockStore) ListCredentials(ctx context.Context) ([]*configv1.Credential, error) {
	if m.ListCredentialsErr != nil {
		return nil, m.ListCredentialsErr
	}
	return m.Storage.ListCredentials(ctx)
}
func (m *mockStore) ListSecrets(ctx context.Context) ([]*configv1.Secret, error) {
	if m.ListSecretsErr != nil {
		return nil, m.ListSecretsErr
	}
	return m.Storage.ListSecrets(ctx)
}
func (m *mockStore) ListProfiles(ctx context.Context) ([]*configv1.ProfileDefinition, error) {
	if m.ListProfilesErr != nil {
		return nil, m.ListProfilesErr
	}
	return m.Storage.ListProfiles(ctx)
}
func (m *mockStore) ListUsers(ctx context.Context) ([]*configv1.User, error) {
	if m.ListUsersErr != nil {
		return nil, m.ListUsersErr
	}
	return m.Storage.ListUsers(ctx)
}
func (m *mockStore) ListServiceTemplates(ctx context.Context) ([]*configv1.ServiceTemplate, error) {
	if m.ListServiceTemplatesErr != nil {
		return nil, m.ListServiceTemplatesErr
	}
	return m.Storage.ListServiceTemplates(ctx)
}

func TestWithRetry(t *testing.T) {
	ctx := context.Background()
	log := logging.GetLogger()

	t.Run("Success immediately", func(t *testing.T) {
		err := withRetry(ctx, log, func() error { return nil })
		assert.NoError(t, err)
	})

	t.Run("Database is locked, retry success", func(t *testing.T) {
		attempts := 0
		err := withRetry(ctx, log, func() error {
			attempts++
			if attempts == 1 {
				return errors.New("SQLITE_BUSY: database is locked")
			}
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, 2, attempts)
	})

	t.Run("Non-lock error, immediate fail", func(t *testing.T) {
		err := withRetry(ctx, log, func() error {
			return errors.New("other error")
		})
		assert.ErrorContains(t, err, "other error")
	})

	t.Run("Context timeout during retries", func(t *testing.T) {
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
		defer cancel()

		err := withRetry(timeoutCtx, log, func() error {
			return errors.New("database is locked")
		})
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("Max retries reached", func(t *testing.T) {
		// Mock out time.After in our head (it will just take a while ~1.5s total)
		start := time.Now()
		err := withRetry(ctx, log, func() error {
			return errors.New("database is locked")
		})
		assert.ErrorContains(t, err, "max retries reached")
		assert.True(t, time.Since(start) > time.Millisecond) // Ensures it actually tried
	})
}

func TestClearData(t *testing.T) {
	tests := []struct {
		name        string
		setupStore  func(store *mockStore)
		wantErr     bool
		errContains string
	}{
		{
			name:       "Success empty database",
			setupStore: func(store *mockStore) {},
			wantErr:    false,
		},
		{
			name: "Failed to list services",
			setupStore: func(store *mockStore) {
				store.ListServicesErr = errors.New("db error")
			},
			wantErr:     true,
			errContains: "failed to list services",
		},
		{
			name: "Errors listing other resources (logged but not fatal)",
			setupStore: func(store *mockStore) {
				store.ListCredentialsErr = errors.New("cred error")
				store.ListSecretsErr = errors.New("sec error")
				store.ListProfilesErr = errors.New("prof error")
				store.ListUsersErr = errors.New("usr error")
				store.ListServiceTemplatesErr = errors.New("tpl error")
			},
			wantErr: false,
		},
		{
			name: "Deletes all existing resources successfully",
			setupStore: func(store *mockStore) {
				ctx := context.Background()
				store.Storage.SaveService(ctx, configv1.UpstreamServiceConfig_builder{Name: proto.String("s1")}.Build())
				store.Storage.SaveCredential(ctx, configv1.Credential_builder{Id: proto.String("c1")}.Build())
				store.Storage.SaveSecret(ctx, configv1.Secret_builder{Id: proto.String("sec1")}.Build())
				store.Storage.SaveProfile(ctx, configv1.ProfileDefinition_builder{Name: proto.String("p1")}.Build())
				store.Storage.CreateUser(ctx, configv1.User_builder{Id: proto.String("u1")}.Build())
				store.Storage.SaveServiceTemplate(ctx, configv1.ServiceTemplate_builder{Id: proto.String("t1")}.Build())
			},
			wantErr: false,
		},
		{
			name: "Deletes with retry lock recovery",
			setupStore: func(store *mockStore) {
				ctx := context.Background()
				store.Storage.SaveService(ctx, configv1.UpstreamServiceConfig_builder{Name: proto.String("s1")}.Build())
				store.DeleteServiceErr = errors.New("sqlite_busy")
			},
			wantErr: false, // it will recover
		},
		{
			name: "Deletes with unrecoverable error (logged but not fatal)",
			setupStore: func(store *mockStore) {
				ctx := context.Background()
				store.Storage.SaveService(ctx, configv1.UpstreamServiceConfig_builder{Name: proto.String("s1")}.Build())
				store.Storage.SaveCredential(ctx, configv1.Credential_builder{Id: proto.String("c1")}.Build())
				store.Storage.SaveSecret(ctx, configv1.Secret_builder{Id: proto.String("sec1")}.Build())
				store.Storage.SaveProfile(ctx, configv1.ProfileDefinition_builder{Name: proto.String("p1")}.Build())
				store.Storage.CreateUser(ctx, configv1.User_builder{Id: proto.String("u1")}.Build())
				store.Storage.SaveServiceTemplate(ctx, configv1.ServiceTemplate_builder{Id: proto.String("t1")}.Build())

				// The errors are not nil-ed so they fail all retries
				store.DeleteServiceErr = errors.New("hard error 1")
				store.DeleteCredentialErr = errors.New("hard error 2")
				store.DeleteSecretErr = errors.New("hard error 3")
				store.DeleteProfileErr = errors.New("hard error 4")
				store.DeleteUserErr = errors.New("hard error 5")
				store.DeleteServiceTemplateErr = errors.New("hard error 6")
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockStore{Storage: memory.NewStore()}
			tt.setupStore(store)
			app := &Application{Storage: store}

			err := app.clearData(context.Background(), logging.GetLogger())
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.ErrorContains(t, err, tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSeedData(t *testing.T) {
	validSvc, _ := protojson.Marshal(configv1.UpstreamServiceConfig_builder{Name: proto.String("s1")}.Build())
	validCred, _ := protojson.Marshal(configv1.Credential_builder{Id: proto.String("c1")}.Build())
	validSec, _ := protojson.Marshal(configv1.Secret_builder{Id: proto.String("sec1")}.Build())
	validProf, _ := protojson.Marshal(configv1.ProfileDefinition_builder{Name: proto.String("p1")}.Build())
	validUser, _ := protojson.Marshal(configv1.User_builder{Id: proto.String("u1")}.Build())
	validTpl, _ := protojson.Marshal(configv1.ServiceTemplate_builder{Id: proto.String("t1")}.Build())

	tests := []struct {
		name        string
		req         SeedRequest
		setupStore  func(store *mockStore)
		wantErr     bool
		errContains string
	}{
		{
			name:       "Success empty request",
			req:        SeedRequest{},
			setupStore: func(store *mockStore) {},
			wantErr:    false,
		},
		{
			name: "Success all entities",
			req: SeedRequest{
				ServicesRaw:    []json.RawMessage{validSvc},
				CredentialsRaw: []json.RawMessage{validCred},
				SecretsRaw:     []json.RawMessage{validSec},
				ProfilesRaw:    []json.RawMessage{validProf},
				UsersRaw:       []json.RawMessage{validUser},
				TemplatesRaw:   []json.RawMessage{validTpl},
			},
			setupStore: func(store *mockStore) {},
			wantErr:    false,
		},
		{
			name: "Recover from lock when seeding services",
			req:  SeedRequest{ServicesRaw: []json.RawMessage{validSvc}},
			setupStore: func(store *mockStore) {
				store.SaveServiceErr = errors.New("database is locked")
			},
			wantErr: false,
		},
		{
			name:        "Invalid JSON for service",
			req:         SeedRequest{ServicesRaw: []json.RawMessage{[]byte(`{bad json}`)}},
			setupStore:  func(store *mockStore) {},
			wantErr:     true,
			errContains: "invalid json",
		},
		{
			name: "Store save failure for service",
			req:  SeedRequest{ServicesRaw: []json.RawMessage{validSvc}},
			setupStore: func(store *mockStore) {
				store.SaveServiceErr = errors.New("hard error 1")
			},
			wantErr:     true,
			errContains: "failed to save service s1",
		},
		{
			name:        "Invalid JSON for credential",
			req:         SeedRequest{CredentialsRaw: []json.RawMessage{[]byte(`{bad json}`)}},
			setupStore:  func(store *mockStore) {},
			wantErr:     true,
			errContains: "invalid json",
		},
		{
			name: "Store save failure for credential",
			req:  SeedRequest{CredentialsRaw: []json.RawMessage{validCred}},
			setupStore: func(store *mockStore) {
				store.SaveCredentialErr = errors.New("hard error 2")
			},
			wantErr:     true,
			errContains: "failed to save credential",
		},
		{
			name:        "Invalid JSON for secret",
			req:         SeedRequest{SecretsRaw: []json.RawMessage{[]byte(`{bad json}`)}},
			setupStore:  func(store *mockStore) {},
			wantErr:     true,
			errContains: "invalid json",
		},
		{
			name: "Store save failure for secret",
			req:  SeedRequest{SecretsRaw: []json.RawMessage{validSec}},
			setupStore: func(store *mockStore) {
				store.SaveSecretErr = errors.New("hard error 3")
			},
			wantErr:     true,
			errContains: "failed to save secret",
		},
		{
			name:        "Invalid JSON for profile",
			req:         SeedRequest{ProfilesRaw: []json.RawMessage{[]byte(`{bad json}`)}},
			setupStore:  func(store *mockStore) {},
			wantErr:     true,
			errContains: "invalid json",
		},
		{
			name: "Store save failure for profile",
			req:  SeedRequest{ProfilesRaw: []json.RawMessage{validProf}},
			setupStore: func(store *mockStore) {
				store.SaveProfileErr = errors.New("hard error 4")
			},
			wantErr:     true,
			errContains: "failed to save profile",
		},
		{
			name:        "Invalid JSON for user",
			req:         SeedRequest{UsersRaw: []json.RawMessage{[]byte(`{bad json}`)}},
			setupStore:  func(store *mockStore) {},
			wantErr:     true,
			errContains: "invalid json",
		},
		{
			name: "Store save failure for user",
			req:  SeedRequest{UsersRaw: []json.RawMessage{validUser}},
			setupStore: func(store *mockStore) {
				store.CreateUserErr = errors.New("hard error 5")
			},
			wantErr:     true,
			errContains: "failed to create user",
		},
		{
			name:        "Invalid JSON for template",
			req:         SeedRequest{TemplatesRaw: []json.RawMessage{[]byte(`{bad json}`)}},
			setupStore:  func(store *mockStore) {},
			wantErr:     true,
			errContains: "invalid json",
		},
		{
			name: "Store save failure for template",
			req:  SeedRequest{TemplatesRaw: []json.RawMessage{validTpl}},
			setupStore: func(store *mockStore) {
				store.SaveServiceTemplateErr = errors.New("hard error 6")
			},
			wantErr:     true,
			errContains: "failed to save service template",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockStore{Storage: memory.NewStore()}
			tt.setupStore(store)
			app := &Application{Storage: store}

			err := app.seedData(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.ErrorContains(t, err, tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestHandleDebugSeed(t *testing.T) {
	// Setup application with memory storage
	store := &mockStore{Storage: memory.NewStore()}
	app := &Application{
		Storage:     store,
		configPaths: []string{}, // Empty paths to avoid FS usage
	}

	// Pre-populate some data to verify it gets cleared
	ctx := context.Background()
	existingService := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("existing-service"),
	}.Build()
	require.NoError(t, store.SaveService(ctx, existingService))

	// Define seed payload
	newService := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("new-service"),
	}.Build()

	svcBytes, err := protojson.Marshal(newService)
	require.NoError(t, err)

	seedReq := SeedRequest{
		ServicesRaw: []json.RawMessage{svcBytes},
	}
	body, err := json.Marshal(seedReq)
	require.NoError(t, err)

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/api/v1/debug/seed", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// Invoke handler
	app.handleDebugSeed().ServeHTTP(w, req)

	// Verify response
	require.Equal(t, http.StatusOK, w.Code)

	// Verify existing data is gone
	svc, err := store.GetService(ctx, "existing-service")
	require.NoError(t, err)
	require.Nil(t, svc)

	// Verify new data is present
	svc, err = store.GetService(ctx, "new-service")
	require.NoError(t, err)
	require.NotNil(t, svc)
	require.Equal(t, "new-service", svc.GetName())
}

func TestHandleDebugSeed_MethodsAndErrors(t *testing.T) {
	// Setup application with memory storage
	store := &mockStore{Storage: memory.NewStore()}
	app := &Application{
		Storage:     store,
		configPaths: []string{},
	}

	t.Run("GET method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/debug/seed", nil)
		w := httptest.NewRecorder()
		app.handleDebugSeed().ServeHTTP(w, req)
		require.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("Invalid JSON body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/debug/seed", bytes.NewBuffer([]byte(`{bad json}`)))
		w := httptest.NewRecorder()
		app.handleDebugSeed().ServeHTTP(w, req)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Clear data failure", func(t *testing.T) {
		store.ListServicesErr = errors.New("cannot list services")
		req := httptest.NewRequest(http.MethodPost, "/api/v1/debug/seed", bytes.NewBuffer([]byte(`{}`)))
		w := httptest.NewRecorder()
		app.handleDebugSeed().ServeHTTP(w, req)
		require.Equal(t, http.StatusInternalServerError, w.Code)
		require.Contains(t, w.Body.String(), "Failed to clear data")
		store.ListServicesErr = nil
	})

	t.Run("Seed data failure - invalid json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/debug/seed", bytes.NewBuffer([]byte(`{"upstream_services": [ "{bad}" ]}`)))
		w := httptest.NewRecorder()
		app.handleDebugSeed().ServeHTTP(w, req)
		require.Equal(t, http.StatusBadRequest, w.Code)
		require.Contains(t, w.Body.String(), "Invalid JSON in seed data")
	})

	t.Run("Seed data failure - other error", func(t *testing.T) {
		validSvc, _ := protojson.Marshal(configv1.UpstreamServiceConfig_builder{Name: proto.String("s1")}.Build())
		body := SeedRequest{ServicesRaw: []json.RawMessage{validSvc}}
		bodyBytes, _ := json.Marshal(body)

		store.SaveServiceErr = errors.New("hard error db failure")

		req := httptest.NewRequest(http.MethodPost, "/api/v1/debug/seed", bytes.NewBuffer(bodyBytes))
		w := httptest.NewRecorder()
		app.handleDebugSeed().ServeHTTP(w, req)

		require.Equal(t, http.StatusInternalServerError, w.Code)
		require.Contains(t, w.Body.String(), "Failed to seed data: failed to save service s1: hard error db failure")

		store.SaveServiceErr = nil // clean up
	})
}
