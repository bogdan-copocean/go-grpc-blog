package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	protos "github.com/bogdan-user/go-grpc-react-blog/proto"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

func main() {
	// if we crash the go code, we get the file name and line number
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	fmt.Println("Connecting to MongoDB...")

	// connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}

	Collection = client.Database("mydb").Collection("blog")

	fmt.Println("Blog Service started...")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v\n", err)
	}

	certFile := "ssl/server.crt"
	keyFile := "ssl/server.key"
	creds, sslErr := credentials.NewServerTLSFromFile(certFile, keyFile)

	if sslErr != nil {
		log.Fatalf("Failed loading certificates: %v\n", sslErr)
		return
	}

	opts := grpc.Creds(creds)

	server := NewServer()
	s := grpc.NewServer(opts)
	protos.RegisterBlogServiceServer(s, server)

	// Register reflection service on gRPC server.
	reflection.Register(s)
	
	go func() {
		fmt.Println("Starting gRPC Server...")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v\n", err)
		}
	}()

	// Wrap the GRPC Server in grpc-web and also host the UI
	grpcWebServer := grpcweb.WrapServer(s)
	// Lets put the wrapped grpc server in our multiplexer struct so
	// it can reach the grpc server in its handler
	multiplex := grpcMultiplexer{
		grpcWebServer,
	}

	// We need a http router
	r := http.NewServeMux()
	// Load the static webpage with a http fileserver
	webapp := http.FileServer(http.Dir("ui/build"))
	// Host the Web Application at /, and wrap it in the GRPC Multiplexer
	// This allows grpc requests to transfer over HTTP1. then be
	// routed by the multiplexer
	r.Handle("/", multiplex.Handler(webapp))
	// Create a HTTP server and bind the router to it, and set wanted address
	srv := &http.Server{
		Handler:      r,
		Addr:         "localhost:8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	// Serve the webapp
	go func() {
		fmt.Println("Starting HTTP Server...\n")
		if err := srv.ListenAndServeTLS("ssl/server.crt", "ssl/server.key"); err != nil {			
			log.Fatalf("Failed to serve: %v\n", err)
		}
	}()

	// Wait for ctrl+c to exit
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	// Block until signal is received
	<-ch
	fmt.Println("Stopping the server...")
	s.Stop()
	fmt.Println("Stopping the listener...")
	lis.Close()
	fmt.Println("Closing MongoDB connection")
	client.Disconnect(context.TODO())
	fmt.Println("Stopping the http server...")
	srv.Close()
}
