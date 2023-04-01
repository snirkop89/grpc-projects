package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"

	pb "github.com/snirkop89/grpc-projects/blog/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

var serverAddr = "localhost:50051"

func main() {
	tls := false
	var opts []grpc.DialOption

	if tls {
		certFile := "ssl/ca.crt"
		creds, err := credentials.NewClientTLSFromFile(certFile, "")
		if err != nil {
			log.Fatal(err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.Dial(serverAddr, opts...)
	if err != nil {
		log.Fatal("Failed dialing:", err)
	}
	defer conn.Close()

	c := pb.NewBlogServiceClient(conn)

	ctx := context.Background()

	id, err := createBlog(ctx, c)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Created blog with id: %q\n", id)

	log.Println("--- Reading valid id")
	blog, err := readBlog(ctx, c, id)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Got blog: %v\n", blog)

	log.Println("--- Invalid id")
	_, err = readBlog(ctx, c, "00-000000-000")
	if err != nil {
		log.Println("expected error:", err)
	}

	log.Println("Updating blog")
	blog.Title = "Updated title"
	err = updateBlog(ctx, c, blog)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(strings.Repeat("#", 50))
	log.Println("Listing all blogs...")
	err = listBlogs(ctx, c)
	if err != nil {
		log.Fatal(err)
	}

	err = deleteBlog(ctx, c, blog.Id)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("#####################", "Listing after delete")
	err = listBlogs(ctx, c)
	if err != nil {
		log.Fatal(err)
	}
}

func createBlog(ctx context.Context, c pb.BlogServiceClient) (string, error) {
	log.Println("createBlog invoked")

	blog := pb.Blog{
		AuthorId: "John",
		Title:    "My first blog",
		Content:  "Content of first blog",
	}
	res, err := c.CreateBlog(ctx, &blog)
	if err != nil {
		return "", fmt.Errorf("failed creating blog: %v", err)
	}

	return res.Id, nil
}

func readBlog(ctx context.Context, c pb.BlogServiceClient, id string) (*pb.Blog, error) {
	log.Println("readBlog invoked")

	req := pb.BlogId{Id: id}
	res, err := c.ReadBlog(ctx, &req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func updateBlog(ctx context.Context, c pb.BlogServiceClient, blog *pb.Blog) error {
	if _, err := c.UpdateBlog(ctx, blog); err != nil {
		return fmt.Errorf("failed updating blog: %v", err)
	}
	return nil
}

func listBlogs(ctx context.Context, c pb.BlogServiceClient) error {
	stream, err := c.ListBlogs(ctx, &emptypb.Empty{})
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

		log.Printf("Blog: %v\n", res)
	}

	return nil
}

func deleteBlog(ctx context.Context, c pb.BlogServiceClient, id string) error {
	log.Println("deleting blog with id:", id)
	data := pb.BlogId{Id: id}
	_, err := c.DeleteBlog(ctx, &data)
	if err != nil {
		return err
	}
	return nil
}
