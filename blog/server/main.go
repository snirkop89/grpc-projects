package main

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "github.com/snirkop89/grpc-projects/blog/proto"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var addr = "0.0.0.0:50051"

var collection *mongo.Collection

type BlogItem struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	AuthorID string             `bson:"author_id"`
	Title    string             `bson:"title"`
	Content  string             `bson:"content"`
}

func (bi BlogItem) toPbBlog() *pb.Blog {
	return &pb.Blog{
		Id:       bi.ID.Hex(),
		AuthorId: bi.AuthorID,
		Title:    bi.Title,
		Content:  bi.Content,
	}
}

type server struct {
	pb.UnimplementedBlogServiceServer
}

func main() {
	if err := initMongoClient(); err != nil {
		log.Fatal(err)
	}

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", addr, err)
	}

	log.Printf("Listening at: %s\n", addr)

	var opts []grpc.ServerOption
	tls := false // Change to false if needed
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
	pb.RegisterBlogServiceServer(s, srv)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed serving: %v", err)
	}

}

func initMongoClient() error {
	mongoClient, err := mongo.NewClient(options.Client().ApplyURI(
		"mongodb://blogapp:password@localhost:27017/",
	))
	if err != nil {
		return fmt.Errorf("failed creating mongo client: %v", err)
	}
	err = mongoClient.Connect(context.Background())
	if err != nil {
		return err
	}
	collection = mongoClient.Database("blogdb").Collection("blog")

	return nil
}
