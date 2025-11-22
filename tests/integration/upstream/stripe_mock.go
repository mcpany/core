// Copyright 2024 Author(s) of MCP Any
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package upstream

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mcpany/core/tests/integration"
	"github.com/stretchr/testify/require"
)

func BuildStripeMockServer(t *testing.T) *integration.ManagedProcess {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "Bearer test", r.Header.Get("Authorization"))
		if r.URL.Path == "/v1/customers" {
			var data map[string]interface{}
			err := json.NewDecoder(r.Body).Decode(&data)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			data["id"] = "cus_12345"
			data["object"] = "customer"
			data["address"] = nil
			data["balance"] = 0
			data["created"] = time.Now().Unix()
			data["currency"] = "usd"
			data["default_source"] = nil
			data["delinquent"] = false
			data["description"] = "test customer"
			data["discount"] = nil
			data["invoice_prefix"] = "INV"
			data["invoice_settings"] = map[string]interface{}{}
			data["livemode"] = false
			data["metadata"] = map[string]interface{}{}
			data["next_invoice_sequence"] = 1
			data["phone"] = nil
			data["preferred_locales"] = []string{}
			data["shipping"] = nil
			data["tax_exempt"] = "none"

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(data)
		} else {
			http.NotFound(w, r)
		}
	}))

	return integration.NewManagedProcess(t, "stripe_mock_server", server.URL, nil, nil)
}
