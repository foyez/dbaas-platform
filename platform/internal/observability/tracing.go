package observability

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// SetupTracing configures OpenTelemetry tracing from the standard OTEL_*
// environment variables (OTEL_EXPORTER_OTLP_ENDPOINT, OTEL_SERVICE_NAME,
// OTEL_TRACES_SAMPLER, ...) and returns a function that flushes pending
// spans on shutdown.
//
// Without OTEL_EXPORTER_OTLP_ENDPOINT tracing stays off: spans are still
// created and propagated, but nothing is exported.
func SetupTracing(ctx context.Context) (func(context.Context) error, error) {
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") == "" {
		return func(context.Context) error { return nil }, nil
	}

	exporter, err := otlptracegrpc.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("create OTLP trace exporter: %w", err)
	}

	// Default + env: service.name from OTEL_SERVICE_NAME, plus
	// OTEL_RESOURCE_ATTRIBUTES and SDK details.
	res, err := resource.Merge(resource.Default(), resource.Environment())
	if err != nil {
		return nil, fmt.Errorf("build trace resource: %w", err)
	}

	// The sampler comes from OTEL_TRACES_SAMPLER (default: parent-based, always on).
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(provider)

	return provider.Shutdown, nil
}
