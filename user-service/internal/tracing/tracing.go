package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"github.com/rs/zerolog/log"
)

// InitTracing initializes OpenTelemetry tracing with Jaeger
func InitTracing(serviceName, jaegerEndpoint string) func() {
	// Create Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerEndpoint)))
	if err != nil {
		log.Error().Err(err).Msg("failed to create jaeger exporter")
		return func() {}
	}

	// Create resource
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion("1.0.0"),
		),
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to create resource")
		return func() {}
	}

	// Create trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	// Set global trace provider
	otel.SetTracerProvider(tp)

	log.Info().Str("service", serviceName).Str("jaeger_endpoint", jaegerEndpoint).Msg("tracing initialized")

	// Return cleanup function
	return func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Error().Err(err).Msg("failed to shutdown trace provider")
		}
	}
}

// GetTracer returns a tracer for the user-service
func GetTracer() trace.Tracer {
	return otel.Tracer("user-service")
}

// StartSpan starts a new span with the given name
func StartSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	return GetTracer().Start(ctx, name)
}

// AddSpanError adds an error to the current span
func AddSpanError(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}
