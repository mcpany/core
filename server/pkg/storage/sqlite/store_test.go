package sqlite

import (
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
		svc := &configv1.UpstreamServiceConfig{
			Name:    proto.String("test-service"),
			Id:      proto.String("test-id"),
			Version: proto.String("1.0.0"),
		}
		if err := store.SaveService(svc); err != nil {
			t.Fatalf("failed to save service: %v", err)
		}

		loaded, err := store.GetService("test-service")
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
		services, err := store.ListServices()
		if err != nil {
			t.Fatalf("failed to list services: %v", err)
		}
		if len(services) != 1 {
			t.Errorf("expected 1 service, got %d", len(services))
		}
	})

	t.Run("Update", func(t *testing.T) {
		svc, _ := store.GetService("test-service")
		svc.Version = proto.String("1.0.1")
		if err := store.SaveService(svc); err != nil {
			t.Fatalf("failed to update service: %v", err)
		}
		loaded, _ := store.GetService("test-service")
		if loaded.GetVersion() != "1.0.1" {
			t.Errorf("expected version 1.0.1, got %s", loaded.GetVersion())
		}
	})

	t.Run("Delete", func(t *testing.T) {
		if err := store.DeleteService("test-service"); err != nil {
			t.Fatalf("failed to delete service: %v", err)
		}
		loaded, err := store.GetService("test-service")
		if err != nil {
			t.Fatalf("failed to get service after delete: %v", err)
		}
		if loaded != nil {
			t.Errorf("expected service to be nil, got %v", loaded)
		}
	})
}
