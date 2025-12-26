package observability

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

type TracingProvider struct {
	tp       *sdktrace.TracerProvider
	shutdown func(context.Context) error
}

func NewTracingProvider(serviceName ServiceName, version ServiceVersion, endpoint string, sampleRate float64, insecureConn bool, logger log.Logger) (*TracingProvider, error) {
	logHelper := log.NewHelper(logger)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Configure OTLP gRPC connection options
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(endpoint),
	}

	if insecureConn {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}

	// Create OTLP exporter
	exporter, err := otlptracegrpc.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	// Create resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(string(serviceName)),
			semconv.ServiceVersion(string(version)),
		),
	)
	if err != nil {
		logHelper.Warnf("failed to create resource: %v", err)
		res = resource.Default()
	}

	// Create tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(sampleRate)),
	)

	// Set global tracer provider and propagator
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	logHelper.Infof("Tracing initialized: endpoint=%s, sample_rate=%.2f", endpoint, sampleRate)

	return &TracingProvider{
		tp: tp,
		shutdown: func(ctx context.Context) error {
			return tp.Shutdown(ctx)
		},
	}, nil
}

func (t *TracingProvider) Shutdown(ctx context.Context) error {
	if t.shutdown != nil {
		return t.shutdown(ctx)
	}
	return nil
}
