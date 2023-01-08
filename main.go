package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/opentracing/opentracing-go"
	"github.com/shinemost/grpc-up/pbs"
	"github.com/shinemost/grpc-up/service"
	"github.com/shinemost/grpc-up/settings"
	"github.com/shinemost/grpc-up/tracer"
	"go.opencensus.io/plugin/ocgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
	// Start z-Pages server.
	// go func() {
	// 	mux := http.NewServeMux()
	// 	zpages.Handle(mux, "/debug")
	// 	addr := "127.0.0.1:8888"
	// 	if err := http.ListenAndServe(addr, mux); err != nil {
	// 		log.Fatalf("Failed to serve zPages")
	// 	}
	// }()

	settings.InitConfigs()

	//启动http Server
	go runGatewayServer()

	// Create a HTTP server for prometheus.
	// httpServer := &http.Server{Handler: promhttp.HandlerFor(reg, promhttp.HandlerOpts{}), Addr: settings.Cfg.Prometheus.Address}

	// Register stats and trace exporters to export
	// the collected data.
	// view.RegisterExporter(&exporter.PrintExporter{})

	// Register the views to collect server request count.
	// if err := view.Register(ocgrpc.DefaultServerViews...); err != nil {
	// 	log.Fatal(err)
	// }

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

	// initialize jaegertracer
	jaegertracer, closer, err := tracer.NewTracer("grpc-server")
	if err != nil {
		log.Fatalln(err)
	}
	defer closer.Close()

	opentracing.SetGlobalTracer(jaegertracer)

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
		grpc.UnaryInterceptor(grpc_opentracing.UnaryServerInterceptor(grpc_opentracing.WithTracer(jaegertracer))),
	)

	//RPC服务端多路复用，一个RPCserver注册多个服务
	pbs.RegisterProductInfoServer(s, &service.Server{})
	pbs.RegisterOrderManagementServer(s, &service.OrderServer{})

	// Initialize all metrics.
	// grpcMetrics.InitializeMetrics(s)
	// Start your http server for prometheus.
	// go func() {
	// 	if err := httpServer.ListenAndServe(); err != nil {
	// 		log.Fatal("Unable to start a http server.")
	// 	}
	// }()

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

//	func main() {
//		mux := http.NewServeMux()
//		zpages.Handle(mux, "/debug")
//		addr := ":8888"
//		if err := http.ListenAndServe(addr, mux); err != nil {
//			log.Fatalf("Failed to serve zPages")
//		}
//	}
// var (
// 	// Create a metrics registry.
// 	reg = prometheus.NewRegistry()

// 	// Create some standard server metrics.
// 	grpcMetrics = grpc_prometheus.NewServerMetrics()
// )

// func init() {
// 	// Register standard server metrics and customized metrics to registry.
// 	reg.MustRegister(grpcMetrics, service.CustomizedCounterMetric)
// }

// func initTracing() {
// 	// This is a demo app with low QPS. trace.AlwaysSample() is used here
// 	// to make sure traces are available for observation and analysis.
// 	// In a production environment or high QPS setup please use
// 	// trace.ProbabilitySampler set at the desired probability.
// 	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
// 	agentEndpointURI := "localhost:6831"
// 	collectorEndpointURI := "http://localhost:14268/api/traces"
// 	exporter, err := jaeger.NewExporter(jaeger.Options{
// 		CollectorEndpoint: collectorEndpointURI,
// 		AgentEndpoint:     agentEndpointURI,
// 		ServiceName:       "grpc-server",
// 	})
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	trace.RegisterExporter(exporter)

// }

func runGatewayServer() {

	serverMuxOption := runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	})
	grpcMux := runtime.NewServeMux(serverMuxOption)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := pbs.RegisterProductInfoHandlerServer(ctx, grpcMux, &service.Server{})
	if err != nil {
		log.Fatalf("Fail to register gRPC service endpoint: %v", err)
		return
	}
	log.Printf("start HTTP server at %s", settings.Cfg.Web.Port)

	if err := http.ListenAndServe(settings.Cfg.Web.Port, grpcMux); err != nil {
		log.Fatalf("Could not setup HTTP endpoint: %v", err)
	}

}
