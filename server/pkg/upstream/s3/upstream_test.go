// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package s3

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockS3Client is a mock implementation of S3ClientInterface
type MockS3Client struct {
	mock.Mock
}

func (m *MockS3Client) ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*s3.ListObjectsV2Output), args.Error(1)
}

func (m *MockS3Client) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*s3.GetObjectOutput), args.Error(1)
}

func (m *MockS3Client) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*s3.PutObjectOutput), args.Error(1)
}

func (m *MockS3Client) DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*s3.DeleteObjectOutput), args.Error(1)
}

func (m *MockS3Client) HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*s3.HeadObjectOutput), args.Error(1)
}

func (m *MockS3Client) AbortMultipartUpload(ctx context.Context, params *s3.AbortMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.AbortMultipartUploadOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*s3.AbortMultipartUploadOutput), args.Error(1)
}

func (m *MockS3Client) CompleteMultipartUpload(ctx context.Context, params *s3.CompleteMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.CompleteMultipartUploadOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*s3.CompleteMultipartUploadOutput), args.Error(1)
}

func (m *MockS3Client) CreateMultipartUpload(ctx context.Context, params *s3.CreateMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.CreateMultipartUploadOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*s3.CreateMultipartUploadOutput), args.Error(1)
}

func (m *MockS3Client) UploadPart(ctx context.Context, params *s3.UploadPartInput, optFns ...func(*s3.Options)) (*s3.UploadPartOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*s3.UploadPartOutput), args.Error(1)
}

func TestS3Upstream_ListObjects(t *testing.T) {
	u := &Upstream{}
	s3Config := &configv1.S3UpstreamService{
		Bucket: aws.String("test-bucket"),
		Region: aws.String("us-east-1"),
	}

	mockClient := new(MockS3Client)
	mockClient.On("ListObjectsV2", mock.Anything, mock.MatchedBy(func(input *s3.ListObjectsV2Input) bool {
		return *input.Bucket == "test-bucket"
	})).Return(&s3.ListObjectsV2Output{
		Contents: []types.Object{
			{
				Key:          aws.String("test.txt"),
				Size:         aws.Int64(123),
				LastModified: aws.Time(time.Now()),
			},
		},
	}, nil)

	tools := u.defineTools(s3Config, mockClient)
	var listTool *toolHandler
	for _, t := range tools {
		if t.Name == "list_objects" {
			listTool = t
			break
		}
	}
	require.NotNil(t, listTool)

	res, err := listTool.Handler(map[string]interface{}{})
	require.NoError(t, err)

	objects := res["objects"].([]interface{})
	require.Len(t, objects, 1)
	obj := objects[0].(map[string]interface{})
	assert.Equal(t, "test.txt", obj["key"])
	assert.Equal(t, int64(123), obj["size"])

	mockClient.AssertExpectations(t)
}

func TestS3Upstream_GetObject(t *testing.T) {
	u := &Upstream{}
	s3Config := &configv1.S3UpstreamService{
		Bucket: aws.String("test-bucket"),
	}

	mockClient := new(MockS3Client)
	mockClient.On("GetObject", mock.Anything, mock.MatchedBy(func(input *s3.GetObjectInput) bool {
		return *input.Bucket == "test-bucket" && *input.Key == "test.txt"
	})).Return(&s3.GetObjectOutput{
		Body: io.NopCloser(strings.NewReader("hello world")),
	}, nil)

	tools := u.defineTools(s3Config, mockClient)
	var getTool *toolHandler
	for _, t := range tools {
		if t.Name == "get_object" {
			getTool = t
			break
		}
	}
	require.NotNil(t, getTool)

	res, err := getTool.Handler(map[string]interface{}{"key": "test.txt"})
	require.NoError(t, err)

	assert.Equal(t, "hello world", res["content"])
	assert.Equal(t, false, res["is_base64"])

	mockClient.AssertExpectations(t)
}

func TestS3Upstream_PutObject(t *testing.T) {
	u := &Upstream{}
	s3Config := &configv1.S3UpstreamService{
		Bucket: aws.String("test-bucket"),
	}

	mockClient := new(MockS3Client)
	// manager.Uploader calls PutObject for small files
	mockClient.On("PutObject", mock.Anything, mock.MatchedBy(func(input *s3.PutObjectInput) bool {
		return *input.Bucket == "test-bucket" && *input.Key == "test.txt"
	})).Return(&s3.PutObjectOutput{}, nil)

	tools := u.defineTools(s3Config, mockClient)
	var putTool *toolHandler
	for _, t := range tools {
		if t.Name == "put_object" {
			putTool = t
			break
		}
	}
	require.NotNil(t, putTool)

	res, err := putTool.Handler(map[string]interface{}{
		"key":     "test.txt",
		"content": "new content",
	})
	require.NoError(t, err)
	assert.Equal(t, true, res["success"])

	mockClient.AssertExpectations(t)
}

func TestS3Upstream_DeleteObject(t *testing.T) {
	u := &Upstream{}
	s3Config := &configv1.S3UpstreamService{
		Bucket: aws.String("test-bucket"),
	}

	mockClient := new(MockS3Client)
	mockClient.On("DeleteObject", mock.Anything, mock.MatchedBy(func(input *s3.DeleteObjectInput) bool {
		return *input.Bucket == "test-bucket" && *input.Key == "test.txt"
	})).Return(&s3.DeleteObjectOutput{}, nil)

	tools := u.defineTools(s3Config, mockClient)
	var delTool *toolHandler
	for _, t := range tools {
		if t.Name == "delete_object" {
			delTool = t
			break
		}
	}
	require.NotNil(t, delTool)

	res, err := delTool.Handler(map[string]interface{}{"key": "test.txt"})
	require.NoError(t, err)
	assert.Equal(t, true, res["success"])

	mockClient.AssertExpectations(t)
}

func TestS3Upstream_PrefixSecurity(t *testing.T) {
	u := &Upstream{}
	s3Config := &configv1.S3UpstreamService{
		Bucket: aws.String("test-bucket"),
		Prefix: aws.String("safe/"),
	}

	mockClient := new(MockS3Client)
	tools := u.defineTools(s3Config, mockClient)
	var getTool *toolHandler
	for _, t := range tools {
		if t.Name == "get_object" {
			getTool = t
			break
		}
	}

	// Should fail because key doesn't start with prefix
	_, err := getTool.Handler(map[string]interface{}{"key": "unsafe/file.txt"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "access denied")

	// Should succeed (mock call expected)
	mockClient.On("GetObject", mock.Anything, mock.MatchedBy(func(input *s3.GetObjectInput) bool {
		return *input.Key == "safe/file.txt"
	})).Return(&s3.GetObjectOutput{
		Body: io.NopCloser(strings.NewReader("ok")),
	}, nil)

	_, err = getTool.Handler(map[string]interface{}{"key": "safe/file.txt"})
	assert.NoError(t, err)
}

func TestS3Upstream_ReadOnly(t *testing.T) {
	u := &Upstream{}
	s3Config := &configv1.S3UpstreamService{
		Bucket:   aws.String("test-bucket"),
		ReadOnly: aws.Bool(true),
	}
	mockClient := new(MockS3Client)
	tools := u.defineTools(s3Config, mockClient)

	var putTool *toolHandler
	for _, t := range tools {
		if t.Name == "put_object" {
			putTool = t
			break
		}
	}

	_, err := putTool.Handler(map[string]interface{}{"key": "test.txt", "content": "foo"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "read-only")
}
