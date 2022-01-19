package main

import (
	"context"
	"fmt"
	"log"

	"go.opentelemetry.io/otel/api/core"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/key"
	"go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/exporter/trace/jaeger"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func initTracer(name string) func() {
	// Create Jaeger Exporter
	exporter, err := jaeger.NewExporter(
		jaeger.WithCollectorEndpoint("http://localhost:14268/api/traces"),
		jaeger.WithProcess(jaeger.Process{
			ServiceName: name,
			Tags: []core.KeyValue{
				key.String("exporter", "jaeger"),
				key.Float64("float", 312.23),
			},
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	// For demoing purposes, always sample. In a production application, you should
	// configure this to a trace.ProbabilitySampler set at the desired
	// probability.
	tp, err := sdktrace.NewProvider(
		sdktrace.WithConfig(sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
		sdktrace.WithSyncer(exporter))
	if err != nil {
		log.Fatal(err)
	}
	global.SetTraceProvider(tp)
	return func() {
		exporter.Flush()
	}
}

func main() {
	fn := initTracer("trace-main")
	defer fn()

	ctx := context.Background()

	tr := global.TraceProvider().Tracer("main-foo")
	ctx, span := tr.Start(ctx, "foo")
	sub(ctx)
	span.End()
}

func sub(ctx context.Context) {
	tr := global.TraceProvider().Tracer("sub-bar")
	_, span := tr.Start(ctx, "bar")
	defer span.End()

	// Call service2
	fmt.Printf("%016x\n", span.SpanContext().TraceID)
	fmt.Printf("%016x\n", span.SpanContext().SpanID)
	service(fmt.Sprintf("%016x", span.SpanContext().TraceID), fmt.Sprintf("%016x", span.SpanContext().SpanID))
}

func service(traceID string, spanID string) {
	fn := initTracer("trace-service")
	defer fn()

	tid, _ := core.TraceIDFromHex(traceID)
	sid, _ := core.SpanIDFromHex(spanID)

	ctx := context.Background()
	tr := global.TraceProvider().Tracer("service-step01")
	sc := core.SpanContext{
		TraceID:    tid,
		SpanID:     sid,
		TraceFlags: 0x0,
	}
	_, span := tr.Start(trace.ContextWithRemoteSpanContext(ctx, sc), "service-span1")
	span.End()

	fmt.Printf("%016x\n", span.SpanContext().TraceID)
	fmt.Printf("%016x\n", span.SpanContext().SpanID)

	service2(fmt.Sprintf("%016x", span.SpanContext().TraceID), fmt.Sprintf("%016x", span.SpanContext().SpanID))
}

func service2(traceID string, spanID string) {
	fn := initTracer("trace-service2")
	defer fn()

	tid, _ := core.TraceIDFromHex(traceID)
	sid, _ := core.SpanIDFromHex(spanID)

	ctx := context.Background()
	tr := global.TraceProvider().Tracer("service2-step01")
	sc := core.SpanContext{
		TraceID:    tid,
		SpanID:     sid,
		TraceFlags: 0x0,
	}
	_, span := tr.Start(trace.ContextWithRemoteSpanContext(ctx, sc), "service2-span1")
	span.End()
	fmt.Printf("%016x\n", span.SpanContext().TraceID)
	fmt.Printf("%016x\n", span.SpanContext().SpanID)
}
