package main

import (
	"log"
	"net"

	pb "github.com/snirkop89/grpc-projects/greet/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

var (
	addr string = "0.0.0.0:50051"
)

type server struct {
	// pb.GreetServiceServer
	pb.UnimplementedGreetServiceServer
}

var _ pb.GreetServiceServer = (*server)(nil)

func main() {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", addr, err)
	}

	log.Printf("Listening at: %s\n", addr)

	var opts []grpc.ServerOption
	tls := true // Change to false if needed
	if tls {
		certFile := "ssl/server.crt"
		keyFile := "ssl/server.pem"
		creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
		if err != nil {
			log.Fatal("failed loading certificates:", err)
		}
		opts = append(opts, grpc.Creds(creds))
	}

	var srv *server
	s := grpc.NewServer(opts...)
	pb.RegisterGreetServiceServer(s, srv)
	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed serving: %v", err)
	}

}
