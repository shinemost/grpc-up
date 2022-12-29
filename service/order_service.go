package service

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes/wrappers"
	pb "github.com/shinemost/grpc-up/pbs"
	epb "google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"io"
	"log"
	"strings"
	"time"
)

const orderBatchSize = 3

var orderMap = make(map[string]*pb.Order)

type OrderServer struct {
	pb.UnimplementedOrderManagementServer
	orderMap map[string]*pb.Order
}

func (s *OrderServer) AddOrder(ctx context.Context, orderReq *pb.Order) (*wrappers.StringValue, error) {

	if orderReq.Id == "-1" {
		errorStatus := status.New(codes.InvalidArgument, "非法ID参数")
		ds, err := errorStatus.WithDetails(&epb.BadRequest_FieldViolation{
			Field:       "ID",
			Description: fmt.Sprintf("Order ID 是非法的 %s : %s", orderReq.Id, orderReq.Description),
		})
		if err != nil {
			return nil, errorStatus.Err()
		}
		return nil, ds.Err()
	}

	orderMap[orderReq.Id] = orderReq

	// ***** Reading Metadata from Client *****
	md, metadataAvailable := metadata.FromIncomingContext(ctx)
	if !metadataAvailable {
		return nil, status.Errorf(codes.DataLoss, "error: failed to get metadata")
	}
	if t, ok := md["timestamp"]; ok {
		fmt.Printf("timestamp from metadata:\n")
		for i, e := range t {
			fmt.Printf("====> Metadata %d. %s\n", i, e)
		}
	}

	// Creating and sending a header.
	header := metadata.New(map[string]string{"location": "San Jose", "timestamp": time.Now().Format(time.StampNano)})
	_ = grpc.SendHeader(ctx, header)

	//sleepDuration := 5
	//log.Println("sleeping for ", sleepDuration, " s")
	//time.Sleep(time.Duration(sleepDuration) * time.Second)

	//服务端判断是否超时错误
	//if ctx.Err() == context.DeadlineExceeded {
	//	log.Printf("RPC has reached deadline exceeded state: %s", ctx.Err())
	//	return nil, ctx.Err()
	//}

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
		// You can determine whether the current RPC is cancelled by the other party.
		if stream.Context().Err() == context.Canceled {
			log.Printf(" Context Cacelled for this stream: -> %s", stream.Context().Err())
			log.Printf("Stopped update any more order of this stream!")
			return stream.Context().Err()
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
