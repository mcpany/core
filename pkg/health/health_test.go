/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package health

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestCheckHTTP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	checker := NewChecker()
	hcc := configv1.HttpHealthCheck_builder{}.Build()
	hcc.SetUrl(server.URL)
	hcc.SetExpectedCode(http.StatusOK)
	hcc.SetTimeout(durationpb.New(5 * time.Second))
	err := checker.checkHTTP(context.Background(), hcc)

	if err != nil {
		t.Errorf("Expected health check to pass, but got error: %v", err)
	}
}
