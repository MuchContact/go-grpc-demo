package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"

	pb "github.com/crazygit/go-grpc-demo/gen/greeting"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer
}

func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}
func (s *server) SayHelloTwo(ss pb.Greeter_SayHelloTwoServer) error {
	for {
		recv, err := ss.Recv()
		if err == io.EOF {
			fmt.Printf("end of stream")
			return nil
		} else if err != nil {
			return err
		}
		name := recv.GetName()
		fmt.Printf("Rec %v \n", name)
		response := &pb.HelloReply{
			Message: "Hello " + name,
		}
		ss.Send(response)
	}
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	// Register reflection service on gRPC server.

	reflection.Register(s)
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
