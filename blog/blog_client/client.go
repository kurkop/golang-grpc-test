package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/kurkop/golang-grpc-test/blog/blogpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	fmt.Println("Blog Client")

	tls := false
	opts := grpc.WithInsecure()

	if tls {
		certFile := "ssl/ca.crt"
		creds, sslErr := credentials.NewClientTLSFromFile(certFile, "")
		if sslErr != nil {
			log.Fatalf("Error while loading CA trust certificate: %v", sslErr)
		}
		opts = grpc.WithTransportCredentials(creds)
	}

	cc, err := grpc.Dial("localhost:50055", opts)
	if err != nil {
		log.Fatalf("Could not connect: %+v", err)
	}
	defer cc.Close()

	c := blogpb.NewBlogServiceClient(cc)
	// Create Blog
	fmt.Println("Creating the blog")
	blog := blogpb.Blog{
		AuthorId: "Stephane",
		Title:    "My First Blog",
		Content:  "Content of the first blog",
	}
	createBlogRes, err := c.CreateBlog(context.Background(), &blogpb.CreateBlogRequest{Blog: &blog})
	if err != nil {
		log.Fatalf("Unexpected error: %v", err)
	}
	fmt.Printf("Blog has been created: %v", createBlogRes)

	// Read Blog
	fmt.Println("Reading the blog")
	_, err2 := c.ReadBlog(context.Background(), &blogpb.ReadBlogRequest{BlogId: "13wfasdf"})
	if err2 != nil {
		fmt.Printf("Error happend while reading: %v\n", err2)
	}

	readBlogRes, err3 := c.ReadBlog(context.Background(), &blogpb.ReadBlogRequest{BlogId: createBlogRes.GetBlog().GetId()})
	if err3 != nil {
		fmt.Printf("Error happend while reading2: %v\n", err3)
	}
	fmt.Printf("Blog has gotten: %v\n", readBlogRes)

	// Update Blog
	fmt.Println("Update the blog")
	newBlog := &blogpb.Blog{
		Id:       createBlogRes.GetBlog().GetId(),
		AuthorId: "Stephane",
		Title:    "My First Blog (Edited)",
		Content:  "Content of the first blog, with some awesome addtions!",
	}

	updateRes, updateErr := c.UpdateBlog(context.Background(), &blogpb.UpdateBlogRequest{Blog: newBlog})
	if updateErr != nil {
		fmt.Printf("Error happend while updating: %v\n", updateErr)
	}
	fmt.Printf("Blog was updated: %v\n", updateRes)

	// Delete Blog
	fmt.Println("Delete Blog")
	deleteRes, deleteErr := c.DeleteBlog(context.Background(), &blogpb.DeleteBlogRequest{BlogId: createBlogRes.GetBlog().GetId()})
	if deleteErr != nil {
		fmt.Printf("Error happend while deleting: %v\n", deleteErr)
	}
	fmt.Printf("Blog was deleted: %v\n", deleteRes)

	// List Blogs
	fmt.Println("List Blogs")

	listStream, err := c.ListBlog(context.Background(), &blogpb.ListBlogRequest{})
	if err != nil {
		log.Fatalf("Error while calling Gret RPC: %v", err)
	}

	for {
		msg, err := listStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("erro while reading stream: %v", err)
		}
		log.Printf("Response from ListBlog: %v", msg)
	}

}
