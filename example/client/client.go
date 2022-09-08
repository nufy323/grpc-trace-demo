package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/nufy323/grpc-demo/example/api"
	"github.com/nufy323/grpc-demo/utrace"
	"github.com/nufy323/grpc-demo/utrace/logger"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func main() {

	ctx := context.Background()

	//init logger
	logger.InitLogger(logger.WithPrettyPrint(true), logger.WithLoggerLevel("trace"))

	//init tracer
	shutdown := utrace.InitTracer(context.Background(), "manager")
	defer shutdown()

	ctx = utrace.StartSpan(ctx, "", "begin", nil)
	defer utrace.FinishSpan(ctx)

	target := os.Getenv("GRPC_TARGET")
	if target == "" {
		target = ":9999"
	}

	utrace.TraceLog(ctx).Infof("connecting to: %s", target)

	conn, err := grpc.Dial(target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	)
	if err != nil {
		log.Fatal(err)
		return
	}

	defer func() {
		_ = conn.Close()
	}()

	client := api.NewHelloServiceClient(conn)
	if err := sayHello(ctx, client); err != nil {
		log.Fatal(err)
		return
	}

	utrace.TraceLog(ctx).Infoln("after say hello success")

}

func sayHello(ctx context.Context, client api.HelloServiceClient) error {
	ctx = utrace.StartSpan(ctx, "", "sayHello-test", map[string]interface{}{
		"groupId":    "mysql-test",
		"instanceId": "mysql-123fdf",
	})
	defer utrace.FinishSpan(ctx)

	utrace.TraceLog(ctx).Infoln("calling say hello")

	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(
		"timestamp", time.Now().Format(time.StampNano),
		"client-id", "web-api-client",
		"user-id", "test-user",
	))

	resp, err := client.SayHello(ctx, &api.HelloRequest{Greeting: "World"})
	if err != nil {
		return err
	}
	utrace.TraceLog(ctx).Infoln("reply:", resp.Reply)

	return nil
}
