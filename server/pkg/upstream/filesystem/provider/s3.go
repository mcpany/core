// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"fmt"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go/aws"             //nolint:staticcheck
	"github.com/aws/aws-sdk-go/aws/credentials" //nolint:staticcheck
	"github.com/aws/aws-sdk-go/aws/session"     //nolint:staticcheck
	s3 "github.com/fclairamb/afero-s3"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/spf13/afero"
)

// S3Provider represents the public S3Provider entity.
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
type S3Provider struct {
	fs afero.Fs
}

// NewS3Provider serves as a public interface for interacting with NewS3Provider.
//
// Summary: Constructs and returns an initialized provider ready for consumption.
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
func NewS3Provider(config *configv1.S3Fs) (*S3Provider, error) {
	if config == nil {
		return nil, fmt.Errorf("s3 config is nil")
	}

	awsConfig := aws.NewConfig()

	if config.GetRegion() != "" {
		awsConfig.WithRegion(config.GetRegion())
	}

	if config.GetAccessKeyId() != "" && config.GetSecretAccessKey() != "" {
		awsConfig.WithCredentials(credentials.NewStaticCredentials(
			config.GetAccessKeyId(),
			config.GetSecretAccessKey(),
			config.GetSessionToken(),
		))
	}

	if config.GetEndpoint() != "" {
		awsConfig.WithEndpoint(config.GetEndpoint())
		// Needed for MinIO and some S3 compatible services
		awsConfig.WithS3ForcePathStyle(true)
	}

	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	// Create S3 filesystem
	// Note: afero-s3 uses the bucket name as the root
	fs := s3.NewFs(config.GetBucket(), sess)

	return &S3Provider{fs: fs}, nil
}

// GetFs serves as a public interface for interacting with GetFs.
//
// Summary: Fetches and returns the underlying fs from the system state.
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
func (p *S3Provider) GetFs() afero.Fs {
	return p.fs
}

// ResolvePath serves as a public interface for interacting with ResolvePath.
//
// Summary: Resolve the path appropriately based on current system conditions.
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
func (p *S3Provider) ResolvePath(virtualPath string) (string, error) {
	// For S3, just clean the path. It's virtual relative to the bucket.
	// Join with "/" to ensure we resolve relative paths against a root, preventing ".." traversal
	// effectively sandboxing to the bucket root.
	// Use path package (not filepath) because S3 keys always use '/' separator.
	cleanPath := path.Clean("/" + virtualPath)

	// Strip the leading slash because S3 keys don't usually start with /
	cleanPath = strings.TrimPrefix(cleanPath, "/")

	if cleanPath == "" || cleanPath == "." {
		return "", fmt.Errorf("invalid path")
	}
	return cleanPath, nil
}

// Close serves as a public interface for interacting with Close.
//
// Summary: Close the  appropriately based on current system conditions.
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
func (p *S3Provider) Close() error {
	// S3 provider doesn't hold open connections that need explicit closing typically,
	// but satisfy the interface.
	return nil
}
