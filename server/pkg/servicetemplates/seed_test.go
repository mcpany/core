package servicetemplates

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStore is a mock implementation of storage.Storage
type MockStore struct {
	mock.Mock
}

func (m *MockStore) Load(ctx context.Context) (*configv1.McpAnyServerConfig, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*configv1.McpAnyServerConfig), args.Error(1)
}

func (m *MockStore) HasConfigSources() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockStore) SaveService(ctx context.Context, service *configv1.UpstreamServiceConfig) error {
	args := m.Called(ctx, service)
	return args.Error(0)
}

func (m *MockStore) GetService(ctx context.Context, name string) (*configv1.UpstreamServiceConfig, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*configv1.UpstreamServiceConfig), args.Error(1)
}

func (m *MockStore) ListServices(ctx context.Context) ([]*configv1.UpstreamServiceConfig, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*configv1.UpstreamServiceConfig), args.Error(1)
}

func (m *MockStore) DeleteService(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func (m *MockStore) GetGlobalSettings(ctx context.Context) (*configv1.GlobalSettings, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*configv1.GlobalSettings), args.Error(1)
}

func (m *MockStore) SaveGlobalSettings(ctx context.Context, settings *configv1.GlobalSettings) error {
	args := m.Called(ctx, settings)
	return args.Error(0)
}

func (m *MockStore) ListSecrets(ctx context.Context) ([]*configv1.Secret, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*configv1.Secret), args.Error(1)
}

func (m *MockStore) GetSecret(ctx context.Context, id string) (*configv1.Secret, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*configv1.Secret), args.Error(1)
}

func (m *MockStore) SaveSecret(ctx context.Context, secret *configv1.Secret) error {
	args := m.Called(ctx, secret)
	return args.Error(0)
}

func (m *MockStore) ListServiceTemplates(ctx context.Context) ([]*configv1.ServiceTemplate, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*configv1.ServiceTemplate), args.Error(1)
}

func (m *MockStore) GetServiceTemplate(ctx context.Context, id string) (*configv1.ServiceTemplate, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*configv1.ServiceTemplate), args.Error(1)
}

func (m *MockStore) SaveServiceTemplate(ctx context.Context, template *configv1.ServiceTemplate) error {
	args := m.Called(ctx, template)
	return args.Error(0)
}

func (m *MockStore) DeleteServiceTemplate(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockStore) DeleteSecret(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockStore) CreateUser(ctx context.Context, user *configv1.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockStore) GetUser(ctx context.Context, id string) (*configv1.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*configv1.User), args.Error(1)
}

func (m *MockStore) ListUsers(ctx context.Context) ([]*configv1.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*configv1.User), args.Error(1)
}

func (m *MockStore) UpdateUser(ctx context.Context, user *configv1.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockStore) DeleteUser(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockStore) ListProfiles(ctx context.Context) ([]*configv1.ProfileDefinition, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*configv1.ProfileDefinition), args.Error(1)
}

func (m *MockStore) GetProfile(ctx context.Context, name string) (*configv1.ProfileDefinition, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*configv1.ProfileDefinition), args.Error(1)
}

func (m *MockStore) SaveProfile(ctx context.Context, profile *configv1.ProfileDefinition) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *MockStore) DeleteProfile(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func (m *MockStore) ListServiceCollections(ctx context.Context) ([]*configv1.Collection, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*configv1.Collection), args.Error(1)
}

func (m *MockStore) GetServiceCollection(ctx context.Context, name string) (*configv1.Collection, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*configv1.Collection), args.Error(1)
}

func (m *MockStore) SaveServiceCollection(ctx context.Context, collection *configv1.Collection) error {
	args := m.Called(ctx, collection)
	return args.Error(0)
}

func (m *MockStore) DeleteServiceCollection(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func (m *MockStore) SaveToken(ctx context.Context, token *configv1.UserToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockStore) GetToken(ctx context.Context, userID, serviceID string) (*configv1.UserToken, error) {
	args := m.Called(ctx, userID, serviceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*configv1.UserToken), args.Error(1)
}

func (m *MockStore) DeleteToken(ctx context.Context, userID, serviceID string) error {
	args := m.Called(ctx, userID, serviceID)
	return args.Error(0)
}

func (m *MockStore) ListCredentials(ctx context.Context) ([]*configv1.Credential, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*configv1.Credential), args.Error(1)
}

func (m *MockStore) GetCredential(ctx context.Context, id string) (*configv1.Credential, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*configv1.Credential), args.Error(1)
}

func (m *MockStore) SaveCredential(ctx context.Context, cred *configv1.Credential) error {
	args := m.Called(ctx, cred)
	return args.Error(0)
}

func (m *MockStore) DeleteCredential(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockStore) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestSeeder_Seed(t *testing.T) {
	t.Run("successfully seeds built-in templates", func(t *testing.T) {
		tmpDir := t.TempDir()
		mockStore := new(MockStore)
		seeder := &Seeder{
			Store:       mockStore,
			ExamplesDir: tmpDir,
		}

		// Verify that SaveServiceTemplate is called for the expected built-in templates
		// We expect 7 built-in templates: google-calendar, github, gitlab, slack, notion, linear, jira
		expectedTemplates := []string{"google-calendar", "github", "gitlab", "slack", "notion", "linear", "jira"}

		for _, id := range expectedTemplates {
			mockStore.On("SaveServiceTemplate", mock.Anything, mock.MatchedBy(func(tmpl *configv1.ServiceTemplate) bool {
				return tmpl.GetId() == id
			})).Return(nil).Once()
		}

		err := seeder.Seed(context.Background())
		assert.NoError(t, err)

		mockStore.AssertExpectations(t)
	})

	t.Run("handles errors from store", func(t *testing.T) {
		tmpDir := t.TempDir()
		mockStore := new(MockStore)
		seeder := &Seeder{
			Store:       mockStore,
			ExamplesDir: tmpDir,
		}

		// Return error on the first call. We don't know which order they come in, but checking the loop in code:
		// it iterates over built-in templates.
		// Since we want to fail fast, any failure should return error.
		mockStore.On("SaveServiceTemplate", mock.Anything, mock.Anything).Return(assert.AnError).Once()

		err := seeder.Seed(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save template")

		mockStore.AssertExpectations(t)
	})

	t.Run("handles error reading examples directory", func(t *testing.T) {
		mockStore := new(MockStore)
		seeder := &Seeder{
			Store:       mockStore,
			ExamplesDir: "/tmp/non-existent-dir",
		}

		err := seeder.Seed(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read examples dir")
	})

	t.Run("processes files in examples directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		// Create a subdirectory
		err := os.Mkdir(filepath.Join(tmpDir, "example-service"), 0755)
		assert.NoError(t, err)
		// Create config.yaml
		err = os.WriteFile(filepath.Join(tmpDir, "example-service", "config.yaml"), []byte("upstream_services: []"), 0644)
		assert.NoError(t, err)

		mockStore := new(MockStore)
		seeder := &Seeder{
			Store:       mockStore,
			ExamplesDir: tmpDir,
		}

		// Expectations for built-in templates still apply
		expectedTemplates := []string{"google-calendar", "github", "gitlab", "slack", "notion", "linear", "jira"}
		for _, id := range expectedTemplates {
			mockStore.On("SaveServiceTemplate", mock.Anything, mock.MatchedBy(func(tmpl *configv1.ServiceTemplate) bool {
				return tmpl.GetId() == id
			})).Return(nil).Once()
		}

		err = seeder.Seed(context.Background())
		assert.NoError(t, err)
		mockStore.AssertExpectations(t)
	})

	t.Run("handles malformed yaml in examples", func(t *testing.T) {
		tmpDir := t.TempDir()
		err := os.Mkdir(filepath.Join(tmpDir, "bad-service"), 0755)
		assert.NoError(t, err)
		err = os.WriteFile(filepath.Join(tmpDir, "bad-service", "config.yaml"), []byte("invalid yaml: :"), 0644)
		assert.NoError(t, err)

		mockStore := new(MockStore)
		seeder := &Seeder{Store: mockStore, ExamplesDir: tmpDir}

		// Built-ins still saved
		expectedTemplates := []string{"google-calendar", "github", "gitlab", "slack", "notion", "linear", "jira"}
		for _, id := range expectedTemplates {
			mockStore.On("SaveServiceTemplate", mock.Anything, mock.MatchedBy(func(tmpl *configv1.ServiceTemplate) bool {
				return tmpl.GetId() == id
			})).Return(nil).Once()
		}

		err = seeder.Seed(context.Background())
		assert.NoError(t, err) // Should continue
		mockStore.AssertExpectations(t)
	})
}
