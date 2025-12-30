// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mcpany/core/pkg/util"
)

// GitHubClient is a client for interacting with the GitHub API.
type GitHubClient struct {
	BaseURL    string
	HTTPClient *http.Client
	Token      string
}

// NewGitHubClient creates a new GitHubClient.
func NewGitHubClient(token string) *GitHubClient {
	return &GitHubClient{
		BaseURL: "https://api.github.com",
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				DialContext: util.SafeDialContext,
			},
			Timeout: 10 * time.Second,
		},
		Token: token,
	}
}

// PullRequest represents a GitHub pull request.
type PullRequest struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	User   User   `json:"user"`
	Head   Head   `json:"head"`
}

// User represents a GitHub user.
type User struct {
	Login string `json:"login"`
}

// Head represents the head of a pull request.
type Head struct {
	Sha string `json:"sha"`
}

// CheckRun represents a GitHub check run.
type CheckRun struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
}

// CheckRunsResponse represents the response from the check runs API.
type CheckRunsResponse struct {
	TotalCount int        `json:"total_count"`
	CheckRuns  []CheckRun `json:"check_runs"`
}

// IssueComment represents a comment on an issue or PR.
type IssueComment struct {
	ID   int64  `json:"id"`
	User User   `json:"user"`
	Body string `json:"body"`
}

// ListOpenPRs lists open pull requests for a repository.
func (c *GitHubClient) ListOpenPRs(ctx context.Context, owner, repo string) ([]PullRequest, error) {
	var allPRs []PullRequest
	page := 1
	for {
		url := fmt.Sprintf("%s/repos/%s/%s/pulls?state=open&per_page=100&page=%d", c.BaseURL, owner, repo, page)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}

		c.setHeaders(req)

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to list PRs: %s", resp.Status)
		}

		var prs []PullRequest
		if err := json.NewDecoder(resp.Body).Decode(&prs); err != nil {
			return nil, err
		}

		if len(prs) == 0 {
			break
		}

		allPRs = append(allPRs, prs...)
		page++
	}

	return allPRs, nil
}

// GetCheckRuns lists check runs for a specific commit ref.
func (c *GitHubClient) GetCheckRuns(ctx context.Context, owner, repo, ref string) (*CheckRunsResponse, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/commits/%s/check-runs", c.BaseURL, owner, repo, ref)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	c.setHeaders(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get check runs: %s", resp.Status)
	}

	var runs CheckRunsResponse
	if err := json.NewDecoder(resp.Body).Decode(&runs); err != nil {
		return nil, err
	}

	return &runs, nil
}

// PostComment posts a comment on an issue or PR.
func (c *GitHubClient) PostComment(ctx context.Context, owner, repo string, number int, body string) error {
	url := fmt.Sprintf("%s/repos/%s/%s/issues/%d/comments", c.BaseURL, owner, repo, number)
	payload := map[string]string{"body": body}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}

	c.setHeaders(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to post comment: %s", resp.Status)
	}

	return nil
}

// ListComments lists comments on an issue or PR.
func (c *GitHubClient) ListComments(ctx context.Context, owner, repo string, number int) ([]IssueComment, error) {
	var allComments []IssueComment
	page := 1
	for {
		url := fmt.Sprintf("%s/repos/%s/%s/issues/%d/comments?per_page=100&page=%d", c.BaseURL, owner, repo, number, page)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}

		c.setHeaders(req)

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to list comments: %s", resp.Status)
		}

		var comments []IssueComment
		if err := json.NewDecoder(resp.Body).Decode(&comments); err != nil {
			return nil, err
		}

		if len(comments) == 0 {
			break
		}

		allComments = append(allComments, comments...)
		page++
	}

	return allComments, nil
}

// GetUser gets the authenticated user.
func (c *GitHubClient) GetUser(ctx context.Context) (*User, error) {
	url := fmt.Sprintf("%s/user", c.BaseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user: %s", resp.Status)
	}
	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (c *GitHubClient) setHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
}
