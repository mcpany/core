// Copyright 2025 Author(s) of MCP Any
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
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"regexp"
	"log/slog"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/pkg/util"
	"github.com/mcpany/core/pkg/logging"
)

var (
	githubAPIURL        = "https://api.github.com"
	githubRawContentURL = "https://raw.githubusercontent.com"
)

const (
	githubURLRegex      = `^https://github\.com/([^/]+)/([^/]+)/?(tree/|blob/)?([^/]+)?/?(.*)?`
)

type GitHub struct {
	Owner       string
	Repo        string
	Path        string
	Ref         string
	URLType     string
	log         *slog.Logger
	apiURL      string
	rawContentURL string
}

func NewGitHub(ctx context.Context, rawURL string) (*GitHub, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	re := regexp.MustCompile(githubURLRegex)
	matches := re.FindStringSubmatch(parsedURL.String())

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
		Owner:       matches[1],
		Repo:        matches[2],
		Ref:         ref,
		Path:        matches[5],
		URLType:     urlType,
		log:         logging.GetLogger().With("component", "GitHub"),
		apiURL:      githubAPIURL,
		rawContentURL: githubRawContentURL,
	}, nil
}

func isGitHubURL(rawURL string) bool {
	re := regexp.MustCompile(githubURLRegex)
	return re.MatchString(rawURL)
}

func (g *GitHub) ToRawContentURL() string {
	return fmt.Sprintf("%s/%s/%s/%s/%s", g.rawContentURL, g.Owner, g.Repo, g.Ref, g.Path)
}

type Content struct {
	Type        string `json:"type"`
	HTMLURL     string `json:"html_url"`
	DownloadURL string `json:"download_url"`
}

func (g *GitHub) List(ctx context.Context, auth *configv1.UpstreamAuthentication) ([]Content, error) {
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

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch contents from github api: %w", err)
	}
	defer resp.Body.Close()

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

func (g *GitHub) applyAuthentication(req *http.Request, auth *configv1.UpstreamAuthentication) error {
	if auth == nil {
		return nil
	}

	if apiKey := auth.GetApiKey(); apiKey != nil {
		apiKeyValue, err := util.ResolveSecret(apiKey.GetApiKey())
		if err != nil {
			return err
		}
		req.Header.Set(apiKey.GetHeaderName(), apiKeyValue)
	} else if bearerToken := auth.GetBearerToken(); bearerToken != nil {
		tokenValue, err := util.ResolveSecret(bearerToken.GetToken())
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+tokenValue)
	} else if basicAuth := auth.GetBasicAuth(); basicAuth != nil {
		passwordValue, err := util.ResolveSecret(basicAuth.GetPassword())
		if err != nil {
			return err
		}
		req.SetBasicAuth(basicAuth.GetUsername(), passwordValue)
	}

	return nil
}
