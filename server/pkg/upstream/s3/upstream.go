// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package s3

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"path"
	"strings"
	"unicode/utf8"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/mcpany/core/pkg/logging"
	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/upstream"
	"github.com/mcpany/core/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// S3ClientInterface defines the interface for S3 operations we use.
// This allows for mocking in tests.
type S3ClientInterface interface {
	ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
	HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error)
	AbortMultipartUpload(ctx context.Context, params *s3.AbortMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.AbortMultipartUploadOutput, error)
	CompleteMultipartUpload(ctx context.Context, params *s3.CompleteMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.CompleteMultipartUploadOutput, error)
	CreateMultipartUpload(ctx context.Context, params *s3.CreateMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.CreateMultipartUploadOutput, error)
	UploadPart(ctx context.Context, params *s3.UploadPartInput, optFns ...func(*s3.Options)) (*s3.UploadPartOutput, error)
}

// Ensure s3.Client satisfies S3ClientInterface
var _ S3ClientInterface = (*s3.Client)(nil)

type Upstream struct{}

func NewUpstream() upstream.Upstream {
	return &Upstream{}
}

func (u *Upstream) Shutdown(_ context.Context) error {
	return nil
}

func (u *Upstream) Register(
	ctx context.Context,
	serviceConfig *configv1.UpstreamServiceConfig,
	toolManager tool.ManagerInterface,
	promptManager prompt.ManagerInterface,
	resourceManager resource.ManagerInterface,
	isReload bool,
) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
	log := logging.GetLogger()

	// Calculate SHA256 for the ID
	h := sha256.New()
	h.Write([]byte(serviceConfig.GetName()))
	serviceConfig.SetId(hex.EncodeToString(h.Sum(nil)))

	// Sanitize the service name
	sanitizedName, err := util.SanitizeServiceName(serviceConfig.GetName())
	if err != nil {
		return "", nil, nil, err
	}
	serviceConfig.SetSanitizedName(sanitizedName)
	serviceID := sanitizedName

	s3Config := serviceConfig.GetS3Service()
	if s3Config == nil {
		return "", nil, nil, fmt.Errorf("s3 service config is nil")
	}

	client, err := u.createS3Client(ctx, s3Config)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to create s3 client: %w", err)
	}

	info := &tool.ServiceInfo{
		Name:   serviceConfig.GetName(),
		Config: serviceConfig,
	}
	toolManager.AddServiceInfo(serviceID, info)

	// Define built-in tools
	tools := u.defineTools(s3Config, client)

	discoveredTools := make([]*configv1.ToolDefinition, 0)

	for _, t := range tools {
		toolName := t.Name

		inputSchema, err := structpb.NewStruct(map[string]interface{}{
			"type":       "object",
			"properties": t.Input,
		})
		if err != nil {
			log.Error("Failed to create input schema", "tool", toolName, "error", err)
			continue
		}

		outputSchema, err := structpb.NewStruct(map[string]interface{}{
			"type":       "object",
			"properties": t.Output,
		})
		if err != nil {
			log.Error("Failed to create output schema", "tool", toolName, "error", err)
			continue
		}

		toolDef := configv1.ToolDefinition_builder{
			Name:      proto.String(toolName),
			ServiceId: proto.String(serviceID),
		}.Build()

		handler := t.Handler
		callable := &s3Callable{handler: handler}

		// Create a callable tool
		callableTool, err := tool.NewCallableTool(toolDef, serviceConfig, callable, inputSchema, outputSchema)
		if err != nil {
			log.Error("Failed to create callable tool", "tool", toolName, "error", err)
			continue
		}

		if err := toolManager.AddTool(callableTool); err != nil {
			log.Error("Failed to add tool", "tool", toolName, "error", err)
			continue
		}

		discoveredTools = append(discoveredTools, toolDef)
	}

	log.Info("Registered s3 service", "serviceID", serviceID, "tools", len(discoveredTools))
	return serviceID, discoveredTools, nil, nil
}

func (u *Upstream) createS3Client(ctx context.Context, s3Config *configv1.S3UpstreamService) (S3ClientInterface, error) {
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(s3Config.GetRegion()),
	}

	if s3Config.GetAccessKeyId() != "" && s3Config.GetSecretAccessKey() != "" {
		opts = append(opts, config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			s3Config.GetAccessKeyId(),
			s3Config.GetSecretAccessKey(),
			"",
		)))
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, err
	}

	clientOpts := []func(*s3.Options){}
	if s3Config.GetEndpoint() != "" {
		clientOpts = append(clientOpts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(s3Config.GetEndpoint())
			// For MinIO and others, we might need PathStyle
			o.UsePathStyle = true
		})
	}

	return s3.NewFromConfig(cfg, clientOpts...), nil
}

type toolHandler struct {
	Name        string
	Description string
	Input       map[string]interface{}
	Output      map[string]interface{}
	Handler     func(args map[string]interface{}) (map[string]interface{}, error)
}

func (u *Upstream) defineTools(s3Config *configv1.S3UpstreamService, client S3ClientInterface) []*toolHandler {
	return []*toolHandler{
		u.newListObjectsTool(s3Config, client),
		u.newGetObjectTool(s3Config, client),
		u.newPutObjectTool(s3Config, client),
		u.newDeleteObjectTool(s3Config, client),
		u.newGetObjectMetadataTool(s3Config, client),
	}
}

func (u *Upstream) newListObjectsTool(s3Config *configv1.S3UpstreamService, client S3ClientInterface) *toolHandler {
	return &toolHandler{
		Name:        "list_objects",
		Description: "List objects in the S3 bucket.",
		Input: map[string]interface{}{
			"prefix":   map[string]interface{}{"type": "string", "description": "The prefix to filter objects."},
			"max_keys": map[string]interface{}{"type": "integer", "description": "Maximum number of keys to return (default 100)."},
		},
		Output: map[string]interface{}{
			"objects": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"key":           map[string]interface{}{"type": "string"},
						"size":          map[string]interface{}{"type": "integer"},
						"last_modified": map[string]interface{}{"type": "string"},
					},
				},
			},
		},
		Handler: func(args map[string]interface{}) (map[string]interface{}, error) {
			prefix := ""
			if p, ok := args["prefix"].(string); ok {
				prefix = p
			}

			// Enforce config prefix
			if s3Config.GetPrefix() != "" {
				if !strings.HasPrefix(prefix, s3Config.GetPrefix()) {
					prefix = path.Join(s3Config.GetPrefix(), prefix)
				}
			}

			maxKeys := int32(100)
			if m, ok := args["max_keys"].(float64); ok { // JSON numbers are float64
				if m > 0 && m <= 2147483647 {
					maxKeys = int32(m)
				}
			}
			if m, ok := args["max_keys"].(int64); ok {
				if m > 0 && m <= 2147483647 {
					maxKeys = int32(m)
				}
			}
			if m, ok := args["max_keys"].(int); ok {
				if m > 0 && m <= 2147483647 {
					maxKeys = int32(m)
				}
			}

			input := &s3.ListObjectsV2Input{
				Bucket:  aws.String(s3Config.GetBucket()),
				Prefix:  aws.String(prefix),
				MaxKeys: aws.Int32(maxKeys),
			}

			resp, err := client.ListObjectsV2(context.Background(), input)
			if err != nil {
				return nil, err
			}

			objects := []interface{}{}
			for _, obj := range resp.Contents {
				objects = append(objects, map[string]interface{}{
					"key":           aws.ToString(obj.Key),
					"size":          aws.ToInt64(obj.Size),
					"last_modified": aws.ToTime(obj.LastModified).String(),
				})
			}

			return map[string]interface{}{"objects": objects}, nil
		},
	}
}

func (u *Upstream) newGetObjectTool(s3Config *configv1.S3UpstreamService, client S3ClientInterface) *toolHandler {
	return &toolHandler{
		Name:        "get_object",
		Description: "Get the content of an object.",
		Input: map[string]interface{}{
			"key": map[string]interface{}{"type": "string", "description": "The key of the object."},
		},
		Output: map[string]interface{}{
			"content":   map[string]interface{}{"type": "string"},
			"is_base64": map[string]interface{}{"type": "boolean"},
			"truncated": map[string]interface{}{"type": "boolean"},
		},
		Handler: func(args map[string]interface{}) (map[string]interface{}, error) {
			key, ok := args["key"].(string)
			if !ok {
				return nil, fmt.Errorf("key is required")
			}

			if err := u.validateKey(key, s3Config); err != nil {
				return nil, err
			}

			input := &s3.GetObjectInput{
				Bucket: aws.String(s3Config.GetBucket()),
				Key:    aws.String(key),
			}

			resp, err := client.GetObject(context.Background(), input)
			if err != nil {
				return nil, err
			}
			defer func() { _ = resp.Body.Close() }()

			// Read body
			// Limit to 4MB to prevent OOM
			const maxReadSize = 4 * 1024 * 1024
			limitReader := io.LimitReader(resp.Body, maxReadSize+1) // Read one extra byte to detect truncation
			bodyBytes, err := io.ReadAll(limitReader)
			if err != nil {
				return nil, err
			}

			truncated := false
			if len(bodyBytes) > maxReadSize {
				truncated = true
				bodyBytes = bodyBytes[:maxReadSize]
			}

			// Check if binary
			isBase64 := false
			content := ""

			if utf8.Valid(bodyBytes) {
				content = string(bodyBytes)
			} else {
				isBase64 = true
				content = base64.StdEncoding.EncodeToString(bodyBytes)
			}

			return map[string]interface{}{
				"content":   content,
				"is_base64": isBase64,
				"truncated": truncated,
			}, nil
		},
	}
}

func (u *Upstream) newPutObjectTool(s3Config *configv1.S3UpstreamService, client S3ClientInterface) *toolHandler {
	return &toolHandler{
		Name:        "put_object",
		Description: "Upload content to an object.",
		Input: map[string]interface{}{
			"key":       map[string]interface{}{"type": "string", "description": "The key of the object."},
			"content":   map[string]interface{}{"type": "string", "description": "The content to upload."},
			"is_base64": map[string]interface{}{"type": "boolean", "description": "Set to true if content is base64 encoded binary."},
		},
		Output: map[string]interface{}{
			"success": map[string]interface{}{"type": "boolean"},
		},
		Handler: func(args map[string]interface{}) (map[string]interface{}, error) {
			if s3Config.GetReadOnly() {
				return nil, fmt.Errorf("s3 service is read-only")
			}

			key, ok := args["key"].(string)
			if !ok {
				return nil, fmt.Errorf("key is required")
			}
			content, ok := args["content"].(string)
			if !ok {
				return nil, fmt.Errorf("content is required")
			}

			if err := u.validateKey(key, s3Config); err != nil {
				return nil, err
			}

			isBase64 := false
			if b, ok := args["is_base64"].(bool); ok {
				isBase64 = b
			}

			var body io.Reader
			if isBase64 {
				data, err := base64.StdEncoding.DecodeString(content)
				if err != nil {
					return nil, fmt.Errorf("invalid base64 content: %w", err)
				}
				body = strings.NewReader(string(data))
			} else {
				body = strings.NewReader(content)
			}

			// Use Uploader for better handling, but client.PutObject is fine for simple use
			// NewUploader accepts UploadAPIClient interface which our client satisfies
			uploader := manager.NewUploader(client)
			_, err := uploader.Upload(context.Background(), &s3.PutObjectInput{
				Bucket: aws.String(s3Config.GetBucket()),
				Key:    aws.String(key),
				Body:   body,
			})
			if err != nil {
				return nil, err
			}

			return map[string]interface{}{"success": true}, nil
		},
	}
}

func (u *Upstream) newDeleteObjectTool(s3Config *configv1.S3UpstreamService, client S3ClientInterface) *toolHandler {
	return &toolHandler{
		Name:        "delete_object",
		Description: "Delete an object.",
		Input: map[string]interface{}{
			"key": map[string]interface{}{"type": "string", "description": "The key of the object."},
		},
		Output: map[string]interface{}{
			"success": map[string]interface{}{"type": "boolean"},
		},
		Handler: func(args map[string]interface{}) (map[string]interface{}, error) {
			if s3Config.GetReadOnly() {
				return nil, fmt.Errorf("s3 service is read-only")
			}

			key, ok := args["key"].(string)
			if !ok {
				return nil, fmt.Errorf("key is required")
			}

			if err := u.validateKey(key, s3Config); err != nil {
				return nil, err
			}

			_, err := client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
				Bucket: aws.String(s3Config.GetBucket()),
				Key:    aws.String(key),
			})
			if err != nil {
				return nil, err
			}

			return map[string]interface{}{"success": true}, nil
		},
	}
}

func (u *Upstream) newGetObjectMetadataTool(s3Config *configv1.S3UpstreamService, client S3ClientInterface) *toolHandler {
	return &toolHandler{
		Name:        "get_object_metadata",
		Description: "Get metadata for an object.",
		Input: map[string]interface{}{
			"key": map[string]interface{}{"type": "string", "description": "The key of the object."},
		},
		Output: map[string]interface{}{
			"metadata":       map[string]interface{}{"type": "object"},
			"content_type":   map[string]interface{}{"type": "string"},
			"content_length": map[string]interface{}{"type": "integer"},
		},
		Handler: func(args map[string]interface{}) (map[string]interface{}, error) {
			key, ok := args["key"].(string)
			if !ok {
				return nil, fmt.Errorf("key is required")
			}

			if err := u.validateKey(key, s3Config); err != nil {
				return nil, err
			}

			resp, err := client.HeadObject(context.Background(), &s3.HeadObjectInput{
				Bucket: aws.String(s3Config.GetBucket()),
				Key:    aws.String(key),
			})
			if err != nil {
				return nil, err
			}

			metadata := make(map[string]interface{})
			for k, v := range resp.Metadata {
				metadata[k] = v
			}

			return map[string]interface{}{
				"metadata":       metadata,
				"content_type":   aws.ToString(resp.ContentType),
				"content_length": aws.ToInt64(resp.ContentLength),
			}, nil
		},
	}
}

func (u *Upstream) validateKey(key string, config *configv1.S3UpstreamService) error {
	if config.GetPrefix() != "" {
		if !strings.HasPrefix(key, config.GetPrefix()) {
			return fmt.Errorf("access denied: key %s does not match prefix %s", key, config.GetPrefix())
		}
	}
	return nil
}

type s3Callable struct {
	handler func(args map[string]interface{}) (map[string]interface{}, error)
}

func (c *s3Callable) Call(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	return c.handler(req.Arguments)
}
