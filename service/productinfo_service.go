package service

import (
	"context"

	pb "github.com/shinemost/grpc-up/pbs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	productMap map[string]*pb.Product
}

func (server *server) AddProduct(context.Context, *pb.Product) (*pb.ProductID, error) {

	return nil, status.Errorf(codes.Unimplemented, "method AddProduct not implemented")
}
func (server *server) GetProduct(context.Context, *pb.ProductID) (*pb.Product, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetProduct not implemented")
}
