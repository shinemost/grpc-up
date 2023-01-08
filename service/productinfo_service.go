package service

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"go.opencensus.io/trace"

	uuid "github.com/satori/go.uuid"
	pb "github.com/shinemost/grpc-up/pbs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	CustomizedCounterMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "product_mgt_server_handle_count",
		Help: "Total number of RPCs handled on the server.",
	}, []string{"name"})
)

type Server struct {
	pb.UnimplementedProductInfoServer
	productMap map[string]*pb.Product
}

func (s *Server) AddProduct(ctx context.Context, in *pb.Product) (*pb.ProductID, error) {
	ctx, span := trace.StartSpan(ctx, "product.AddProduct")
	defer span.End()
	CustomizedCounterMetric.WithLabelValues(in.Name).Inc()
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
