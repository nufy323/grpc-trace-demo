package utrace

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/uptrace/uptrace-go/uptrace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

func PrintTraceID(ctx context.Context) {
	fmt.Println("trace:", TraceURL(trace.SpanFromContext(ctx)))
}

func TraceURL(span trace.Span) string {
	switch {
	case os.Getenv("UPTRACE_DSN") != "":
		return uptrace.TraceURL(span)
	case os.Getenv("OTEL_EXPORTER_JAEGER_ENDPOINT") != "":
		return fmt.Sprintf("http://localhost:16686/trace/%s", span.SpanContext().TraceID())
	default:
		return fmt.Sprintf("http://localhost:16686/trace/%s", span.SpanContext().TraceID())
	}
}

const defaultCollectorEndpoint = "localhost:4317"

func InitTracer(ctx context.Context, svcName string) func() {
	return configureOpentelemetry(ctx, svcName)
}

func configureOpentelemetry(ctx context.Context, svcName string) func() {
	jaegerCollectorEndpoint := os.Getenv("OTEL_EXPORTER_JAEGER_ENDPOINT")
	otelExporterType := os.Getenv("OTEL_EXPORTER_TYPE")
	switch otelExporterType {
	case "jaeger":
		return configureCommonExporter(ctx, svcName, jaegerCollectorEndpoint)
	case "stdout":
		return configureStdout(ctx)
	default:
		return configureCommonExporter(ctx, svcName, jaegerCollectorEndpoint)
	}
}

//export spans to jaeger collecor
func configureCommonExporter(ctx context.Context, serviceName string, endpoint string) func() {
	fmt.Println("use common exporter without jaeger")

	r, err := resource.New(ctx, []resource.Option{
		//设置服务名
		resource.WithAttributes(attribute.KeyValue{
			Key: "service.name", Value: attribute.StringValue(serviceName),
		}),
	}...)
	if err != nil {
		panic(err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(r),
	)

	//set global tracer provider
	otel.SetTracerProvider(tp)
	// propagate context
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	//set log exporter
	logExporter := NewLogExporter()
	bsp := sdktrace.NewBatchSpanProcessor(logExporter)
	tp.RegisterSpanProcessor(bsp)

	endpoint = defaultCollectorEndpoint
	// if endpoont is not empty, set grpc exporter
	if endpoint != "" {
		opts := []otlptracegrpc.Option{
			otlptracegrpc.WithTimeout(5 * time.Second),
			otlptracegrpc.WithRetry(otlptracegrpc.RetryConfig{}),
			otlptracegrpc.WithEndpoint(endpoint),
			otlptracegrpc.WithInsecure(),
		}

		grpcExporter, err := otlptracegrpc.New(ctx, opts...)
		if err != nil {
			panic(err)
		}

		bsp := sdktrace.NewBatchSpanProcessor(grpcExporter)
		tp.RegisterSpanProcessor(bsp)
	}

	return func() {
		if err := tp.Shutdown(ctx); err != nil {
			panic(err)
		}
	}
}

// export spans to stdout
func configureStdout(ctx context.Context) func() {
	fmt.Println("use stdout exporter without jaeger")
	provider := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	exp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		panic(err)
	}

	bsp := sdktrace.NewSimpleSpanProcessor(exp)
	provider.RegisterSpanProcessor(bsp)

	return func() {
		if err := provider.Shutdown(ctx); err != nil {
			panic(err)
		}
	}
}
