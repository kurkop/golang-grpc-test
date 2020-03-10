package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/kurkop/golang-grpc-test/greet/greetpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

func main() {
	fmt.Println("Hello I'm a client")

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

	cc, err := grpc.Dial("localhost:50052", opts)
	if err != nil {
		log.Fatalf("Could not connect: %+v", err)
	}
	defer cc.Close()

	c := greetpb.NewGreetServiceClient(cc)
	// fmt.Printf("Create client: %#v", c)

	// doUnary(c)
	// doServerStreaming(c)
	// doClientStreaming(c)
	// doBiDiStreaming(c)
	doUnaryWithDeadline(c)
}

func doUnary(c greetpb.GreetServiceClient) {
	fmt.Println("Starting to do a Unary RPC...")
	req := &greetpb.GreetRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Juan",
			LastName:  "Arias",
		},
	}
	res, err := c.Greet(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while calling Gret RPC: %v", err)
	}
	log.Printf("Response from Greet: %v", res.Result)
}

func doServerStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Starting to do a server streaming RPC...")
	req := &greetpb.GreetManyTimesRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Juan",
			LastName:  "Arias",
		},
	}
	resStream, err := c.GreetManyTimes(context.Background(), req)
	if err != nil {
		log.Fatalf("erro while calling GreetManyTimes RPC: %v", err)
	}
	for {
		msg, err := resStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("erro while reading stream: %v", err)
		}
		log.Printf("Response from GetManyTimes: %v", msg.GetResult())
	}

}

func doClientStreaming(c greetpb.GreetServiceClient) {
	log.Printf("Starting to do a Client Streaming RPC")
	request := []*greetpb.LongGreetRequest{
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "JP",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Andres",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "John",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Ruth",
			},
		},
	}
	stream, err := c.LongGreet(context.Background())
	if err != nil {
		log.Fatalf("error while calling LongGreet: %v", err)
	}
	for _, req := range request {
		fmt.Printf("Sending req: %v\n", req)
		stream.Send(req)
		time.Sleep(100 * time.Microsecond)
	}
	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Error while receiving response from LongGreet: %v", err)
	}

	fmt.Printf("LongGreet Response %v\n", res)

}

func doBiDiStreaming(c greetpb.GreetServiceClient) {
	log.Printf("Starting to do a BiDi Streaming RPC")

	// We create a stream by invoking the client
	stream, err := c.GreetEveryone(context.Background())
	if err != nil {
		log.Fatalf("Error while creating stream: %v", err)
		return
	}

	requests := []*greetpb.GreetEveryonRequest{
		&greetpb.GreetEveryonRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "JP",
			},
		},
		&greetpb.GreetEveryonRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Andres",
			},
		},
		&greetpb.GreetEveryonRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "John",
			},
		},
		&greetpb.GreetEveryonRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Ruth",
			},
		},
	}

	waitc := make(chan struct{})
	// We send a much of messages to the client (go routine)
	go func() {
		for _, req := range requests {
			fmt.Printf("Sending: %v\n", req)
			stream.Send(req)
			time.Sleep(1000 * time.Millisecond)
		}
		stream.CloseSend()
	}()

	// We receive a bunch of messages from the client (go routine)
	go func() {

		for {
			res, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("Error while receiving: %v", err)
				break
			}
			fmt.Printf("Received: %v\n", res.GetResult())
		}
		close(waitc)

	}()

	<-waitc

	// Block until everthing is done
}

func doUnaryWithDeadline(c greetpb.GreetServiceClient) {
	fmt.Println("Starting to do a UnaryWithDeadline RPC...")
	req := &greetpb.GreetWithDeadLineRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Juan",
			LastName:  "Arias",
		},
	}
	// ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := c.GreetWithDeadLine(ctx, req)
	if err != nil {

		statusError, ok := status.FromError(err)
		if ok {
			if statusError.Code() == codes.DeadlineExceeded {
				log.Fatalf("Timeout was hit! Deadline was exceeded")
			} else {
				fmt.Printf("unexpeted error: %v", statusError)
			}
			return
		} else {
			log.Fatalf("Error while calling GreetWithDeadline RPC: %v", err)
			return
		}
	}
	log.Printf("Response from Greet: %v", res.GetResponse())
}
