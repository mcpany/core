// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package worker

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/mcpany/core/pkg/client"
	"github.com/mcpany/core/pkg/logging"
	configv1 "github.com/mcpany/core/proto/config/v1"
)

// PRCheckWorker checks PRs for failing checks and posts comments.
type PRCheckWorker struct {
	githubClient *client.GitHubClient
	config       *configv1.PRCheckWorkerConfig
	stopCh       chan struct{}
	wg           sync.WaitGroup
	repoOwner    string
	repoName     string
}

// NewPRCheckWorker creates a new PRCheckWorker.
func NewPRCheckWorker(cfg *configv1.PRCheckWorkerConfig) *PRCheckWorker {
	// If config is nil, provide a default or empty config
	if cfg == nil {
		enabled := true
		interval := "10m"
		cfg = &configv1.PRCheckWorkerConfig{
			Enabled:  &enabled,
			Interval: &interval,
		}
	}

	return &PRCheckWorker{
		githubClient: client.NewGitHubClient(cfg.GetGithubToken()),
		config:       cfg,
		stopCh:       make(chan struct{}),
		repoOwner:    cfg.GetRepoOwner(),
		repoName:     cfg.GetRepoName(),
	}
}

// Start starts the worker.
func (w *PRCheckWorker) Start(ctx context.Context) {
	if !w.config.GetEnabled() {
		return
	}

	// If repo owner/name are not configured, we can't proceed.
	if w.repoOwner == "" || w.repoName == "" {
		logging.GetLogger().Warn("PRCheckWorker: Repo owner or name not configured, worker will not run")
		return
	}

	interval, err := time.ParseDuration(w.config.GetInterval())
	if err != nil {
		logging.GetLogger().Error("Invalid interval for PRCheckWorker", "error", err, "interval", w.config.GetInterval())
		interval = 10 * time.Minute
	}

	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		logging.GetLogger().Info("PRCheckWorker started", "interval", interval)

		// Run immediately on start
		w.runCheck(ctx)

		for {
			select {
			case <-ctx.Done():
				return
			case <-w.stopCh:
				return
			case <-ticker.C:
				w.runCheck(ctx)
			}
		}
	}()
}

// Stop stops the worker.
func (w *PRCheckWorker) Stop() {
	close(w.stopCh)
	w.wg.Wait()
}

func (w *PRCheckWorker) runCheck(ctx context.Context) {
	log := logging.GetLogger().With("worker", "PRCheckWorker")

	prs, err := w.githubClient.ListOpenPRs(ctx, w.repoOwner, w.repoName)
	if err != nil {
		log.Error("Failed to list open PRs", "error", err)
		return
	}

	me, err := w.githubClient.GetUser(ctx)
	if err != nil {
		log.Error("Failed to get current user", "error", err)
		return
	}
	myLogin := me.Login

	for _, pr := range prs {
		// Filter by creator "Jules bot" - assuming username contains "jules" or is configurable?
		// The requirement says "opened by Jules bot". Let's assume the username is "jules-bot" or similar.
		// Or maybe we just check all PRs?
		// "scan the open PRs opened by Jules bot"
		// I will check if the PR user login matches "jules-bot" or contains "jules" (safer to match exact or config).
		// For now, I'll check if the login is "jules-bot" or "Jules".
		if !isJulesBot(pr.User.Login) {
			continue
		}

		checkRuns, err := w.githubClient.GetCheckRuns(ctx, w.repoOwner, w.repoName, pr.Head.Sha)
		if err != nil {
			log.Error("Failed to get check runs", "pr", pr.Number, "error", err)
			continue
		}

		var failingChecks []string
		for _, check := range checkRuns.CheckRuns {
			if check.Conclusion == "failure" || check.Conclusion == "timed_out" || check.Conclusion == "cancelled" {
				failingChecks = append(failingChecks, check.Name)
			}
		}

		if len(failingChecks) > 0 {
			// Check if we already commented
			comments, err := w.githubClient.ListComments(ctx, w.repoOwner, w.repoName, pr.Number)
			if err != nil {
				log.Error("Failed to list comments", "pr", pr.Number, "error", err)
				continue
			}

			// "when the last comment is already made by us, do not post new comments"
			if len(comments) > 0 {
				lastComment := comments[len(comments)-1]
				if lastComment.User.Login == myLogin {
					log.Info("Last comment is already by us, skipping", "pr", pr.Number)
					continue
				}
			}

			// Post comment
			msg := fmt.Sprintf("@jules the git hub actions are failing. Failing github actions: {%s}.", strings.Join(failingChecks, ", "))
			if err := w.githubClient.PostComment(ctx, w.repoOwner, w.repoName, pr.Number, msg); err != nil {
				log.Error("Failed to post comment", "pr", pr.Number, "error", err)
			} else {
				log.Info("Posted failure comment", "pr", pr.Number, "failing_checks", failingChecks)
			}
		}
	}
}

func isJulesBot(login string) bool {
	// TODO: Make this configurable or more robust
	lower := strings.ToLower(login)
	return strings.Contains(lower, "jules")
}
