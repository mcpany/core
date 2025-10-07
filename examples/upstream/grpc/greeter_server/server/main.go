package main

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
	pb "github.com/mcpxy/core/examples/upstream/grpc/greeter_server/proto"
)

const (
	port = ":50051"
)

// server is used to implement greeter.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements greeter.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
