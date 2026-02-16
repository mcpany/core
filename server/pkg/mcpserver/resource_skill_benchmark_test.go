// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mcpany/core/server/pkg/skill"
)

func BenchmarkSkillResource_Read(b *testing.B) {
	// This benchmark measures the performance impact of caching the resolved path.
	// Setup temporary skill directory
	tmpDir := b.TempDir()
	skillDir := filepath.Join(tmpDir, "test-skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		b.Fatalf("Failed to create skill dir: %v", err)
	}

	// Create SKILL.md
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# Test Skill"), 0644); err != nil {
		b.Fatalf("Failed to create SKILL.md: %v", err)
	}

	// Create asset
	assetPath := "asset.txt"
	if err := os.WriteFile(filepath.Join(skillDir, assetPath), []byte("asset content"), 0644); err != nil {
		b.Fatalf("Failed to create asset: %v", err)
	}

	s := &skill.Skill{
		Frontmatter: skill.Frontmatter{
			Name: "test-skill",
		},
		Path: skillDir,
	}

	// Test case 1: Main SKILL.md
	b.Run("SKILL.md", func(b *testing.B) {
		res := NewSkillResource(s)
		ctx := context.Background()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := res.Read(ctx)
			if err != nil {
				b.Fatalf("Read failed: %v", err)
			}
		}
	})

	// Test case 2: Asset (involves EvalSymlinks logic)
	b.Run("Asset", func(b *testing.B) {
		res := NewSkillAssetResource(s, assetPath)
		ctx := context.Background()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := res.Read(ctx)
			if err != nil {
				b.Fatalf("Read failed: %v", err)
			}
		}
	})
}
