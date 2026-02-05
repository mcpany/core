// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package sqlite

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

func TestStore(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mcpany-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	store := NewStore(db)

	t.Run("SaveAndLoad", func(t *testing.T) {
		svc := configv1.UpstreamServiceConfig_builder{
			Name:    proto.String("test-service"),
			Id:      proto.String("test-id"),
			Version: proto.String("1.0.0"),
		}.Build()
		if err := store.SaveService(context.Background(), svc); err != nil {
			t.Fatalf("failed to save service: %v", err)
		}

		loaded, err := store.GetService(context.Background(), "test-service")
		if err != nil {
			t.Fatalf("failed to get service: %v", err)
		}
		if loaded.GetName() != svc.GetName() {
			t.Errorf("expected name %s, got %s", svc.GetName(), loaded.GetName())
		}
		if loaded.GetId() != svc.GetId() {
			t.Errorf("expected id %s, got %s", svc.GetId(), loaded.GetId())
		}
	})

	t.Run("List", func(t *testing.T) {
		services, err := store.ListServices(context.Background())
		if err != nil {
			t.Fatalf("failed to list services: %v", err)
		}
		if len(services) != 1 {
			t.Errorf("expected 1 service, got %d", len(services))
		}
	})

	t.Run("Update", func(t *testing.T) {
		svc, _ := store.GetService(context.Background(), "test-service")
		// Correct way to update opaque object is to use builder from it?
		// Or creating new builder?
		// Opaque objects are immutable usually?
		// If I can't mutate `svc.Version`, I must create a new object using builder with modified fields.
		// Or use `ToBuilder()` if available?
		// For now, I'll rebuild it since I know the fields.
		// Wait, `svc.Version = ...` was used. Struct fields are likely hidden or read-only effectively?
		// If I can't set `svc.Version`, I need to create a new one.

		newSvc := configv1.UpstreamServiceConfig_builder{
			Name:    proto.String(svc.GetName()),
			Id:      proto.String(svc.GetId()),
			Version: proto.String("1.0.1"),
		}.Build()

		if err := store.SaveService(context.Background(), newSvc); err != nil {
			t.Fatalf("failed to update service: %v", err)
		}
		loaded, _ := store.GetService(context.Background(), "test-service")
		if loaded.GetVersion() != "1.0.1" {
			t.Errorf("expected version 1.0.1, got %s", loaded.GetVersion())
		}
	})

	t.Run("Delete", func(t *testing.T) {
		if err := store.DeleteService(context.Background(), "test-service"); err != nil {
			t.Fatalf("failed to delete service: %v", err)
		}
		loaded, err := store.GetService(context.Background(), "test-service")
		if err != nil {
			t.Fatalf("failed to get service after delete: %v", err)
		}
		if loaded != nil {
			t.Errorf("expected service to be nil, got %v", loaded)
		}
	})

	t.Run("GlobalSettings", func(t *testing.T) {
		level := configv1.GlobalSettings_LOG_LEVEL_DEBUG
		settings := configv1.GlobalSettings_builder{
			LogLevel: &level,
		}.Build()
		err := store.SaveGlobalSettings(context.Background(), settings)
		if err != nil {
			t.Fatalf("failed to save global settings: %v", err)
		}

		loaded, err := store.GetGlobalSettings(context.Background())
		if err != nil {
			t.Fatalf("failed to get global settings: %v", err)
		}
		if loaded.GetLogLevel() != configv1.GlobalSettings_LOG_LEVEL_DEBUG {
			t.Errorf("expected level debug, got %v", loaded.GetLogLevel())
		}
	})

	t.Run("Secrets", func(t *testing.T) {
		secret := configv1.Secret_builder{
			Id:   proto.String("sec-1"),
			Name: proto.String("my-secret"),
			Key:  proto.String("api_key"),
		}.Build()
		err := store.SaveSecret(context.Background(), secret)
		if err != nil {
			t.Fatalf("failed to save secret: %v", err)
		}

		// List
		secrets, err := store.ListSecrets(context.Background())
		if err != nil {
			t.Fatalf("failed to list secrets: %v", err)
		}
		if len(secrets) != 1 {
			t.Errorf("expected 1 secret, got %d", len(secrets))
		}

		// Get
		loaded, err := store.GetSecret(context.Background(), "sec-1")
		if err != nil {
			t.Fatalf("failed to get secret: %v", err)
		}
		if loaded.GetName() != "my-secret" {
			t.Errorf("expected name my-secret, got %s", loaded.GetName())
		}

		// Update
		// secret.Name = proto.String("updated-secret") // Cannot mutate
		updatedSecret := configv1.Secret_builder{
			Id:   proto.String(secret.GetId()),
			Name: proto.String("updated-secret"),
			Key:  proto.String(secret.GetKey()),
		}.Build()

		err = store.SaveSecret(context.Background(), updatedSecret)
		if err != nil {
			t.Fatalf("failed to update secret: %v", err)
		}
		loaded, _ = store.GetSecret(context.Background(), "sec-1")
		if loaded.GetName() != "updated-secret" {
			t.Errorf("expected name updated-secret, got %s", loaded.GetName())
		}

		// Delete
		err = store.DeleteSecret(context.Background(), "sec-1")
		if err != nil {
			t.Fatalf("failed to delete secret: %v", err)
		}
		loaded, _ = store.GetSecret(context.Background(), "sec-1")
		if loaded != nil {
			t.Errorf("expected secret to be nil after delete, got %v", loaded)
		}
	})

	t.Run("Users", func(t *testing.T) {
		user := configv1.User_builder{
			Id:    proto.String("user-1"),
			Roles: []string{"admin"},
		}.Build()
		// Create
		err := store.CreateUser(context.Background(), user)
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		// Get
		got, err := store.GetUser(context.Background(), "user-1")
		if err != nil {
			t.Fatalf("failed to get user: %v", err)
		}
		if got.GetId() != "user-1" {
			t.Errorf("expected id user-1, got %s", got.GetId())
		}

		// List
		users, err := store.ListUsers(context.Background())
		if err != nil {
			t.Fatalf("failed to list users: %v", err)
		}
		if len(users) != 1 {
			t.Errorf("expected 1 user, got %d", len(users))
		}

		// Update
		// user.Roles = []string{"admin", "editor"}
		updatedUser := configv1.User_builder{
			Id:    proto.String("user-1"),
			Roles: []string{"admin", "editor"},
		}.Build()
		err = store.UpdateUser(context.Background(), updatedUser)
		if err != nil {
			t.Fatalf("failed to update user: %v", err)
		}
		got, _ = store.GetUser(context.Background(), "user-1")
		if len(got.GetRoles()) != 2 {
			t.Errorf("expected 2 roles, got %d", len(got.GetRoles()))
		}

		// Update Non-Existent
		err = store.UpdateUser(context.Background(), configv1.User_builder{Id: proto.String("non-existent")}.Build())
		if err == nil {
			t.Error("expected error updating non-existent user, got nil")
		}

		// Delete
		err = store.DeleteUser(context.Background(), "user-1")
		if err != nil {
			t.Fatalf("failed to delete user: %v", err)
		}
		got, err = store.GetUser(context.Background(), "user-1")
		if err != nil {
			t.Fatalf("failed to get user after delete: %v", err)
		}
		if got != nil {
			t.Errorf("expected user to be nil, got %v", got)
		}
	})

	t.Run("Profiles", func(t *testing.T) {
		prof := configv1.ProfileDefinition_builder{
			Name: proto.String("prof-1"),
		}.Build()
		// Save
		err := store.SaveProfile(context.Background(), prof)
		if err != nil {
			t.Fatalf("failed to save profile: %v", err)
		}

		// Get
		got, err := store.GetProfile(context.Background(), "prof-1")
		if err != nil {
			t.Fatalf("failed to get profile: %v", err)
		}
		if got.GetName() != "prof-1" {
			t.Errorf("expected name prof-1, got %s", got.GetName())
		}

		// List
		profs, err := store.ListProfiles(context.Background())
		if err != nil {
			t.Fatalf("failed to list profiles: %v", err)
		}
		if len(profs) != 1 {
			t.Errorf("expected 1 profile, got %d", len(profs))
		}

		// Delete
		err = store.DeleteProfile(context.Background(), "prof-1")
		if err != nil {
			t.Fatalf("failed to delete profile: %v", err)
		}
		got, _ = store.GetProfile(context.Background(), "prof-1")
		if got != nil {
			t.Errorf("expected profile to be nil, got %v", got)
		}
	})

	t.Run("ServiceCollections", func(t *testing.T) {
		col := configv1.Collection_builder{
			Name: proto.String("col-1"),
		}.Build()
		// Save
		err := store.SaveServiceCollection(context.Background(), col)
		if err != nil {
			t.Fatalf("failed to save collection: %v", err)
		}

		// Get
		got, err := store.GetServiceCollection(context.Background(), "col-1")
		if err != nil {
			t.Fatalf("failed to get collection: %v", err)
		}
		if got.GetName() != "col-1" {
			t.Errorf("expected name col-1, got %s", got.GetName())
		}

		// List
		cols, err := store.ListServiceCollections(context.Background())
		if err != nil {
			t.Fatalf("failed to list collections: %v", err)
		}
		if len(cols) != 1 {
			t.Errorf("expected 1 collection, got %d", len(cols))
		}

		// Delete
		err = store.DeleteServiceCollection(context.Background(), "col-1")
		if err != nil {
			t.Fatalf("failed to delete collection: %v", err)
		}
		got, _ = store.GetServiceCollection(context.Background(), "col-1")
		if got != nil {
			t.Errorf("expected collection to be nil, got %v", got)
		}
	})

	t.Run("Credentials", func(t *testing.T) {
		cred := configv1.Credential_builder{
			Id:   proto.String("cred-1"),
			Name: proto.String("my-cred"),
		}.Build()
		// Save
		err := store.SaveCredential(context.Background(), cred)
		if err != nil {
			t.Fatalf("failed to save credential: %v", err)
		}

		// Get
		got, err := store.GetCredential(context.Background(), "cred-1")
		if err != nil {
			t.Fatalf("failed to get credential: %v", err)
		}
		if got.GetName() != "my-cred" {
			t.Errorf("expected name my-cred, got %s", got.GetName())
		}

		// List
		creds, err := store.ListCredentials(context.Background())
		if err != nil {
			t.Fatalf("failed to list credentials: %v", err)
		}
		if len(creds) != 1 {
			t.Errorf("expected 1 credential, got %d", len(creds))
		}

		// Delete
		err = store.DeleteCredential(context.Background(), "cred-1")
		if err != nil {
			t.Fatalf("failed to delete credential: %v", err)
		}
		got, _ = store.GetCredential(context.Background(), "cred-1")
		if got != nil {
			t.Errorf("expected credential to be nil, got %v", got)
		}
	})

	t.Run("Tokens", func(t *testing.T) {
		token := configv1.UserToken_builder{
			UserId:      proto.String("user-1"),
			ServiceId:   proto.String("svc-1"),
			AccessToken: proto.String("xyz"),
		}.Build()
		// Save
		err := store.SaveToken(context.Background(), token)
		if err != nil {
			t.Fatalf("failed to save token: %v", err)
		}

		// Get
		got, err := store.GetToken(context.Background(), "user-1", "svc-1")
		if err != nil {
			t.Fatalf("failed to get token: %v", err)
		}
		if got.GetAccessToken() != "xyz" {
			t.Errorf("expected token xyz, got %s", got.GetAccessToken())
		}

		// Delete
		err = store.DeleteToken(context.Background(), "user-1", "svc-1")
		if err != nil {
			t.Fatalf("failed to delete token: %v", err)
		}
		got, _ = store.GetToken(context.Background(), "user-1", "svc-1")
		if got != nil {
			t.Errorf("expected token to be nil, got %v", got)
		}
	})
}

func TestSaveServiceValidation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mcpany-test-validation-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	store := NewStore(db)

	// Test case: Empty name
	svc := configv1.UpstreamServiceConfig_builder{
		Name: proto.String(""),
		Id:   proto.String("test-id"),
	}.Build()
	err = store.SaveService(context.Background(), svc)
	if err == nil {
		t.Error("expected error for empty service name, got nil")
	} else if err.Error() != "service name is required" {
		t.Errorf("expected error 'service name is required', got '%v'", err)
	}
}

func TestNewDBErrors(t *testing.T) {
	// Test case: Invalid path (directory creation failure)
	// We can try to create a DB in a read-only directory or a non-existent parent that we can't create
	// But in a sandbox, permissions are tricky.
	// We can try using a path that is actually a directory.
	tmpDir, err := os.MkdirTemp("", "mcpany-test-db-errors-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a directory where the file should be
	dbPath := filepath.Join(tmpDir, "directory-as-file")
	if err := os.Mkdir(dbPath, 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	_, err = NewDB(dbPath)
	if err == nil {
		t.Error("expected error when opening a directory as a DB file, got nil")
	}
}

func TestLoadInvalidData(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mcpany-test-invalid-data-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	store := NewStore(db)

	// Manually insert invalid JSON
	_, err = db.Exec("INSERT INTO upstream_services (id, name, config_json) VALUES (?, ?, ?)", "bad-id", "bad-service", "{invalid-json")
	if err != nil {
		t.Fatalf("failed to insert invalid data: %v", err)
	}

	// Try to load
	_, err = store.Load(context.Background())
	if err == nil {
		t.Error("expected error when loading invalid JSON, got nil")
	} else {
		// Error message depends on protojson implementation, but should fail
		t.Logf("Got expected error: %v", err)
	}

	// Try GetService
	_, err = store.GetService(context.Background(), "bad-service")
	if err == nil {
		t.Error("expected error when getting service with invalid JSON, got nil")
	}
}

func TestDBErrors(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mcpany-test-db-closed-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	store := NewStore(db)

	// Close the DB to force errors
	db.Close()

	// Test Load
	_, err = store.Load(context.Background())
	if err == nil {
		t.Error("expected error on Load with closed DB, got nil")
	}

	// Test SaveService
	svc := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service"),
	}.Build()
	err = store.SaveService(context.Background(), svc)
	if err == nil {
		t.Error("expected error on SaveService with closed DB, got nil")
	}

	// Test GetService
	_, err = store.GetService(context.Background(), "test-service")
	if err == nil {
		t.Error("expected error on GetService with closed DB, got nil")
	}

	// Test DeleteService
	err = store.DeleteService(context.Background(), "test-service")
	if err == nil {
		t.Error("expected error on DeleteService with closed DB, got nil")
	}
}
