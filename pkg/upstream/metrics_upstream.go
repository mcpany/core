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

package upstream

import (
	"context"
	"time"

	"github.com/mcpany/core/pkg/metrics"
	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
)

// MetricsUpstream is a decorator that adds metrics to an upstream service.
type MetricsUpstream struct {
	upstream Upstream
	name     string
}

// NewMetricsUpstream creates a new MetricsUpstream.
func NewMetricsUpstream(upstream Upstream, name string) Upstream {
	return &MetricsUpstream{
		upstream: upstream,
		name:     name,
	}
}

// Register wraps the Register method of the underlying upstream service and
// adds metrics for latency and success/failure.
func (u *MetricsUpstream) Register(
	ctx context.Context,
	serviceConfig *configv1.UpstreamServiceConfig,
	toolManager tool.ToolManagerInterface,
	promptManager prompt.PromptManagerInterface,
	resourceManager resource.ResourceManagerInterface,
	isReload bool,
) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
	metrics.IncrCounter([]string{"upstream", u.name, "register", "requests"}, 1)
	defer metrics.MeasureSince([]string{"upstream", u.name, "register", "latency"}, time.Now())

	key, tools, resources, err := u.upstream.Register(ctx, serviceConfig, toolManager, promptManager, resourceManager, isReload)
	if err != nil {
		metrics.IncrCounter([]string{"upstream", u.name, "register", "errors"}, 1)
	}
	return key, tools, resources, err
}
