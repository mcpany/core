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

// S3Provider provides access to files in an S3 bucket.
type S3Provider struct {
	fs afero.Fs
}

// NewS3Provider creates a new S3Provider from the given configuration.
//
// config holds the configuration settings.
//
// Returns the result.
// Returns an error if the operation fails.
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

// GetFs returns the underlying filesystem.
//
// Returns the result.
func (p *S3Provider) GetFs() afero.Fs {
	return p.fs
}

// ResolvePath resolves the virtual path to a real path in the bucket.
//
// virtualPath is the virtualPath.
//
// Returns the result.
// Returns an error if the operation fails.
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

// Close closes the provider.
//
// Returns an error if the operation fails.
func (p *S3Provider) Close() error {
	// S3 provider doesn't hold open connections that need explicit closing typically,
	// but satisfy the interface.
	return nil
}
