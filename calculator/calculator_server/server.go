package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"net"

	"github.com/kurkop/golang-grpc-test/calculator/calculatorpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type server struct {
}

func (*server) Sum(ctx context.Context, req *calculatorpb.SumRequest) (*calculatorpb.SumResponse, error) {
	value1 := req.GetValue1()
	value2 := req.GetValue2()
	fmt.Printf("Running Sum function: %v and %v", value1, value2)
	res := &calculatorpb.SumResponse{Result: value1 + value2}
	return res, nil
}

func (*server) PrimeNumberDecomposition(req *calculatorpb.PrimeNumberDecompositionRequest, stream calculatorpb.CalculatorService_PrimeNumberDecompositionServer) error {
	number := req.GetNumber()
	divisor := int32(2)
	fmt.Println("Running PrimeNumberDecompositio Stream")
	for number > 1 {
		if number%divisor == 0 {
			res := &calculatorpb.PrimeNumberDecompositionResponse{Result: divisor}
			stream.Send(res)
			number = number / divisor
		} else {
			divisor++
		}
	}
	return nil
}

func (*server) ComputeAverage(stream calculatorpb.CalculatorService_ComputeAverageServer) error {
	fmt.Printf("LongGreet function was invoked with a streaming request\n")
	result := float32(0.0)
	i := float32(0.0)

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			stream.SendAndClose(&calculatorpb.ComputeAverageResponse{
				Result: result / i,
			})
			return nil
		}
		if err != nil {
			log.Fatalf("Error while reading client stream %v", err)
		}
		i++
		value := req.GetValue()
		result += value
	}
}

func (*server) FindMaximum(stream calculatorpb.CalculatorService_FindMaximumServer) error {
	fmt.Printf("FindMaximum function was invoked with a streaming request\n")
	maximum := int32(0)
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalf("Error while reading client stream: %v", err)
		}
		number := req.GetValue()
		if number > maximum {
			maximum = number
			sendErr := stream.Send(&calculatorpb.FindMaximumResponse{
				Result: number,
			})
			if sendErr != nil {
				return err
			}
		}
	}

}

func (*server) SquareRoot(ctx context.Context, req *calculatorpb.SquareRootRequest) (*calculatorpb.SquareRootResponse, error) {
	number := req.GetNumber()
	if number < 0 {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Received a negative nubmer: %v", number),
		)
	}
	return &calculatorpb.SquareRootResponse{NumberRoot: math.Sqrt(float64(number))}, nil
}

func main() {
	fmt.Println("Welcome to calculator server")
	lis, err := net.Listen("tcp", "0.0.0.0:50053")
	if err != nil {
		log.Fatalf("error running server: %v", err)
	}

	s := grpc.NewServer()

	calculatorpb.RegisterCalculatorServiceServer(s, &server{})

	// Register refletion service on gRPC server
	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to server: %v", err)
	}
}
