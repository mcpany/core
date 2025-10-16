/*
 * Copyright 2025 Author(s) of MCP-XY
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

package protobufparser

import (
	"context"
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	reflectpb "google.golang.org/grpc/reflection/grpc_reflection_v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"

	mcpopt "github.com/mcpxy/core/proto/mcp_options/v1"
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
	defer conn.Close()

	return parseProtoWithExistingConnection(ctx, conn)
}

// ParseProtoWithExistingConnection performs reflection on an existing gRPC
// connection. It is a lower-level function that allows for more control over
// the connection lifecycle, which is useful in testing or scenarios where a
// connection is already established.
func parseProtoWithExistingConnection(ctx context.Context, conn *grpc.ClientConn) (*descriptorpb.FileDescriptorSet, error) {
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
				if methodOpts == nil {
					continue
				}

				if proto.HasExtension(methodOpts, mcpopt.E_ToolName) {
					toolName := proto.GetExtension(methodOpts, mcpopt.E_ToolName).(string)
					toolDesc := ""
					if proto.HasExtension(methodOpts, mcpopt.E_ToolDescription) {
						toolDesc = proto.GetExtension(methodOpts, mcpopt.E_ToolDescription).(string)
					}

					readOnlyHint := proto.GetExtension(methodOpts, mcpopt.E_McpToolReadonlyHint).(bool)
					destructiveHint := proto.GetExtension(methodOpts, mcpopt.E_McpToolDestructiveHint).(bool)
					idempotentHint := proto.GetExtension(methodOpts, mcpopt.E_McpToolIdempotentHint).(bool)
					openWorldHint := proto.GetExtension(methodOpts, mcpopt.E_McpToolOpenworldHint).(bool)

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
