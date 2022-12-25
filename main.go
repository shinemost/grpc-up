package main

import (
	"log"
	"net"

	"github.com/shinemost/grpc-up/pbs"
	"github.com/shinemost/grpc-up/service"
	"github.com/shinemost/grpc-up/settings"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// settings.InitConfigs()
	lis, err := net.Listen("tcp", settings.Cfg.Address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pbs.RegisterProductInfoServer(s, &service.Server{})

	//服务器反射方法，客户端可以获取到server元数据
	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
