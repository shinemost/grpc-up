package service

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/golang/protobuf/ptypes/wrappers"
	pb "github.com/shinemost/grpc-up/pbs"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type OrderServer struct {
	pb.UnimplementedOrderManagementServer
	orderMap map[string]*pb.Order
}

func (s *OrderServer) AddOrder(ctx context.Context, orderReq *pb.Order) (*wrappers.StringValue, error) {

	log.Printf("Order Added. ID : %v", orderReq.Id)

	if s.orderMap == nil {
		s.orderMap = make(map[string]*pb.Order)
	}
	s.orderMap[orderReq.Id] = orderReq
	return &wrapperspb.StringValue{Value: "Order Added: " + orderReq.Id}, nil
}

func (s *OrderServer) SearchOrders(searchQuery *wrappers.StringValue, stream pb.OrderManagement_SearchOrdersServer) error {

	for key, order := range s.orderMap {
		log.Print(key, order)
		for _, itemStr := range order.Items {
			log.Print(itemStr)
			if strings.Contains(itemStr, searchQuery.Value) {
				err := stream.Send(order)
				if err != nil {
					return fmt.Errorf("error sending message to stream:%v", err)
				}
				log.Print("Matching Order Found:" + key)
				break
			}
		}
	}
	return nil
}
