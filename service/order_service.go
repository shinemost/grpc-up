package service

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes/wrappers"
	pb "github.com/shinemost/grpc-up/pbs"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const orderBatchSize = 3

var orderMap = make(map[string]*pb.Order)

type OrderServer struct {
	pb.UnimplementedOrderManagementServer
	orderMap map[string]*pb.Order
}

func (s *OrderServer) AddOrder(ctx context.Context, orderReq *pb.Order) (*wrappers.StringValue, error) {

	orderMap[orderReq.Id] = orderReq

	sleepDuration := 5
	log.Println("sleeping for ", sleepDuration, " s")
	time.Sleep(time.Duration(sleepDuration) * time.Second)

	//服务端判断是否超时错误
	if ctx.Err() == context.DeadlineExceeded {
		log.Printf("RPC has reached deadline exceeded state: %s", ctx.Err())
		return nil, ctx.Err()
	}
	log.Printf("Order Added. ID : %v", orderReq.Id)

	return &wrapperspb.StringValue{Value: "Order Added: " + orderReq.Id}, nil
}

func (s *OrderServer) SearchOrders(searchQuery *wrappers.StringValue, stream pb.OrderManagement_SearchOrdersServer) error {

	for key, order := range orderMap {
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

func (s *OrderServer) UpdateOrders(stream pb.OrderManagement_UpdateOrdersServer) error {
	ordersStr := "Updated Order IDs:"
	for {
		order, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&wrapperspb.StringValue{Value: "Orders processed" + ordersStr})
		}
		orderMap[order.Id] = order
		log.Println("Order ID", order.Id, ":Updated")
		ordersStr += order.Id + ","
	}

}

func (s *OrderServer) ProcessOrders(stream pb.OrderManagement_ProcessOrdersServer) error {
	//stream即可以读也可以写
	batchMarker := 1
	var combinedShipmentMap = make(map[string]pb.CombinedShipment)
	for {
		orderId, err := stream.Recv()
		log.Printf("Reading Proc order : %s", orderId)
		if err == io.EOF {
			// Client has sent all the messages
			// Send remaining shipments
			log.Printf("EOF : %s", orderId)
			for _, shipment := range combinedShipmentMap {
				if err := stream.Send(&shipment); err != nil {
					return err
				}
			}
			return nil
		}
		if err != nil {
			log.Println(err)
			return err
		}

		destination := orderMap[orderId.GetValue()].Destination
		shipment, found := combinedShipmentMap[destination]

		if found {
			ord := orderMap[orderId.GetValue()]
			shipment.OrdersList = append(shipment.OrdersList, ord)
			combinedShipmentMap[destination] = shipment
		} else {
			comShip := pb.CombinedShipment{Id: "cmb - " + (orderMap[orderId.GetValue()].Destination), Status: "Processed!"}
			ord := orderMap[orderId.GetValue()]
			comShip.OrdersList = append(shipment.OrdersList, ord)
			combinedShipmentMap[destination] = comShip
			log.Print(len(comShip.OrdersList), comShip.GetId())
		}

		//达到3个时批量发送
		if batchMarker == orderBatchSize {
			for _, comb := range combinedShipmentMap {
				log.Printf("Shipping : %v -> %v", comb.Id, len(comb.OrdersList))
				if err := stream.Send(&comb); err != nil {
					return err
				}
			}
			batchMarker = 0
			combinedShipmentMap = make(map[string]pb.CombinedShipment)
		} else {
			batchMarker++
		}
	}
}
