package main

import (
	"context"
	"log"
	"net"
	"strings"

	pb "github.com/aberyotaro/grpc-sample/gen/uppercase"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedUppercaseServiceServer
}

func (s *server) ToUpper(_ context.Context, req *pb.UppercaseRequest) (*pb.UppercaseResponse, error) {
	log.Printf("ToUpper called: %q", req.Text)
	return &pb.UppercaseResponse{Text: strings.ToUpper(req.Text)}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterUppercaseServiceServer(s, &server{})
	log.Println("uppercase service listening on :50052")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
