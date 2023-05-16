package sources

import (
	"github.com/gin-gonic/gin"
	"github.com/mskKote/prospero_backend/internal/controller/http/v1/routes"
	"net/http"
)

//var logger = logging.GetLogger()

// Service - зависимые сервисы
type services interface {
	routes.ISourcesUsecase
}

// Usecase использование сервисов
type Usecase struct {
	services
}

func (h *Usecase) AddSourceRSS(c *gin.Context) {
	//ctx := c.Request.Context()

	//span := trace.SpanFromContext(ctx)
	//span.SetAttributes(
	//	attribute.String("Строка поиска", searchStr))

	//traceID := span.SpanContext().TraceID().String()
	//logger.InfoContext(ctx, "GrandFilter trace "+traceID)
	//c.Header("x-trace-id", traceID)
	c.JSON(http.StatusOK, gin.H{"message": " -- ok"})
}

func (h *Usecase) GetSourcesRSS(c *gin.Context) {
	//ctx := c.Request.Context()

	//span := trace.SpanFromContext(ctx)
	//span.SetAttributes(
	//	attribute.String("Строка поиска", searchStr))

	//traceID := span.SpanContext().TraceID().String()
	//logger.InfoContext(ctx, "GrandFilter trace "+traceID)
	//c.Header("x-trace-id", traceID)
	c.JSON(http.StatusOK, gin.H{"message": " -- ok"})
}

func (h *Usecase) DeleteSourcesRSS(c *gin.Context) {
	//ctx := c.Request.Context()

	//span := trace.SpanFromContext(ctx)
	//span.SetAttributes(
	//	attribute.String("Строка поиска", searchStr))

	//traceID := span.SpanContext().TraceID().String()
	//logger.InfoContext(ctx, "GrandFilter trace "+traceID)
	//c.Header("x-trace-id", traceID)
	c.JSON(http.StatusOK, gin.H{"message": " -- ok"})
}
