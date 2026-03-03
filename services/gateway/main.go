package main

import (
	"context"
	"log"
	"net"

	countpb "github.com/aberyotaro/grpc-sample/gen/count"
	gatewaypb "github.com/aberyotaro/grpc-sample/gen/gateway"
	uppercasepb "github.com/aberyotaro/grpc-sample/gen/uppercase"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type server struct {
	gatewaypb.UnimplementedGatewayServiceServer
	uppercaseClient uppercasepb.UppercaseServiceClient
	countClient     countpb.CountServiceClient
}

func (s *server) Process(ctx context.Context, req *gatewaypb.ProcessRequest) (*gatewaypb.ProcessResponse, error) {
	log.Printf("Process called: %q", req.Text)

	upperRes, err := s.uppercaseClient.ToUpper(ctx, &uppercasepb.UppercaseRequest{Text: req.Text})
	if err != nil {
		return nil, err
	}

	countRes, err := s.countClient.Count(ctx, &countpb.CountRequest{Text: req.Text})
	if err != nil {
		return nil, err
	}

	return &gatewaypb.ProcessResponse{
		Uppercase: upperRes.Text,
		Count:     countRes.Count,
	}, nil
}

func main() {
	uppercaseConn, err := grpc.NewClient("uppercase:50052",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to uppercase: %v", err)
	}
	defer uppercaseConn.Close()

	countConn, err := grpc.NewClient("count:50053",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to count: %v", err)
	}
	defer countConn.Close()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	gatewaypb.RegisterGatewayServiceServer(s, &server{
		uppercaseClient: uppercasepb.NewUppercaseServiceClient(uppercaseConn),
		countClient:     countpb.NewCountServiceClient(countConn),
	})
	log.Println("gateway service listening on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
