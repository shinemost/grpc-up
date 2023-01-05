package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"go.opencensus.io/examples/exporter"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/stats/view"
	"google.golang.org/grpc/credentials"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/shinemost/grpc-up/pbs"
	"github.com/shinemost/grpc-up/service"
	"github.com/shinemost/grpc-up/settings"
	"go.opencensus.io/zpages"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	settings.InitConfigs()

	// Start z-Pages server.
	go func() {
		mux := http.NewServeMux()
		zpages.Handle(mux, "/debug")
		log.Fatal(http.ListenAndServe("127.0.0.1:9999", mux))
	}()

	// Register stats and trace exporters to export
	// the collected data.
	view.RegisterExporter(&exporter.PrintExporter{})

	// Register the views to collect server request count.
	if err := view.Register(ocgrpc.DefaultServerViews...); err != nil {
		log.Fatal(err)
	}

	cert, err := tls.LoadX509KeyPair(settings.Cfg.CrtFile, settings.Cfg.KeyFile)

	if err != nil {
		log.Fatalf("failed to load x509 key pair : %s", err)
	}

	certPool := x509.NewCertPool()
	ca, err := os.ReadFile(settings.Cfg.CaFile)
	if err != nil {
		log.Fatalf("could not read ca certificate: %s", err)
	}

	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		log.Fatalf("failed to append ca cert")
	}

	s := grpc.NewServer(
		//grpc.UnaryInterceptor(interceptor.OrderUnaryServerInterceptor),
		//grpc.StreamInterceptor(interceptor.OrderServerStreamInterceptor),
		//grpc.UnaryInterceptor(interceptor.EnsureVaildBasicCredentials),
		//grpc.UnaryInterceptor(interceptor.EnsureVaildTokenCredentials),
		grpc.Creds(
			credentials.NewTLS(&tls.Config{
				ClientAuth:   tls.RequireAndVerifyClientCert,
				Certificates: []tls.Certificate{cert},
				ClientCAs:    certPool,
			}),
		),
		grpc.StatsHandler(&ocgrpc.ServerHandler{}),
	)

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

// func main() {
// 	var wg sync.WaitGroup
// 	for _, addr := range service.Addrs {
// 		wg.Add(1)
// 		go func(addr string) {
// 			defer wg.Done()
// 			service.StartServer(addr)
// 		}(addr)
// 	}
// 	wg.Wait()
// }
