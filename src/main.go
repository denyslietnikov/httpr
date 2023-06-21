package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/hirosassa/zerodriver"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

var (
	// MetricsHost exporter host:port
	MetricsHost = os.Getenv("METRICS_HOST")
	appVersion  = os.Getenv("VERSION")
	logger      zerodriver.Logger
)

type RequestCounter struct {
	counts map[string]int
	mu     sync.Mutex
}

func (rc *RequestCounter) Increment(port string) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.counts[port]++
}

func main() {
	router1 := http.NewServeMux()
	router2 := http.NewServeMux()
	logger := zerodriver.NewProductionLogger()

	counter := &RequestCounter{
		counts: make(map[string]int),
	}

	// Routing HTTP requests for Port 8181
	router1.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		counter.Increment(":8181")
		count := counter.counts[":8181"]
		fmt.Fprintln(w, "Hello from Port 8181!")
		fmt.Fprintf(w, "Total requests on Port 8181: %d\n", count)

		tracer := otel.GetTracerProvider().Tracer("httpr_tracer")

		// Extract trace context from the incoming request
		ctx := r.Context()

		// Start a new span with the extracted context
		_, span := tracer.Start(ctx, "httpr_span_:8181")
		defer span.End()

		// Add custom attributes to the span
		span.SetAttributes(semconv.HTTPMethodKey.String(r.Method))
		span.SetAttributes(semconv.HTTPURLKey.String(r.URL.String()))

		// Example of accessing trace ID from the span's SpanContext
		traceID := span.SpanContext().TraceID().String()
		fmt.Fprintf(w, "Trace ID: %s\n", traceID)

		// Record metrics
		metrics(ctx, ":8181", count, traceID)
	})

	// Routing HTTP requests for Port 8282
	router2.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		counter.Increment(":8282")
		count := counter.counts[":8282"]
		fmt.Fprintln(w, "Hello from Port 8282!")
		fmt.Fprintf(w, "Total requests on Port 8282: %d\n", count)

		tracer := otel.GetTracerProvider().Tracer("httpr_tracer")

		// Extract trace context from the incoming request
		ctx := r.Context()

		// Start a new span with the extracted context
		_, span := tracer.Start(ctx, "httpr_span_:8282")
		defer span.End()

		// Add custom attributes to the span
		span.SetAttributes(semconv.HTTPMethodKey.String(r.Method))
		span.SetAttributes(semconv.HTTPURLKey.String(r.URL.String()))

		// Example of accessing trace ID from the span's SpanContext
		traceID := span.SpanContext().TraceID().String()
		fmt.Fprintf(w, "Trace ID: %s\n", traceID)

		// Record metrics
		metrics(ctx, ":8282", count, traceID)
	})

	go startHTTPServer(":8181", router1)
	go startHTTPServer(":8282", router2)

	initMetrics(context.Background())
	initTracing(context.Background())

	logger.Info().Str("HTTPR", appVersion).Msg("HTTP server listening on ports :8181 and :8282")

	select {}
}

func startHTTPServer(port string, handler http.Handler) {
	err := http.ListenAndServe(port, handler)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to start HTTP server")
	}
}

// Initialize OpenTelemetry (metrics)
func initMetrics(ctx context.Context) {
	// Create a new OTLP Metric gRPC exporter with the specified endpoint and options
	exporter, _ := otlpmetricgrpc.New(
		ctx,
		otlpmetricgrpc.WithEndpoint(MetricsHost),
		otlpmetricgrpc.WithInsecure(),
	)

	// Define the resource with attributes that are common to all metrics.
	// labels/tags/resources that are common to all metrics.
	resource := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(fmt.Sprintf("httpr_%s", appVersion)),
	)

	// Create a new MeterProvider with the specified resource and reader
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(resource),
		sdkmetric.WithReader(
			// collects and exports metric data every 30 seconds.
			sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(30*time.Second)),
		),
	)

	// Set the global MeterProvider to the newly created MeterProvider
	otel.SetMeterProvider(mp)
}

// Initialize OpenTelemetry (trace)
func initTracing(ctx context.Context) {

	// Create a new OTLP Trace gRPC exporter with the specified endpoint and options
	exporter, _ := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithEndpoint(MetricsHost),
		otlptracegrpc.WithInsecure(),
	)
	resource := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(fmt.Sprintf("httpr_%s", appVersion)),
	)

	// Create a new TracerProvider with the specified exporter
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource),
	)

	// Set the global TracerProvider to the newly created TracerProvider
	otel.SetTracerProvider(tp)
}

func metrics(ctx context.Context, router string, count int, traceID string) {
	// Get the global MeterProvider and create a new Meter with the name "router"
	meter := otel.GetMeterProvider().Meter("httpr_meter")

	// Get or create an Int64Counter instrument with the name "httpr_<counter>"
	counter, _ := meter.Int64Counter(fmt.Sprintf("httpr_counter_%s", router))

	// Add a value of 1 to the Int64Counter
	counter.Add(ctx, 1)

	// Log the metrics and traceID using log.Println
	log.Println("Metrics", "Router", router, "Requests", count, "TraceID", traceID)
}

func init() {
	ctx := context.Background()
	initMetrics(ctx)
}
