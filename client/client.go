package main

import (
	"context"
	"fmt"
	"io"
	"log"

	protos "github.com/bogdan-user/go-grpc-react-blog/proto"
	"google.golang.org/grpc"
)

func main() {

	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not connect to: %v\n", err)
	}
	defer cc.Close()

	c := protos.NewBlogServiceClient(cc)

	CreateBlogUnary(c)
	// ReadBlogUnary(c)
	// UpdateBlogUnary(c)
	// DeleteBlogUnary(c)
	// ListAllBlogUnary(c)

}

func CreateBlogUnary(c protos.BlogServiceClient) {
	fmt.Println("Creating a blog post...")

	blogRequest := protos.CreateBlogRequest{
		Blog: &protos.Blog{
			AuthorId: "Test to see how it works",
			Title:    "Test to see how it works",
			Content:  "Test to see how it works",
		},
	}

	res, err := c.CreateBlog(context.Background(), &blogRequest)
	if err != nil {
		log.Fatalf("Error while creating the blog: %v\n", err)
	}

	log.Printf("Response from CreateBlog RPC: %v\n", res.Blog)
}

func ReadBlogUnary(c protos.BlogServiceClient) {
	fmt.Println("Reading the blog...")

	blogRequest := protos.ReadBlogRequest{
		BlogId: "612799388e0a217939049563",
	}

	res, err := c.ReadBlog(context.Background(), &blogRequest)
	if err != nil {
		log.Fatalf("Error while reading the blog with the id %v: %v\n", blogRequest.BlogId, err)
	}

	log.Printf("Response from ReadBlog RPC: %v\n", res.Blog)
}

func UpdateBlogUnary(c protos.BlogServiceClient) {
	fmt.Println("Reading the blog...")

	blogID := "612b7a5ce047da0dc238f89b"

	blogRequest := protos.UpdateBlogRequest{
		Blog: &protos.Blog{
			Id:       blogID,
			AuthorId: "Changed",
			Title:    "Changed",
			Content:  "Changed",
		},
	}

	res, err := c.UpdateBlog(context.Background(), &blogRequest)

	if err != nil {
		log.Fatalf("Error while updating the blog with the id %v: %v\n", blogRequest.Blog.GetId(), err)
	}

	log.Printf("Response from UpdateBlog RPC: %v\n", res.Blog)
}

func DeleteBlogUnary(c protos.BlogServiceClient) {
	fmt.Println("Deleting the blog...")

	blogID := "61278773c47c2cefe0ba555a"

	blogRequest := protos.DeleteBlogRequest{
		BlogId: blogID,
	}

	res, err := c.DeleteBlog(context.Background(), &blogRequest)

	if err != nil {
		log.Fatalf("Error while deleting the blog: %v\n", err)
	}

	log.Printf("Response from DeleteBlog RPC: %v\n", res.BlogId)
}

func ListAllBlogUnary(c protos.BlogServiceClient) {
	fmt.Println("List blogs...")

	blogRequest := protos.ListBlogRequest{}

	stream, err := c.ListBlog(context.Background(), &blogRequest)

	if err != nil {
		log.Fatalf("Error while listing the blogs: %v\n", err)
	}

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("Error while receiving: %v\n", err)
			break
		}

		log.Printf("Response from ListAllBlog RPC: %v\n", res.Blog)

	}
	log.Println("ListAllBlog RPC done")

}
