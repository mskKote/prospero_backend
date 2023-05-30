package tracing

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mskKote/prospero_backend/pkg/config"
	"github.com/mskKote/prospero_backend/pkg/logging"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	traceSDK "go.opentelemetry.io/otel/sdk/trace"
	semConv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

const (
	ProsperoHeader   = "prospero-trace-id"
	ProsperoTraceKey = "prospero-trace-key"
)

var (
	cfg    = config.GetConfig()
	logger = logging.GetLogger()
)

func Startup(router *gin.Engine) *traceSDK.TracerProvider {

	url := fmt.Sprintf("http://%s:%s/api/traces",
		cfg.Tracing.Host,
		cfg.Tracing.Port)

	tp, err := tracerProvider(url)
	if err != nil {
		logger.Fatal("Ошибка на старте", zap.Error(err))
	}

	// Register our TracerProvider as the global so any imported
	// instrumentation in the future will default to using it.
	otel.SetTracerProvider(tp)
	router.Use(otelgin.Middleware(cfg.Service, otelgin.WithTracerProvider(tp)))
	return tp
}

// tracerProvider returns an OpenTelemetry TracerProvider configured to use
// the Jaeger exporter that will send spans to the provided url. The returned
// TracerProvider will also use a Resource configured with all the information
// about the application.
func tracerProvider(url string) (*traceSDK.TracerProvider, error) {
	// Create the Jaeger exporter
	exp, err := jaeger.New(
		jaeger.WithCollectorEndpoint(
			jaeger.WithEndpoint(url),
		),
	)

	if err != nil {
		return nil, err
	}

	tp := traceSDK.NewTracerProvider(
		// Always be sure to batch in production.
		traceSDK.WithBatcher(exp),
		// Record information about this application in a Resource.
		traceSDK.WithResource(resource.NewWithAttributes(
			semConv.SchemaURL,
			semConv.ServiceName(cfg.Service),
			attribute.String("environment", cfg.Environment),
			attribute.Int64("ID", 1),
		)),
	)
	return tp, nil
}

func LogRequestTrace(c *gin.Context) {
	ctx := c.Request.Context()
	span := trace.SpanFromContext(ctx)
	traceID := span.SpanContext().TraceID().String()

	logger.InfoContext(ctx, fmt.Sprintf("path=[%s] trace=[%s]", c.FullPath(), traceID))
	//span.SetAttributes(attribute.String("path", c.FullPath()))
}

func TraceHeader(c *gin.Context) {
	ctx := c.Request.Context()
	span := trace.SpanFromContext(ctx)
	traceID := span.SpanContext().TraceID().String()
	c.Header(ProsperoHeader, traceID)
}

func SpanLogErr(span trace.Span, err error) {
	span.SetStatus(codes.Error, "Словили ошибку в ходе выполнения программы")
	span.SetAttributes(
		attribute.String("Ошибка", fmt.Sprintf("%v", err)))
}

func GetTracer(c *gin.Context) trace.Tracer {
	if val, ok := c.Get("otel-go-contrib-tracer"); ok {
		return val.(trace.Tracer)
	}
	return nil
}
func TracerToContext(ctx context.Context, tracer trace.Tracer) context.Context {
	return context.WithValue(ctx, ProsperoTraceKey, tracer)
}

func TracerFromContext(ctx context.Context) trace.Tracer {
	t := ctx.Value(ProsperoTraceKey)
	return t.(trace.Tracer)
}
