package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"strconv"
	"time"

	pb "github.com/crazygit/go-grpc-demo/gen/greeting"
	"google.golang.org/grpc"
)

const (
	defaultName = "world"
)

var (
	addr = flag.String("addr", "localhost:50051", "the address to connect to")
	name = flag.String("name", defaultName, "Name to greet")
)

func main() {
	flag.Parse()
	// Set up a connection to the server.
	conn, err := grpc.Dial(*addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: *name})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.GetMessage())

	twoStream, err := c.SayHelloTwo(ctx)
	if err != nil {
		fmt.Errorf("", err)
	}
	for i := 0; i < 10; i++ {
		twoStream.Send(&pb.HelloRequest{Name: "user-" + strconv.Itoa(i)})
	}
	twoStream.CloseSend()
	for {
		recv, err := twoStream.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Errorf("", err)
			break
		} else {

			fmt.Printf("Rec %v \n", recv.Message)
		}
	}
	fmt.Printf("close send...")
	select {}
}
