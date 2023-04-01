package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"

	pb "github.com/snirkop89/grpc-projects/greet/proto"
)

var serverAddr = "localhost:50051"

func main() {
	tls := true
	var opts []grpc.DialOption

	if tls {
		certFile := "ssl/ca.crt"
		creds, err := credentials.NewClientTLSFromFile(certFile, "")
		if err != nil {
			log.Fatal(err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	}

	conn, err := grpc.Dial(serverAddr, opts...)
	if err != nil {
		log.Fatal("Failed dialing:", err)
	}
	defer conn.Close()

	client := pb.NewGreetServiceClient(conn)

	err = doGreet(context.TODO(), client)
	if err != nil {
		log.Fatal(err)
	}

	err = doGreetManyTimes(context.TODO(), client)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(strings.Repeat("#", 20))
	err = doLongGreet(context.TODO(), client)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(strings.Repeat("#", 20))
	err = doGreetEveryone(context.TODO(), client)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(strings.Repeat("#", 20))
	log.Println("Request will succeed")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = doGreetWithDeadline(ctx, client)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Request will fail due to deadline")
	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = doGreetWithDeadline(ctx, client)
	if err != nil {
		log.Fatal(err)
	}
}

func doGreet(ctx context.Context, client pb.GreetServiceClient) error {
	log.Println("doGreet invoked")
	res, err := client.Greet(ctx, &pb.GreetRequest{
		FirstName: "John",
	})
	if err != nil {
		return err
	}

	log.Printf("Greeting: %s\n", res.Result)
	return nil
}

func doGreetManyTimes(ctx context.Context, client pb.GreetServiceClient) error {
	log.Println("doGreetManyTime invoked")

	stream, err := client.GreetManyTimes(ctx, &pb.GreetRequest{
		FirstName: "Maria",
	})
	if err != nil {
		return err
	}

	for {
		res, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		log.Printf("Received: %s\n", res.Result)
	}
	return nil
}

func doLongGreet(ctx context.Context, client pb.GreetServiceClient) error {
	log.Println("doLongGreet invoked")

	reqs := []*pb.GreetRequest{
		{FirstName: "John"},
		{FirstName: "Maria"},
		{FirstName: "Shawn"},
	}
	stream, err := client.LongGreet(ctx)
	if err != nil {
		return err
	}

	for _, req := range reqs {
		log.Printf("Sending req: %v\n", req)
		if err := stream.Send(req); err != nil {
			return err
		}
		time.Sleep(time.Second)
	}

	res, err := stream.CloseAndRecv()
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("closeAndRecv(): %v", err)
	}
	log.Printf("Response: %s\n", res.Result)

	return nil
}

func doGreetEveryone(ctx context.Context, client pb.GreetServiceClient) error {
	log.Println("doGreetEveryone invoked")

	stream, err := client.GreetEveryone(ctx)
	if err != nil {
		return err
	}

	reqs := []*pb.GreetRequest{
		{FirstName: "John"},
		{FirstName: "Maria"},
		{FirstName: "Shawn"},
	}

	// Send and receive simultaneously
	waitc := make(chan struct{})

	go func() {
		for _, req := range reqs {
			log.Printf("Sending request: %v\n", req)
			stream.Send(req)
			time.Sleep(time.Second)
		}
		stream.CloseSend()
	}()

	go func() {
		for {
			res, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				close(waitc)
				break
			}
			if err != nil {
				log.Printf("Error while receiving: %v\n", err)
				continue
			}
			log.Printf("Received: %v\n", res.Result)
		}
	}()

	<-waitc

	return nil
}

func doGreetWithDeadline(ctx context.Context, client pb.GreetServiceClient) error {
	log.Println("doGreetWithDeadline invoked")

	res, err := client.GreetWithDeadline(ctx, &pb.GreetRequest{
		FirstName: "John",
	})
	if err != nil {
		e, ok := status.FromError(err)
		if ok {
			// grpc error
			if e.Code() == codes.DeadlineExceeded {
				log.Println("Deadline exceeded!")
				return nil
			}
		}
		// non grpc error
		return fmt.Errorf("unexpected error: %v", err)
	}

	log.Printf("Received: %v\n", res.Result)

	return nil
}
