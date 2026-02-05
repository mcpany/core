// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package protobufparser provides a parser for Protocol Buffers.
package protobufparser

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/bufbuild/protocompile"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	reflectpb "google.golang.org/grpc/reflection/grpc_reflection_v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"

	configv1 "github.com/mcpany/core/proto/config/v1"
	mcpopt "github.com/mcpany/core/proto/mcp_options/v1"
)

// ParsedMcpAnnotations holds the structured data extracted from MCP
// (Model Context Protocol) annotations within a set of protobuf files.
type ParsedMcpAnnotations struct {
	Tools     []McpTool
	Prompts   []McpPrompt
	Resources []McpResource
}

// McpTool represents the information extracted from a gRPC method that has been
// annotated as an MCP tool.
type McpTool struct {
	Name            string
	Description     string
	ServiceName     string
	MethodName      string
	FullMethodName  string // e.g., /package.ServiceName/MethodName
	RequestType     string // Fully qualified name
	ResponseType    string // Fully qualified name
	RequestFields   []McpField
	ResponseFields  []McpField
	ReadOnlyHint    bool
	DestructiveHint bool
	IdempotentHint  bool
	OpenWorldHint   bool
}

// McpField represents a field within a protobuf message, including its name,
// description, type, and whether it is repeated.
type McpField struct {
	Name        string
	Description string
	Type        string
	IsRepeated  bool
}

// GetName returns the name of the McpField.
//
// Returns the result.
func (f *McpField) GetName() string {
	return f.Name
}

// GetDescription returns the description of the McpField.
//
// Returns the result.
func (f *McpField) GetDescription() string {
	return f.Description
}

// GetType returns the type of the McpField.
//
// Returns the result.
func (f *McpField) GetType() string {
	return f.Type
}

// GetIsRepeated returns true if the McpField is a repeated field.
//
// Returns the result.
func (f *McpField) GetIsRepeated() bool {
	return f.IsRepeated
}

// ParseProtoFromDefs parses a set of protobuf definitions from a slice of
// ProtoDefinition and a ProtoCollection. It writes the proto files to a
// temporary directory, invokes protoc to generate a FileDescriptorSet, and
// then returns the parsed FileDescriptorSet.
func ParseProtoFromDefs(
	ctx context.Context,
	protoDefinitions []*configv1.ProtoDefinition,
	protoCollections []*configv1.ProtoCollection,
) (*descriptorpb.FileDescriptorSet, error) {
	// Create a temporary directory to store the proto files
	tempDir, err := os.MkdirTemp("", "proto-defs-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	var protoFiles []string

	// Process ProtoCollection first
	for _, protoCollection := range protoCollections {
		if protoCollection != nil {
			collectionFiles, err := processProtoCollection(protoCollection, tempDir)
			if err != nil {
				return nil, fmt.Errorf("failed to process proto collection: %w", err)
			}
			protoFiles = append(protoFiles, collectionFiles...)
		}
	}

	// Process ProtoDefinitions
	for _, def := range protoDefinitions {
		switch def.WhichProtoRef() {
		case configv1.ProtoDefinition_ProtoFile_case:
			protoFile := def.GetProtoFile()
			filePath, err := writeProtoFile(protoFile, tempDir)
			if err != nil {
				return nil, fmt.Errorf("failed to write proto file: %w", err)
			}
			protoFiles = append(protoFiles, filePath)
		case configv1.ProtoDefinition_ProtoDescriptor_case:
			// For now, we assume proto descriptors are handled by protoc
			// by being included in the import paths.
		}
	}

	if len(protoFiles) == 0 {
		return nil, fmt.Errorf("no proto files found to parse")
	}

	// Use protocompile to generate the FileDescriptorSet
	importPaths := []string{tempDir}
	// Add project root and proto directories to import paths if they exist.
	// We try to find the root by looking for go.mod starting from the current directory.
	if cwd, err := os.Getwd(); err == nil {
		dir := cwd
		for {
			// Look for common project root markers. We prefer the one containing our build environment.
			if _, err := os.Stat(filepath.Join(dir, "build/env/bin/include")); err == nil {
				importPaths = append(importPaths, dir)
				importPaths = append(importPaths, filepath.Join(dir, "proto"))
				importPaths = append(importPaths, filepath.Join(dir, "build/env/bin/include"))
				break
			}
			// Fallback to go.mod but only if it also has a proto dir (to avoid being trapped in sub-modules like 'server')
			if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
				if _, err := os.Stat(filepath.Join(dir, "proto")); err == nil {
					importPaths = append(importPaths, dir)
					importPaths = append(importPaths, filepath.Join(dir, "proto"))
					break
				}
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}

	compiler := protocompile.Compiler{
		Resolver: &protocompile.SourceResolver{
			ImportPaths: importPaths,
		},
	}

	// The file names passed to Compile must be relative to the import paths.
	// Since tempDir is our import path, we need to get the base names of the files.
	relativeFiles := make([]string, len(protoFiles))
	for i, p := range protoFiles {
		rel, err := filepath.Rel(tempDir, p)
		if err != nil {
			return nil, fmt.Errorf("failed to get relative path for %s: %w", p, err)
		}
		relativeFiles[i] = rel
	}

	fileDescriptors, err := compiler.Compile(ctx, relativeFiles...)
	if err != nil {
		return nil, fmt.Errorf("failed to parse proto files: %w", err)
	}

	fds := &descriptorpb.FileDescriptorSet{}
	seen := make(map[string]bool)
	var collect func(fd protoreflect.FileDescriptor)
	collect = func(fd protoreflect.FileDescriptor) {
		if seen[fd.Path()] {
			return
		}
		seen[fd.Path()] = true
		fds.File = append(fds.File, protodesc.ToFileDescriptorProto(fd))
		imports := fd.Imports()
		for i := 0; i < imports.Len(); i++ {
			if imp := imports.Get(i).FileDescriptor; imp != nil {
				collect(imp)
			}
		}
	}

	for _, fd := range fileDescriptors {
		collect(fd)
	}

	return fds, nil
}

func processProtoCollection(
	collection *configv1.ProtoCollection,
	tempDir string,
) ([]string, error) {
	var protoFiles []string
	regex, err := regexp.Compile(collection.GetPathMatchRegex())
	if err != nil {
		return nil, fmt.Errorf("invalid path_match_regex: %w", err)
	}

	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && regex.MatchString(path) {
			relPath, err := filepath.Rel(collection.GetRootPath(), path)
			if err != nil {
				return fmt.Errorf("failed to get relative path for %s: %w", path, err)
			}

			destPath := filepath.Join(tempDir, relPath)
			if err := os.MkdirAll(filepath.Dir(destPath), 0o750); err != nil {
				return fmt.Errorf("failed to create directory for proto file: %w", err)
			}

			content, err := os.ReadFile(path) //nolint:gosec // Path is controlled by configuration/regex
			if err != nil {
				return fmt.Errorf("failed to read proto file %s: %w", path, err)
			}
			if err := os.WriteFile(destPath, content, 0o600); err != nil {
				return fmt.Errorf("failed to write proto file to temp dir: %w", err)
			}
			protoFiles = append(protoFiles, destPath)
		}
		return nil
	}

	if collection.GetIsRecursive() {
		if err := filepath.Walk(collection.GetRootPath(), walkFunc); err != nil {
			return nil, fmt.Errorf("failed to walk root path: %w", err)
		}
	} else {
		entries, err := os.ReadDir(collection.GetRootPath())
		if err != nil {
			return nil, fmt.Errorf("failed to read root path: %w", err)
		}
		for _, entry := range entries {
			fullPath := filepath.Join(collection.GetRootPath(), entry.Name())
			info, err := entry.Info()
			if err != nil {
				return nil, fmt.Errorf("failed to get file info: %w", err)
			}
			if err := walkFunc(fullPath, info, nil); err != nil {
				return nil, err
			}
		}
	}

	return protoFiles, nil
}

func writeProtoFile(protoFile *configv1.ProtoFile, tempDir string) (string, error) {
	filePath := filepath.Join(tempDir, protoFile.GetFileName())
	if err := os.MkdirAll(filepath.Dir(filePath), 0o750); err != nil {
		return "", fmt.Errorf("failed to create directory for proto file: %w", err)
	}

	var content []byte
	var err error
	switch protoFile.WhichFileRef() {
	case configv1.ProtoFile_FileContent_case:
		content = []byte(protoFile.GetFileContent())
	case configv1.ProtoFile_FilePath_case:
		filePathRef := protoFile.GetFilePath()

		content, err = os.ReadFile(filePathRef) //nolint:gosec // Path is from config
		if err != nil {
			return "", fmt.Errorf("failed to read proto file from path %s: %w", filePathRef, err)
		}
	default:
		return "", fmt.Errorf("proto file definition for '%s' has neither content nor a path", protoFile.GetFileName())
	}

	if err := os.WriteFile(filePath, content, 0o600); err != nil {
		return "", fmt.Errorf("failed to write proto file: %w", err)
	}
	return filePath, nil
}

// McpPrompt represents the information extracted from a gRPC method that has
// been annotated as an MCP prompt.
type McpPrompt struct {
	Name           string
	Description    string
	Template       string
	ServiceName    string
	MethodName     string
	FullMethodName string
	RequestType    string
	ResponseType   string
}

// McpResource represents a protobuf message that has been annotated as an MCP
// resource.
type McpResource struct {
	Name        string
	Description string
	MessageType string
}

// ParseProtoByReflection connects to a gRPC service that has server reflection
// enabled, discovers its entire set of protobuf definitions, including all
// dependencies, and returns them as a complete FileDescriptorSet.
//
// ctx is the context for the reflection process, including timeouts.
// target is the address of the gRPC service to connect to.
func ParseProtoByReflection(ctx context.Context, target string) (*descriptorpb.FileDescriptorSet, error) {
	// Create a context with a timeout for the entire reflection process
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// 1. Connect to the gRPC service
	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC service at %s: %w", target, err)
	}
	defer func() { _ = conn.Close() }()

	return parseProtoWithExistingConnection(ctx, conn)
}

// ParseProtoWithExistingConnection performs reflection on an existing gRPC
// connection. It is a lower-level function that allows for more control over
// the connection lifecycle, which is useful in testing or scenarios where a
// connection is already established.
func parseProtoWithExistingConnection(ctx context.Context, conn grpc.ClientConnInterface) (*descriptorpb.FileDescriptorSet, error) {
	// 2. Create a reflection client
	reflectionClient := reflectpb.NewServerReflectionClient(conn)
	// The reflection stream should be valid for the lifetime of the queries.
	stream, err := reflectionClient.ServerReflectionInfo(ctx, grpc.WaitForReady(true))
	if err != nil {
		return nil, fmt.Errorf("failed to create reflection stream: %w", err)
	}
	defer func() {
		if err := stream.CloseSend(); err != nil {
			fmt.Printf("Failed to close send stream: %v\n", err)
		}
	}()

	// 3. List all services to get starting points for discovery
	serviceNames, err := listServices(stream)
	if err != nil {
		return nil, fmt.Errorf("failed to list services via reflection: %w", err)
	}
	if len(serviceNames) == 0 {
		return nil, fmt.Errorf("no services found via reflection")
	}

	// 4. Recursively discover all file descriptors
	allFdps := make(map[string]*descriptorpb.FileDescriptorProto)
	fileQueue := []string{} // A queue of filenames to fetch

	// Seed the queue by finding the files that contain the advertised services
	for _, serviceName := range serviceNames {
		if strings.HasPrefix(serviceName, "grpc.reflection.v1alpha") || serviceName == "grpc.health.v1.Health" {
			continue
		}
		fdp, err := getFileDescriptorForSymbol(stream, serviceName)
		if err != nil {
			return nil, fmt.Errorf("failed to get file descriptor for service '%s': %w", serviceName, err)
		}
		if _, ok := allFdps[fdp.GetName()]; !ok {
			allFdps[fdp.GetName()] = fdp
			fileQueue = append(fileQueue, fdp.GetName())
		}
	}

	// Process the queue to fetch all dependencies
	for i := 0; i < len(fileQueue); i++ { // Use index since queue grows
		filename := fileQueue[i]
		fdp := allFdps[filename]

		for _, depFilename := range fdp.GetDependency() {
			if _, ok := allFdps[depFilename]; !ok {
				depFdp, err := getFileDescriptorByFilename(stream, depFilename)
				if err != nil {
					// It's possible a dependency is a well-known type that the server
					// doesn't explicitly provide. We can ignore errors here and let
					// the final parsing step handle missing dependencies if it's truly an issue.
					fmt.Printf("Could not fetch dependency '%s' for '%s': %v. It may be a well-known type.\n", depFilename, filename, err)
					continue
				}
				if _, ok := allFdps[depFdp.GetName()]; !ok {
					allFdps[depFdp.GetName()] = depFdp
					fileQueue = append(fileQueue, depFdp.GetName())
				}
			}
		}
	}

	// 5. Assemble the final FileDescriptorSet
	fds := &descriptorpb.FileDescriptorSet{}
	for _, fdp := range allFdps {
		fds.File = append(fds.File, fdp)
	}

	return fds, nil
}

// listServices sends a ListServices request over a reflection stream and
// returns the list of discovered service names.
func listServices(stream reflectpb.ServerReflection_ServerReflectionInfoClient) ([]string, error) {
	err := stream.Send(&reflectpb.ServerReflectionRequest{
		MessageRequest: &reflectpb.ServerReflectionRequest_ListServices{},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send ListServices request: %w", err)
	}

	resp, err := stream.Recv()
	if err != nil {
		return nil, fmt.Errorf("failed to receive ListServices response: %w", err)
	}

	listServicesResp := resp.GetListServicesResponse()
	if listServicesResp == nil {
		return nil, fmt.Errorf("invalid response type for ListServices: %T", resp.MessageResponse)
	}

	serviceNames := make([]string, 0, len(listServicesResp.Service))
	for _, s := range listServicesResp.Service {
		serviceNames = append(serviceNames, s.Name)
	}
	return serviceNames, nil
}

// getFileDescriptorForSymbol queries the reflection service for the
// FileDescriptorProto that defines a given symbol (e.g., a service name,
// message name).
func getFileDescriptorForSymbol(stream reflectpb.ServerReflection_ServerReflectionInfoClient, symbolName string) (*descriptorpb.FileDescriptorProto, error) {
	err := stream.Send(&reflectpb.ServerReflectionRequest{
		MessageRequest: &reflectpb.ServerReflectionRequest_FileContainingSymbol{
			FileContainingSymbol: symbolName,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send FileContainingSymbol request for %s: %w", symbolName, err)
	}

	resp, err := stream.Recv()
	if err != nil {
		return nil, fmt.Errorf("failed to receive FileContainingSymbol response for %s: %w", symbolName, err)
	}

	fileDescResp := resp.GetFileDescriptorResponse()
	if fileDescResp == nil || len(fileDescResp.GetFileDescriptorProto()) == 0 {
		return nil, fmt.Errorf("invalid or empty response for FileContainingSymbol for symbol %s: %T", symbolName, resp.MessageResponse)
	}

	// The response contains a slice of bytes, which is a serialized FileDescriptorProto.
	fdp := &descriptorpb.FileDescriptorProto{}
	if err := proto.Unmarshal(fileDescResp.FileDescriptorProto[0], fdp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal FileDescriptorProto for %s: %w", symbolName, err)
	}

	return fdp, nil
}

// getFileDescriptorByFilename queries the reflection service for a
// FileDescriptorProto by its filename (e.g., "path/to/my_service.proto").
func getFileDescriptorByFilename(stream reflectpb.ServerReflection_ServerReflectionInfoClient, filename string) (*descriptorpb.FileDescriptorProto, error) {
	err := stream.Send(&reflectpb.ServerReflectionRequest{
		MessageRequest: &reflectpb.ServerReflectionRequest_FileByFilename{
			FileByFilename: filename,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send FileByFilename request for %s: %w", filename, err)
	}

	resp, err := stream.Recv()
	if err != nil {
		return nil, fmt.Errorf("failed to receive FileByFilename response for %s: %w", filename, err)
	}

	fileDescResp := resp.GetFileDescriptorResponse()
	if fileDescResp == nil || len(fileDescResp.GetFileDescriptorProto()) == 0 {
		return nil, fmt.Errorf("invalid or empty response for FileByFilename for %s: %T", filename, resp.MessageResponse)
	}

	fdp := &descriptorpb.FileDescriptorProto{}
	if err := proto.Unmarshal(fileDescResp.FileDescriptorProto[0], fdp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal FileDescriptorProto for %s: %w", filename, err)
	}

	return fdp, nil
}

// ExtractMcpDefinitions iterates through a FileDescriptorSet, parsing any MCP
// (Model Context Protocol) options found in service methods and messages. It
// extracts definitions for tools, prompts, and resources.
//
// fds is the FileDescriptorSet to be parsed.
// It returns a ParsedMcpAnnotations struct containing the extracted information
// or an error if the parsing fails.
func ExtractMcpDefinitions(fds *descriptorpb.FileDescriptorSet) (*ParsedMcpAnnotations, error) {
	if fds == nil {
		return nil, fmt.Errorf("FileDescriptorSet is nil")
	}

	files, err := protodesc.NewFiles(fds)
	if err != nil {
		return nil, fmt.Errorf("failed to create protodesc files: %w", err)
	}

	results := &ParsedMcpAnnotations{}
	var tools []McpTool
	var prompts []McpPrompt
	var resources []McpResource

	files.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		messages := fd.Messages()
		for i := 0; i < messages.Len(); i++ {
			msgDesc := messages.Get(i)
			opts := msgDesc.Options()
			if proto.HasExtension(opts, mcpopt.E_ResourceName) {
				resourceName := proto.GetExtension(opts, mcpopt.E_ResourceName).(string)
				resourceDesc := ""
				if proto.HasExtension(opts, mcpopt.E_ResourceDescription) {
					resourceDesc = proto.GetExtension(opts, mcpopt.E_ResourceDescription).(string)
				}
				resources = append(resources, McpResource{
					Name:        resourceName,
					Description: resourceDesc,
					MessageType: string(msgDesc.FullName()),
				})
			}
		}

		services := fd.Services()
		for i := 0; i < services.Len(); i++ {
			serviceDesc := services.Get(i)

			methods := serviceDesc.Methods()
			for j := 0; j < methods.Len(); j++ {
				methodDesc := methods.Get(j)
				methodOpts := methodDesc.Options()
				var toolName, toolDesc string
				var readOnlyHint, destructiveHint, idempotentHint, openWorldHint bool

				if methodOpts != nil {
					if proto.HasExtension(methodOpts, mcpopt.E_ToolName) {
						toolName = proto.GetExtension(methodOpts, mcpopt.E_ToolName).(string)
					}
					if proto.HasExtension(methodOpts, mcpopt.E_ToolDescription) {
						toolDesc = proto.GetExtension(methodOpts, mcpopt.E_ToolDescription).(string)
					}
					readOnlyHint = proto.GetExtension(methodOpts, mcpopt.E_McpToolReadonlyHint).(bool)
					destructiveHint = proto.GetExtension(methodOpts, mcpopt.E_McpToolDestructiveHint).(bool)
					idempotentHint = proto.GetExtension(methodOpts, mcpopt.E_McpToolIdempotentHint).(bool)
					openWorldHint = proto.GetExtension(methodOpts, mcpopt.E_McpToolOpenworldHint).(bool)
				}

				if toolName == "" {
					toolName = string(methodDesc.Name())
				}

				tools = append(tools, McpTool{
					Name:            toolName,
					Description:     toolDesc,
					ServiceName:     string(serviceDesc.Name()),
					MethodName:      string(methodDesc.Name()),
					FullMethodName:  fmt.Sprintf("/%s/%s", serviceDesc.FullName(), methodDesc.Name()),
					RequestType:     string(methodDesc.Input().FullName()),
					ResponseType:    string(methodDesc.Output().FullName()),
					RequestFields:   extractFields(methodDesc.Input()),
					ResponseFields:  extractFields(methodDesc.Output()),
					ReadOnlyHint:    readOnlyHint,
					DestructiveHint: destructiveHint,
					IdempotentHint:  idempotentHint,
					OpenWorldHint:   openWorldHint,
				})
			}
		}
		return true
	})

	results.Tools = tools
	results.Prompts = prompts
	results.Resources = resources

	return results, nil
}

// extractFields iterates through the fields of a message descriptor and extracts
// them into a slice of McpField structs, including any field-level
// descriptions defined in MCP annotations.
func extractFields(msgDesc protoreflect.MessageDescriptor) []McpField {
	var fields []McpField
	for i := 0; i < msgDesc.Fields().Len(); i++ {
		field := msgDesc.Fields().Get(i)
		var description string
		if opts := field.Options(); opts != nil && proto.HasExtension(opts, mcpopt.E_FieldDescription) {
			description = proto.GetExtension(opts, mcpopt.E_FieldDescription).(string)
		}
		fields = append(fields, McpField{
			Name:        string(field.Name()),
			Description: description,
			Type:        field.Kind().String(),
			IsRepeated:  field.IsList(),
		})
	}
	return fields
}
