// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/util"
)

var (
	githubAPIURL        = "https://api.github.com"
	githubRawContentURL = "https://raw.githubusercontent.com"
)

const (
	githubURLRegexStr = `^https://github\.com/([^/]+)/([^/]+)/?(tree/|blob/)?([^/]+)?/?(.*)?`
)

var (
	githubURLRe = regexp.MustCompile(githubURLRegexStr)
)

// GitHub represents a client for interacting with the GitHub API to fetch
// configuration files or directories.
type GitHub struct {
	Owner         string
	Repo          string
	Path          string
	Ref           string
	URLType       string
	log           *slog.Logger
	apiURL        string
	rawContentURL string
	httpClient    *http.Client
}

// NewGitHub creates a new GitHub client by parsing a GitHub URL. It supports
// standard GitHub URLs for repositories, trees, and blobs.
//
// Parameters:
//   - ctx: The context for the client creation.
//   - rawURL: The GitHub URL to parse.
//
// Returns:
//   - A pointer to a new GitHub client.
//   - An error if the URL is invalid.
func NewGitHub(_ context.Context, rawURL string) (*GitHub, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	matches := githubURLRe.FindStringSubmatch(parsedURL.String())

	if len(matches) < 6 {
		return nil, fmt.Errorf("invalid GitHub URL format")
	}

	urlType := "tree"
	if strings.Contains(matches[3], "blob") {
		urlType = "blob"
	}

	ref := matches[4]
	if ref == "" {
		ref = "main"
	}

	return &GitHub{
		Owner:         matches[1],
		Repo:          matches[2],
		Ref:           ref,
		Path:          matches[5],
		URLType:       urlType,
		log:           logging.GetLogger().With("component", "GitHub"),
		apiURL:        githubAPIURL,
		rawContentURL: githubRawContentURL,
		httpClient: &http.Client{
			Transport: &http.Transport{
				DialContext: util.SafeDialContext,
			},
		},
	}, nil
}

func isGitHubURL(rawURL string) bool {
	return githubURLRe.MatchString(rawURL)
}

// ToRawContentURL constructs the raw content URL for the configured GitHub path.
//
// Returns:
//
//	string: The raw content URL string.
func (g *GitHub) ToRawContentURL() string {
	return fmt.Sprintf("%s/%s/%s/%s/%s", g.rawContentURL, g.Owner, g.Repo, g.Ref, g.Path)
}

// Content represents a file or directory in a GitHub repository.
type Content struct {
	// Name is the name of the file or directory.
	Name string `json:"name"`
	// Type is the type of content (e.g., "file", "dir").
	Type string `json:"type"`
	// HTMLURL is the URL to view the content on GitHub.
	HTMLURL string `json:"html_url"`
	// DownloadURL is the URL to download the content (only for files).
	DownloadURL string `json:"download_url"`
}

// List fetches the contents of the configured GitHub path. It handles authentication
// if provided and returns a list of Content objects.
//
// Parameters:
//   - ctx: The context for the request.
//   - auth: Optional authentication configuration for accessing private repos.
//
// Returns:
//   - A slice of Content objects.
//   - An error if the fetch fails.
//
// List fetches the contents of the configured GitHub path. It handles authentication
// if provided and returns a list of Content objects.
//
// Parameters:
//   - ctx: The context for the request.
//   - auth: Optional authentication configuration for accessing private repos.
//
// Returns:
//   - A slice of Content objects.
//   - An error if the fetch fails.
func (g *GitHub) List(ctx context.Context, auth *configv1.Authentication) ([]Content, error) {
	apiURL := fmt.Sprintf("%s/repos/%s/%s/contents/%s", g.apiURL, g.Owner, g.Repo, g.Path)
	if g.Ref != "" {
		apiURL = fmt.Sprintf("%s?ref=%s", apiURL, g.Ref)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	if err := g.applyAuthentication(req, auth); err != nil {
		return nil, fmt.Errorf("failed to apply authentication for github api: %w", err)
	}

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch contents from github api: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch contents from github api: status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read github api response body: %w", err)
	}

	var contents []Content
	if err := json.Unmarshal(body, &contents); err != nil {
		var content Content
		if err := json.Unmarshal(body, &content); err != nil {
			return nil, fmt.Errorf("failed to decode github api response: %w", err)
		}
		contents = append(contents, content)
	}

	return contents, nil
}

func (g *GitHub) applyAuthentication(req *http.Request, auth *configv1.Authentication) error {
	if auth == nil {
		return nil
	}

	if apiKey := auth.GetApiKey(); apiKey != nil {
		apiKeyValue, err := util.ResolveSecret(req.Context(), apiKey.GetValue())
		if err != nil {
			return err
		}
		req.Header.Set(apiKey.GetParamName(), apiKeyValue)
	} else if bearerToken := auth.GetBearerToken(); bearerToken != nil {
		tokenValue, err := util.ResolveSecret(req.Context(), bearerToken.GetToken())
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+tokenValue)
	} else if basicAuth := auth.GetBasicAuth(); basicAuth != nil {
		passwordValue, err := util.ResolveSecret(req.Context(), basicAuth.GetPassword())
		if err != nil {
			return err
		}
		req.SetBasicAuth(basicAuth.GetUsername(), passwordValue)
	}

	return nil
}
