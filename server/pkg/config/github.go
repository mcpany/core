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

// GitHub represents the public GitHub entity.
//
// Summary: Defines the structured data model representing a hub.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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

// NewGitHub serves as a public interface for interacting with NewGitHub.
//
// Summary: Constructs and returns an initialized git hub ready for consumption.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// ToRawContentURL serves as a public interface for interacting with ToRawContentURL.
//
// Summary: To the raw content url appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (g *GitHub) ToRawContentURL() string {
	return fmt.Sprintf("%s/%s/%s/%s/%s", g.rawContentURL, g.Owner, g.Repo, g.Ref, g.Path)
}

// Content represents the public Content entity.
//
// Summary: Defines the structured data model representing a .
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type Content struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	HTMLURL     string `json:"html_url"`
	DownloadURL string `json:"download_url"`
}

// List serves as a public interface for interacting with List.
//
// Summary: List the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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
