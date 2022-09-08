package utrace

import (
	"context"

	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/sdk/trace"
)

//log exporter export spans to log file
func NewLogExporter(logger *log.Logger) *logExporter {
	return &logExporter{
		exportLogger: logger,
	}
}

type logExporter struct {
	exportLogger *log.Logger
}

func (le *logExporter) ExportSpans(ctx context.Context, spans []trace.ReadOnlySpan) error {
	for _, span := range spans {
		le.logSpan(span)
	}
	return nil
}

func (le *logExporter) Shutdown(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return nil
}

func (le *logExporter) logSpan(roSpan trace.ReadOnlySpan) {
	le.exportLogger.WithFields(log.Fields{
		"traceID":   roSpan.SpanContext().TraceID(),
		"spanID":    roSpan.SpanContext().SpanID(),
		"pSpanID":   roSpan.Parent().SpanID(),
		"attr":      roSpan.Attributes(),
		"opName":    roSpan.Name(),
		"links":     roSpan.Links(),
		"startTime": roSpan.StartTime(),
		"endTime":   roSpan.EndTime(),
		"events":    roSpan.Events(),
		"logType":   "span",
	}).Trace("")
}
