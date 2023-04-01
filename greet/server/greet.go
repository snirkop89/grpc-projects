package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	pb "github.com/snirkop89/grpc-projects/greet/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *server) Greet(ctx context.Context, in *pb.GreetRequest) (*pb.GreetResponse, error) {
	log.Printf("Greet function was invoked with %v\n", in)
	return &pb.GreetResponse{
		Result: "Hello " + in.FirstName,
	}, nil
}

func (s *server) GreetManyTimes(in *pb.GreetRequest, stream pb.GreetService_GreetManyTimesServer) error {
	log.Printf("GreetManyTimes invoked with: %v\n", in)

	for i := 0; i < 10; i++ {
		res := fmt.Sprintf("Hello %s! [%d]", in.FirstName, i+1)
		err := stream.Send(&pb.GreetResponse{
			Result: res,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *server) LongGreet(stream pb.GreetService_LongGreetServer) error {
	log.Printf("LongGreet() Invoked")

	var res string
	for {
		req, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}

		res += fmt.Sprintf("Hello %s!\n", req.FirstName)
	}

	if err := stream.SendAndClose(&pb.GreetResponse{
		Result: res,
	}); err != nil {
		return fmt.Errorf("sendAndClose(): %v", err)
	}

	return nil
}

func (s *server) GreetEveryone(stream pb.GreetService_GreetEveryoneServer) error {
	log.Println("GreetEveryone invoked")

	for {
		req, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("failed reading client stream: %v", err)
		}

		res := fmt.Sprintf("Hello %s!", req.FirstName)
		err = stream.Send(&pb.GreetResponse{Result: res})
		if err != nil {
			return fmt.Errorf("erroring sending data to client: %v", err)
		}
	}

	return nil
}

func (s *server) GreetWithDeadline(ctx context.Context, in *pb.GreetRequest) (*pb.GreetResponse, error) {
	log.Println("GreetWithDeadline invoked")

	for i := 0; i < 2; i++ {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("The client canceled the rquest!")
			return nil, status.Error(codes.Canceled, "The client canceled the request")
		}
		time.Sleep(time.Second)
	}
	return &pb.GreetResponse{
		Result: "Hello " + in.FirstName,
	}, nil
}
