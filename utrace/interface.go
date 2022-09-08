package utrace

import (
	"context"

	"github.com/nufy323/grpc-demo/utrace/util"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/codes"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

//start new span
func StartSpan(ctx context.Context, subSys string, operationName string, attributes map[string]interface{}) context.Context {
	ctx, span := otel.Tracer("").Start(ctx, operationName, trace.WithSpanKind(trace.SpanKindInternal))

	attrKeyValues := convertKvsToFields(attributes)
	span.SetAttributes(attrKeyValues...)

	return ctx
}

//finish span
func FinishSpan(ctx context.Context) {
	span := trace.SpanFromContext(ctx)
	if span == nil || !span.IsRecording() {
		return
	}
	span.End()
}

func convertKvsToFields(kvs map[string]interface{}) []attribute.KeyValue {
	attrKeyValues := make([]attribute.KeyValue, 0, len(kvs))
	for k, v := range kvs {
		util.Attribute(k, v)
		attrKeyValues = append(attrKeyValues, util.Attribute(k, v))
	}
	return attrKeyValues
}

//add attribute to span
func AddSpanAttribute(ctx context.Context, key string, value interface{}) {
	span := trace.SpanFromContext(ctx)
	if span == nil || !span.IsRecording() {
		return
	}
	span.SetAttributes(util.Attribute(key, value))
}

type SpanStatusCode int

const (
	Error SpanStatusCode = 1
	OK    SpanStatusCode = 2
)

func SetSpanCode(ctx context.Context, code SpanStatusCode, desc string) {
	span := trace.SpanFromContext(ctx)
	if span == nil || !span.IsRecording() {
		return
	}
	span.SetStatus(codes.Code(code), desc)
}

//add event to span
func AddSpanEvent(ctx context.Context, eventName string, key string, value interface{}) {
	span := trace.SpanFromContext(ctx)
	if span == nil || !span.IsRecording() {
		return
	}

	attr := util.Attribute(key, value)
	span.AddEvent(eventName, trace.WithAttributes(attr))
}

//identifies a Span as the root Span for a new trace
func AsyncSpan(ctx context.Context, subSys string, operationName string, attributes map[string]interface{}) context.Context {
	span := trace.SpanFromContext(ctx)
	link := trace.Link{}
	if span != nil || !span.IsRecording() {
		link = trace.LinkFromContext(ctx)
	}
	ctx, _ = otel.Tracer("").Start(ctx, operationName,
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithNewRoot(),
		trace.WithLinks(link),
	)

	return ctx
}

func TraceLog(ctx context.Context) *log.Entry {
	return log.WithContext(ctx)
}
