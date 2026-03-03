package main

import (
	"context"
	"log"
	"net"

	pb "github.com/aberyotaro/grpc-sample/gen/count"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedCountServiceServer
}

func (s *server) Count(_ context.Context, req *pb.CountRequest) (*pb.CountResponse, error) {
	n := int32(len([]rune(req.Text)))
	log.Printf("Count called: %q -> %d", req.Text, n)
	return &pb.CountResponse{Count: n}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterCountServiceServer(s, &server{})
	log.Println("count service listening on :50053")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
