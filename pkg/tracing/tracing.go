package tracing

import (
	"github.com/gin-gonic/gin"
	"github.com/mskKote/prospero_backend/pkg/config"
	"github.com/mskKote/prospero_backend/pkg/logging"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	traceSDK "go.opentelemetry.io/otel/sdk/trace"
	semConv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.uber.org/zap"
)

var (
	cfg    = config.GetConfig()
	logger = logging.GetLogger()
)

func Startup(router *gin.Engine) *traceSDK.TracerProvider {
	tp, err := tracerProvider("http://jaeger:14268/api/traces")
	if err != nil {
		logger.Fatal("Ошибка со стартом", zap.Error(err))
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
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		return nil, err
	}

	environment := "production"
	if cfg.IsDebug {
		environment = "development"
	}

	tp := traceSDK.NewTracerProvider(
		// Always be sure to batch in production.
		traceSDK.WithBatcher(exp),
		// Record information about this application in a Resource.
		traceSDK.WithResource(resource.NewWithAttributes(
			semConv.SchemaURL,
			semConv.ServiceName(cfg.Service),
			attribute.String("environment", environment),
			attribute.Int64("ID", 1),
		)),
	)
	return tp, nil
}
