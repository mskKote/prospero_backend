package search

import (
	"github.com/gin-gonic/gin"
	"github.com/mskKote/prospero_backend/pkg/logging"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"net/http"
)

var logger = logging.GetLogger()

// TODO: сервисы к ElasticSearch

type Service interface {
	GrandFilter(g *gin.Context)
}

type Usecase struct {
	Service
}

func (h *Usecase) GrandFilter(c *gin.Context) {
	ctx := c.Request.Context()
	searchStr := c.Param("search")

	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("Строка поиска", searchStr))

	traceID := span.SpanContext().TraceID().String()
	logger.InfoContext(ctx, "GrandFilter trace "+traceID)
	c.Header("x-trace-id", traceID)
	c.JSON(http.StatusOK, gin.H{"message": searchStr + " -- ok"})
}
