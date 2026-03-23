import re

def process_file(filepath, replacements):
    with open(filepath, 'r') as f:
        content = f.read()

    for old, new in replacements.items():
        content = content.replace(old, new)

    with open(filepath, 'w') as f:
        f.write(content)

replacements = {
    # examples/upstream_service_demo/grpc/greeter_server/proto/greeter_grpc.pb.go
    'type GreeterClient interface {': '// GreeterClient represents a client for the Greeter service.\n//\n// Summary: Represents a Greeter service client.\ntype GreeterClient interface {',
    'func NewGreeterClient(cc grpc.ClientConnInterface) GreeterClient {': '''// NewGreeterClient creates a new Greeter client.
//
// Summary: Creates a new Greeter client.
//
// Parameters:
//   - cc (grpc.ClientConnInterface): The client connection interface.
//
// Returns:
//   - GreeterClient: The client instance.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewGreeterClient(cc grpc.ClientConnInterface) GreeterClient {''',
    'func (c *greeterClient) SayHello(ctx context.Context, in *HelloRequest, opts ...grpc.CallOption) (*HelloReply, error) {': '''// SayHello sends a greeting request.
//
// Summary: Sends a greeting request.
//
// Parameters:
//   - ctx (context.Context): The context for the request.
//   - in (*HelloRequest): The input request.
//   - opts (...grpc.CallOption): The call options.
//
// Returns:
//   - *HelloReply: The greeting reply.
//   - error: An error if the request fails.
//
// Errors:
//   - Returns error if the underlying RPC fails.
//
// Side Effects:
//   - Initiates an RPC call.
func (c *greeterClient) SayHello(ctx context.Context, in *HelloRequest, opts ...grpc.CallOption) (*HelloReply, error) {''',
    'type GreeterServer interface {': '// GreeterServer represents a server for the Greeter service.\n//\n// Summary: Represents a Greeter service server.\ntype GreeterServer interface {',
    'type UnimplementedGreeterServer struct{}': '// UnimplementedGreeterServer is a default implementation of the GreeterServer interface.\n//\n// Summary: Provides a default, unimplemented GreeterServer.\ntype UnimplementedGreeterServer struct{}',
    'func (UnimplementedGreeterServer) SayHello(context.Context, *HelloRequest) (*HelloReply, error) {': '''// SayHello is the default unimplemented SayHello.
//
// Summary: Default unimplemented SayHello.
//
// Parameters:
//   - _ (context.Context): The context.
//   - _ (*HelloRequest): The request.
//
// Returns:
//   - *HelloReply: The reply.
//   - error: An unimplemented error.
//
// Errors:
//   - Returns codes.Unimplemented.
//
// Side Effects:
//   - None.
func (UnimplementedGreeterServer) SayHello(context.Context, *HelloRequest) (*HelloReply, error) {''',
    'type UnsafeGreeterServer interface {': '// UnsafeGreeterServer represents an unsafe Greeter Server interface without forward compatibility.\n//\n// Summary: Represents an Unsafe Greeter Server.\ntype UnsafeGreeterServer interface {',
    'func RegisterGreeterServer(s grpc.ServiceRegistrar, srv GreeterServer) {': '''// RegisterGreeterServer registers the Greeter server.
//
// Summary: Registers the Greeter server.
//
// Parameters:
//   - s (grpc.ServiceRegistrar): The registrar.
//   - srv (GreeterServer): The server instance.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Registers the service with the gRPC server.
func RegisterGreeterServer(s grpc.ServiceRegistrar, srv GreeterServer) {''',
    'var Greeter_ServiceDesc = grpc.ServiceDesc{': '// Greeter_ServiceDesc describes the Greeter service for grpc.RegisterService.\n//\n// Summary: Describes the Greeter service.\nvar Greeter_ServiceDesc = grpc.ServiceDesc{',

    # examples/upstream_service_demo/grpc/greeter_server/proto/greeter.pb.go
    'type HelloRequest struct {': '// HelloRequest represents a request containing a name.\n//\n// Summary: Represents a Hello Request.\ntype HelloRequest struct {',
    'func (x *HelloRequest) Reset() {': '''// Reset clears the message.
//
// Summary: Clears the message.
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
//   - Modifies the receiver.
func (x *HelloRequest) Reset() {''',
    'func (x *HelloRequest) String() string {': '''// String returns a string representation.
//
// Summary: Returns a string representation.
//
// Parameters:
//   - None.
//
// Returns:
//   - string: The string representation.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (x *HelloRequest) String() string {''',
    'func (*HelloRequest) ProtoMessage() {}': '''// ProtoMessage marks this as a proto message.
//
// Summary: Marks as proto message.
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
func (*HelloRequest) ProtoMessage() {}''',
    'func (x *HelloRequest) ProtoReflect() protoreflect.Message {': '''// ProtoReflect returns reflection information.
//
// Summary: Returns reflection information.
//
// Parameters:
//   - None.
//
// Returns:
//   - protoreflect.Message: The reflection message.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (x *HelloRequest) ProtoReflect() protoreflect.Message {''',
    'func (x *HelloRequest) GetName() string {': '''// GetName gets the name.
//
// Summary: Gets the name.
//
// Parameters:
//   - None.
//
// Returns:
//   - string: The name.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (x *HelloRequest) GetName() string {''',
    'type HelloReply struct {': '// HelloReply represents a response containing a greeting message.\n//\n// Summary: Represents a Hello Reply.\ntype HelloReply struct {',
    'func (x *HelloReply) Reset() {': '''// Reset clears the message.
//
// Summary: Clears the message.
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
//   - Modifies the receiver.
func (x *HelloReply) Reset() {''',
    'func (x *HelloReply) String() string {': '''// String returns a string representation.
//
// Summary: Returns a string representation.
//
// Parameters:
//   - None.
//
// Returns:
//   - string: The string representation.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (x *HelloReply) String() string {''',
    'func (*HelloReply) ProtoMessage() {}': '''// ProtoMessage marks this as a proto message.
//
// Summary: Marks as proto message.
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
func (*HelloReply) ProtoMessage() {}''',
    'func (x *HelloReply) ProtoReflect() protoreflect.Message {': '''// ProtoReflect returns reflection information.
//
// Summary: Returns reflection information.
//
// Parameters:
//   - None.
//
// Returns:
//   - protoreflect.Message: The reflection message.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (x *HelloReply) ProtoReflect() protoreflect.Message {''',
    'func (x *HelloReply) GetMessage() string {': '''// GetMessage gets the message.
//
// Summary: Gets the message.
//
// Parameters:
//   - None.
//
// Returns:
//   - string: The message.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (x *HelloReply) GetMessage() string {''',
    'var File_greeter_proto protoreflect.FileDescriptor': '// File_greeter_proto is the proto reflection file descriptor.\n//\n// Summary: Proto reflection file descriptor.\nvar File_greeter_proto protoreflect.FileDescriptor',

    # examples/upstream_service_demo/grpc/greeter_server/server/main.go
    'func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {': '''// SayHello responds to a greeting request.
//
// Summary: Processes a SayHello request and returns a greeting.
//
// Parameters:
//   - ctx (context.Context): The context for the request.
//   - in (*pb.HelloRequest): The incoming HelloRequest.
//
// Returns:
//   - *pb.HelloReply: The greeting response.
//   - error: An error if the request fails.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Logs the received name.
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {''',

    # examples/upstream_service_demo/webrtc/server/main.go and client/main.go
    'type Signal struct {': '// Signal represents a WebRTC signaling message.\n//\n// Summary: Represents a WebRTC signaling message.\ntype Signal struct {',

    # examples/demo/stdio/my-tool/main.go
    'type Request struct {': '// Request represents an incoming stdio request.\n//\n// Summary: Represents an incoming stdio request.\ntype Request struct {',
    'type Response struct {': '// Response represents an outgoing stdio response.\n//\n// Summary: Represents an outgoing stdio response.\ntype Response struct {',

    # docs/features/webhooks/examples/html_to_md/main.go & block_rm/main.go
    'type WebhookRequest struct {': '// WebhookRequest represents an incoming webhook payload.\n//\n// Summary: Represents an incoming webhook payload.\ntype WebhookRequest struct {',
    'type WebhookResponse struct {': '// WebhookResponse represents an outgoing webhook payload.\n//\n// Summary: Represents an outgoing webhook payload.\ntype WebhookResponse struct {',
    'type StatusOK struct{}': '// StatusOK represents a successful status response payload.\n//\n// Summary: Represents a successful status response payload.\ntype StatusOK struct{}',
    'const StatusOK = 200': '// StatusOK represents the HTTP 200 OK status code.\n//\n// Summary: Represents the HTTP 200 OK status code.\nconst StatusOK = 200',
    'type Status struct {': '// Status represents a status response payload.\n//\n// Summary: Represents a status response payload.\ntype Status struct {'
}

for filepath in [
    'server/examples/upstream_service_demo/grpc/greeter_server/proto/greeter_grpc.pb.go',
    'server/examples/upstream_service_demo/grpc/greeter_server/proto/greeter.pb.go',
    'server/examples/upstream_service_demo/grpc/greeter_server/server/main.go',
    'server/examples/upstream_service_demo/webrtc/server/main.go',
    'server/examples/upstream_service_demo/webrtc/client/main.go',
    'server/examples/demo/stdio/my-tool/main.go',
    'server/docs/features/webhooks/examples/html_to_md/main.go',
    'server/docs/features/webhooks/examples/block_rm/main.go'
]:
    process_file(filepath, replacements)

# Fix Descriptor for HelloRequest and HelloReply separately via regex
filepath = 'server/examples/upstream_service_demo/grpc/greeter_server/proto/greeter.pb.go'
with open(filepath, 'r') as f:
    content = f.read()

content = re.sub(r'func \(\*HelloRequest\) Descriptor\(\) \(\[\]byte, \[\]int\) \{', r'''// Descriptor gets the protobuf message descriptor.
//
// Summary: Gets the protobuf message descriptor.
//
// Parameters:
//   - None.
//
// Returns:
//   - []byte: The raw descriptor.
//   - []int: The descriptor index.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (*HelloRequest) Descriptor() ([]byte, []int) {''', content)

content = re.sub(r'func \(\*HelloReply\) Descriptor\(\) \(\[\]byte, \[\]int\) \{', r'''// Descriptor gets the protobuf message descriptor.
//
// Summary: Gets the protobuf message descriptor.
//
// Parameters:
//   - None.
//
// Returns:
//   - []byte: The raw descriptor.
//   - []int: The descriptor index.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (*HelloReply) Descriptor() ([]byte, []int) {''', content)

with open(filepath, 'w') as f:
    f.write(content)
