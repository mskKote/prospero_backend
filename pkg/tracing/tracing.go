package tracing

import (
	"bytes"
	"context"
	"encoding/json"
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
	"io"
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
	router.Use(CustomHeaderMiddleware())
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
}

func TraceHeader(c *gin.Context) {
	ctx := c.Request.Context()
	span := trace.SpanFromContext(ctx)
	traceID := span.SpanContext().TraceID().String()
	c.Header(ProsperoHeader, traceID)
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

//------------------------------------------------------------ Custom Middleware for Jaeger

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w bodyLogWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func CustomHeaderMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Trace
		TraceHeader(c)
		LogRequestTrace(c)

		// Tracer for spans
		tracer := GetTracer(c)
		ctx := c.Request.Context()
		ctx = TracerToContext(ctx, tracer)
		c.Request = c.Request.WithContext(ctx)
		span := trace.SpanFromContext(ctx)

		// Req body&query to span
		var buf bytes.Buffer
		tee := io.TeeReader(c.Request.Body, &buf)
		body, _ := io.ReadAll(tee)
		c.Request.Body = io.NopCloser(&buf)

		span.SetAttributes(attribute.String("1. Тело запроса", string(body)))
		queryParams, _ := json.Marshal(c.Request.URL.Query())
		span.SetAttributes(attribute.String("2. Параметры query", string(queryParams)))

		// Log response
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		c.Next()

		// Error
		if len(c.Errors) > 0 {
			span.SetStatus(codes.Error, "Ошибка в ходе выполнения программы")
			for i, err := range c.Errors {
				span.SetAttributes(
					attribute.String(
						fmt.Sprintf("Ошибка №%d", i),
						fmt.Sprintf("%v", err)))
			}
		}

		// Response
		span.SetAttributes(attribute.String("Ответ", blw.body.String()))
	}
}
