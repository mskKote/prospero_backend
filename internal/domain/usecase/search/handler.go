package search

import (
	"github.com/gin-gonic/gin"
	"github.com/mskKote/prospero_backend/internal/controller/http/v1/routes"
	"github.com/mskKote/prospero_backend/internal/domain/service/publishersService"
	"github.com/mskKote/prospero_backend/pkg/lib"
	"github.com/mskKote/prospero_backend/pkg/logging"
	"github.com/mskKote/prospero_backend/pkg/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"net/http"
)

var logger = logging.GetLogger()

// ISearchUsecase - зависимые сервисы
type ISearchUsecase interface {
	routes.ISearchUsecase
}

// usecase - зависимые сервисы
type usecase struct {
	publishers publishersService.IPublishersService
}

func New(p *publishersService.IPublishersService) ISearchUsecase {
	return &usecase{*p}
}

func (u *usecase) SearchPublisherWithHints(c *gin.Context) {
	tracing.TraceHeader(c)
	tracing.LogRequestTrace(c)
	ctx := c.Request.Context()

	// Строка поиска
	search := c.Param("search")

	if publishers, err := u.publishers.FindPublishersByNameViaES(ctx, search); err != nil {
		lib.ResponseBadRequest(c, err, "Не получилось найти публицистов по "+search)
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message": "ok",
			"data":    publishers,
		})
	}
}

func (u *usecase) GrandFilter(c *gin.Context) {
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
