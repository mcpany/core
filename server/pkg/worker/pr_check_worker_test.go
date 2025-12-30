// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package worker

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/client"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

func TestPRCheckWorker(t *testing.T) {
	// Setup mock GitHub server
	mux := http.NewServeMux()

	// Mock List PRs
	mux.HandleFunc("/repos/owner/repo/pulls", func(w http.ResponseWriter, r *http.Request) {
		prs := []client.PullRequest{
			{
				Number: 1,
				Title:  "PR 1",
				User:   client.User{Login: "jules-bot"},
				Head:   client.Head{Sha: "sha1"},
			},
			{
				Number: 2,
				Title:  "PR 2",
				User:   client.User{Login: "other"},
				Head:   client.Head{Sha: "sha2"},
			},
		}
		json.NewEncoder(w).Encode(prs)
	})

	// Mock Get User
	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(client.User{Login: "my-bot"})
	})

	// Mock Get Check Runs
	mux.HandleFunc("/repos/owner/repo/commits/sha1/check-runs", func(w http.ResponseWriter, r *http.Request) {
		runs := client.CheckRunsResponse{
			TotalCount: 1,
			CheckRuns: []client.CheckRun{
				{Name: "test-check", Conclusion: "failure"},
			},
		}
		json.NewEncoder(w).Encode(runs)
	})

	// Mock List Comments
	mux.HandleFunc("/repos/owner/repo/issues/1/comments", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			json.NewEncoder(w).Encode([]client.IssueComment{})
		} else if r.Method == http.MethodPost {
			var body map[string]string
			json.NewDecoder(r.Body).Decode(&body)
			if body["body"] == "@jules the git hub actions are failing. Failing github actions: {test-check}." {
				w.WriteHeader(http.StatusCreated)
			} else {
				w.WriteHeader(http.StatusBadRequest)
			}
		}
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	// Create worker with mock config and server URL
	enabled := true
	interval := "100ms"
	owner := "owner"
	repo := "repo"
	cfg := &configv1.PRCheckWorkerConfig{
		Enabled:   &enabled,
		Interval:  &interval,
		RepoOwner: &owner,
		RepoName:  &repo,
	}

	w := NewPRCheckWorker(cfg)
	w.githubClient.BaseURL = server.URL // Override BaseURL for testing

	// Run check manually for deterministic testing
	ctx := context.Background()
	w.runCheck(ctx)

	// Since we can't easily assert internal state without logs or side effects,
	// we rely on the fact that if PostComment wasn't called correctly,
	// it would error out (logged) or the test server would 404/400.
	// For better testing we could inspect the server's received requests.
}

func TestPRCheckWorker_DuplicateComment(t *testing.T) {
	// Setup mock GitHub server
	mux := http.NewServeMux()
	var postCalled bool
	var mu sync.Mutex

	// Mock List PRs
	mux.HandleFunc("/repos/owner/repo/pulls", func(w http.ResponseWriter, r *http.Request) {
		prs := []client.PullRequest{
			{
				Number: 1,
				Title:  "PR 1",
				User:   client.User{Login: "jules-bot"},
				Head:   client.Head{Sha: "sha1"},
			},
		}
		json.NewEncoder(w).Encode(prs)
	})

	// Mock Get User
	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(client.User{Login: "my-bot"})
	})

	// Mock Get Check Runs
	mux.HandleFunc("/repos/owner/repo/commits/sha1/check-runs", func(w http.ResponseWriter, r *http.Request) {
		runs := client.CheckRunsResponse{
			TotalCount: 1,
			CheckRuns: []client.CheckRun{
				{Name: "test-check", Conclusion: "failure"},
			},
		}
		json.NewEncoder(w).Encode(runs)
	})

	// Mock List Comments - return existing comment from us
	mux.HandleFunc("/repos/owner/repo/issues/1/comments", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			json.NewEncoder(w).Encode([]client.IssueComment{
				{User: client.User{Login: "my-bot"}, Body: "some comment"},
			})
		} else if r.Method == http.MethodPost {
			mu.Lock()
			postCalled = true
			mu.Unlock()
			w.WriteHeader(http.StatusCreated)
		}
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	enabled := true
	interval := "100ms"
	owner := "owner"
	repo := "repo"
	cfg := &configv1.PRCheckWorkerConfig{
		Enabled:   &enabled,
		Interval:  &interval,
		RepoOwner: &owner,
		RepoName:  &repo,
	}

	w := NewPRCheckWorker(cfg)
	w.githubClient.BaseURL = server.URL
	w.runCheck(context.Background())

	mu.Lock()
	assert.False(t, postCalled, "Should not post comment if last comment is by us")
	mu.Unlock()
}

func TestPRCheckWorker_Interval(t *testing.T) {
	enabled := true
	interval := "1ms"
	owner := "owner"
	repo := "repo"
	cfg := &configv1.PRCheckWorkerConfig{
		Enabled:   &enabled,
		Interval:  &interval,
		RepoOwner: &owner,
		RepoName:  &repo,
	}

	w := NewPRCheckWorker(cfg)
	// Just ensure Start/Stop doesn't panic
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	w.Start(ctx)
	time.Sleep(10 * time.Millisecond)
	w.Stop()
}
