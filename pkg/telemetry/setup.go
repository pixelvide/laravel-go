package telemetry

import (
	"context"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// InitTracer initializes an OpenTelemetry tracer provider with a stdout exporter.
func InitTracer(serviceName string) (*sdktrace.TracerProvider, error) {
	exporter, err := stdouttrace.New(
		stdouttrace.WithPrettyPrint(),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		)),
	)

	otel.SetTracerProvider(tp)
	return tp, nil
}

// LoggerFromContext retrieves the logger from the context.
// If no logger is found, it returns the global logger.
func LoggerFromContext(ctx context.Context) *zerolog.Logger {
	return zerolog.Ctx(ctx)
}

// SetGlobalLogger configures the global zerolog logger.
func SetGlobalLogger() {
	// Default to console pretty print for development
	// In production, this should likely be JSON, but sticking to the user's setup
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}
