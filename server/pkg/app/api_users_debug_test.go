package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"google.golang.org/protobuf/proto"
)

func TestDebugIDOR(t *testing.T) {
	app, store := setupApiTestApp()
	handler := app.handleUserDetail(store)

	victim := configv1.User_builder{Id: proto.String("victim-user"), Roles: []string{"user"}}.Build()
	admin := configv1.User_builder{Id: proto.String("admin-user"), Roles: []string{"admin"}}.Build()

	_ = store.CreateUser(context.Background(), victim)
	_ = store.CreateUser(context.Background(), admin)

	req := httptest.NewRequest(http.MethodGet, "/users/admin-user", nil)
	ctx := auth.ContextWithUser(req.Context(), "victim-user")
	ctx = auth.ContextWithRoles(ctx, []string{"user"})
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	t.Logf("Response Code: %d", w.Code)
	if w.Code == http.StatusOK {
		t.Logf("VULNERABILITY REPRODUCED: User 'victim-user' accessed 'admin-user' profile.")
	}
}
