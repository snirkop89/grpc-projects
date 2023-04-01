package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	pb "github.com/snirkop89/grpc-projects/blog/proto"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *server) CreateBlog(ctx context.Context, in *pb.Blog) (*pb.BlogId, error) {
	log.Printf("Create blog was invoked with %s\n", in)

	data := BlogItem{
		AuthorID: in.AuthorId,
		Title:    in.Title,
		Content:  in.Content,
	}

	res, err := collection.InsertOne(ctx, data)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("internal error: %v", err),
		)
	}

	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, status.Errorf(
			codes.Internal,
			"internal error: invalid id returned",
		)
	}

	return &pb.BlogId{Id: oid.Hex()}, nil
}

func (s *server) ReadBlog(ctx context.Context, in *pb.BlogId) (*pb.Blog, error) {
	log.Printf("ReadBlog invoked with %v\n", in)

	oid, err := s.oidFromReq(in.Id)
	if err != nil {
		return nil, s.errInvalidArgument()
	}

	var data BlogItem
	filter := bson.M{"_id": oid}

	res := collection.FindOne(ctx, filter)
	if err := res.Decode(&data); err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			"Cannot find blog with the provided id",
		)
	}

	return data.toPbBlog(), nil
}

func (s *server) UpdateBlog(ctx context.Context, in *pb.Blog) (*empty.Empty, error) {
	log.Printf("UpdateBlog invoked with %v\n", in)

	oid, err := s.oidFromReq(in.Id)
	if err != nil {
		return nil, s.errInvalidArgument()
	}

	var data = BlogItem{
		AuthorID: in.AuthorId,
		Title:    in.Title,
		Content:  in.Content,
	}
	res, err := collection.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": data})
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Could not update",
		)
	}
	if res.MatchedCount == 0 {
		return nil, status.Errorf(codes.NotFound, "Cannot find blog with provided id")
	}

	return &emptypb.Empty{}, nil
}

func (s *server) ListBlogs(_ *empty.Empty, stream pb.BlogService_ListBlogsServer) error {
	log.Println("ListBlogs invoked")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, primitive.D{{}})
	if err != nil {
		log.Printf("unknown find error: %v\n", err)
		return status.Errorf(codes.Internal, "Unknown internal error")
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var data BlogItem
		err := cursor.Decode(&data)
		if err != nil {
			return status.Errorf(codes.Internal, "Error while decoding data from database")
		}
		stream.Send(data.toPbBlog())
	}

	if err := cursor.Err(); err != nil {
		return status.Errorf(codes.Internal, "unknown internal error")
	}

	return nil
}

func (s *server) DeleteBlog(ctx context.Context, in *pb.BlogId) (*empty.Empty, error) {
	log.Printf("DeleteBlog invoked with %v\n", in)

	oid, err := s.oidFromReq(in.Id)
	if err != nil {
		return nil, s.errInvalidArgument()
	}

	_, err = collection.DeleteOne(ctx, bson.M{"_id": oid})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unknown internal error")
	}

	return &emptypb.Empty{}, nil
}

func (s *server) oidFromReq(id string) (primitive.ObjectID, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return primitive.ObjectID{}, err
	}
	return oid, nil
}

func (s *server) errInvalidArgument() error {
	return status.Errorf(
		codes.InvalidArgument,
		"Cannot parse id",
	)
}
