package main

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/nufy323/grpc-demo/example/api"
	"github.com/nufy323/grpc-demo/utrace"
	"github.com/nufy323/grpc-demo/utrace/logger"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

const addr = ":9999"

type helloServer struct {
	api.HelloServiceServer
}

func (s *helloServer) SayHello(ctx context.Context, in *api.HelloRequest) (*api.HelloResponse, error) {

	utrace.StartSpan(ctx, "", "server-hello", nil)
	defer utrace.FinishSpan(ctx)

	DoSomeThing(ctx)
	utrace.TraceLog(ctx).Infoln("saying Hello")

	go func(ctx context.Context) {
		ctx = utrace.AsyncSpan(ctx, "", "ticker", nil)
		defer utrace.FinishSpan(ctx)
		time.Sleep(3 * time.Second)
		utrace.TraceLog(ctx).Infoln("async finish")
	}(ctx)
	utrace.SetSpanStatus(ctx, utrace.OK, "")
	//do other things
	time.Sleep(3 * time.Second)
	return &api.HelloResponse{Reply: "Hello " + in.Greeting}, nil
}

func DoSomeThing(ctx context.Context) {
	ctx = utrace.StartSpan(ctx, "", "doSomeThing", nil)
	defer utrace.FinishSpan(ctx)

	//do something
	time.Sleep(2 * time.Second)
	utrace.TraceLog(ctx).Infoln("Do something")
}

func main() {

	shutdown := utrace.InitTracer(context.TODO(), "agent")
	defer shutdown()

	logger.InitLogger(logger.WithLoggerLevel("trace"), logger.WithPrettyPrint(true))

	log.Println("serving on", addr)

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
		return
	}

	server := grpc.NewServer(
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
	)

	api.RegisterHelloServiceServer(server, &helloServer{})
	if err := server.Serve(ln); err != nil {
		log.Fatal(err)
		return
	}
}
