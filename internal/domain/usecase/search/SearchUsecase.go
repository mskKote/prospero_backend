package search

import (
	"github.com/gin-gonic/gin"
	"github.com/mskKote/prospero_backend/internal/controller/http/v1/dto"
	"github.com/mskKote/prospero_backend/internal/domain/service/articleService"
	"github.com/mskKote/prospero_backend/internal/domain/service/publishersService"
	"github.com/mskKote/prospero_backend/pkg/lib"
	"github.com/mskKote/prospero_backend/pkg/logging"
	"github.com/mskKote/prospero_backend/pkg/tracing"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"net/http"
)

var logger = logging.GetLogger()

// usecase - зависимые сервисы
type usecase struct {
	publishers publishersService.IPublishersService
	articles   articleService.IArticleService
}

func New(p *publishersService.IPublishersService, a *articleService.IArticleService) ISearchUsecase {
	return &usecase{*p, *a}
}

func (u *usecase) SearchDefaultPublisherWithHints(c *gin.Context) {
	tracing.TraceHeader(c)
	tracing.LogRequestTrace(c)
	ctx := c.Request.Context()

	// Строка поиска
	if publishers, err := u.publishers.FindPublishersByNameViaES(ctx, ""); err != nil {
		lib.ResponseBadRequest(c, err, "Не получилось найти публицистов по умолчанию")
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message": "ok",
			"data":    publishers,
		})
	}
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
	tracing.TraceHeader(c)
	tracing.LogRequestTrace(c)
	tracer := tracing.GetTracer(c)
	ctx := c.Request.Context()
	ctx = tracing.TracerToContext(ctx, tracer)
	span := trace.SpanFromContext(ctx)

	req := dto.GrandFilterRequest{}
	if err := c.ShouldBind(&req); err != nil {
		logger.Error("Ошибка поиска", zap.Error(err))
		span := trace.SpanFromContext(ctx)
		tracing.SpanLogErr(span, err)
		lib.ResponseBadRequest(c, err, "Неправильное тело запроса")
		return
	}

	if grandFilter, err := u.articles.FindWithGrandFilter(ctx, req); err != nil {
		tracing.SpanLogErr(span, err)
		lib.ResponseBadRequest(c, err, "Не смогли найти")
	} else {
		c.JSON(http.StatusOK, gin.H{
			"data":    grandFilter,
			"message": "ok",
		})
	}
}
