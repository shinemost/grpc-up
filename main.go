package main

import (
	"fmt"
	"github.com/shinemost/grpc-up/interceptor"
	"log"
	"net"

	"github.com/shinemost/grpc-up/pbs"
	"github.com/shinemost/grpc-up/service"
	"github.com/shinemost/grpc-up/settings"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	settings.InitConfigs()

	s := grpc.NewServer(grpc.UnaryInterceptor(interceptor.OrderUnaryServerInterceptor),
		grpc.StreamInterceptor(interceptor.OrderServerStreamInterceptor))

	//RPC服务端多路复用，一个RPCserver注册多个服务
	pbs.RegisterProductInfoServer(s, &service.Server{})
	pbs.RegisterOrderManagementServer(s, &service.OrderServer{})

	//服务器反射方法，客户端可以获取到server元数据
	reflection.Register(s)
	fmt.Println(settings.Cfg.Grpc.Address)
	lis, err := net.Listen("tcp", settings.Cfg.Grpc.Address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Printf("start gRPC server at %s", lis.Addr().String())

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
