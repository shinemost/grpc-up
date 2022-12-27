package interceptor

import (
	"context"
	"google.golang.org/grpc"
	"log"
)

func OrderUnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {
	log.Println("===========[Server Interceptor] ", info.FullMethod)

	m, err := handler(ctx, req)

	//处理后置逻辑
	log.Printf(" Post Proc message : %s", m)

	return m, err

}
