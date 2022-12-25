package service

import (
	"context"

	uuid "github.com/satori/go.uuid"
	pb "github.com/shinemost/grpc-up/pbs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedProductInfoServer
	productMap map[string]*pb.Product
}

func (s *Server) AddProduct(ctx context.Context, in *pb.Product) (*pb.ProductID, error) {
	out := uuid.NewV4()
	in.Id = out.String()
	if s.productMap == nil {
		s.productMap = make(map[string]*pb.Product)
	}
	s.productMap[in.Id] = in

	return &pb.ProductID{Value: in.Id}, status.New(codes.OK, "").Err()
}
func (s *Server) GetProduct(ctx context.Context, in *pb.ProductID) (*pb.Product, error) {
	value, exists := s.productMap[in.Value]
	if exists {
		return value, status.New(codes.OK, "").Err()
	}

	return nil, status.Errorf(codes.NotFound, "Product does not exists", in.Value)
}
