package search

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mskKote/prospero_backend/internal/controller/http/v1/dto"
	"github.com/mskKote/prospero_backend/internal/domain/service/articleService"
	"github.com/mskKote/prospero_backend/internal/domain/service/publishersService"
	"github.com/mskKote/prospero_backend/pkg/lib"
	"github.com/mskKote/prospero_backend/pkg/logging"
	"github.com/mskKote/prospero_backend/pkg/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"net/http"
	"strconv"
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

	// Строка поиска
	if publishers, err := u.publishers.FindPublishersByNameViaES(c, ""); err != nil {
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

	// Строка поиска
	search := c.Param("search")

	if publishers, err := u.publishers.FindPublishersByNameViaES(c, search); err != nil {
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
	reqJSON, _ := json.Marshal(req)
	span.SetAttributes(attribute.String("1. Тело запроса", string(reqJSON)))

	sizeQuery := c.Query("size")
	span.SetAttributes(attribute.String("2. Параметр size", sizeQuery))
	size, err := strconv.Atoi(sizeQuery)
	if err != nil {
		size = 150
	}

	if grandFilter, total, err := u.articles.FindWithGrandFilter(ctx, req, size); err != nil {
		tracing.SpanLogErr(span, err)
		lib.ResponseBadRequest(c, err, "Не смогли найти статьи")
	} else {
		var respSpan []string
		for _, dbo := range grandFilter {
			respSpan = append(respSpan,
				fmt.Sprintf("[%s] %s", dbo.Publisher.Name, dbo.Name))
		}
		span.SetAttributes(attribute.StringSlice("Полученные статьи", respSpan))
		c.JSON(http.StatusOK, gin.H{
			"data":    grandFilter,
			"total":   total,
			"message": "ok",
		})
	}
}

func (u *usecase) SearchLanguages(c *gin.Context) {
	tracing.TraceHeader(c)
	tracing.LogRequestTrace(c)
	tracer := tracing.GetTracer(c)
	ctx := c.Request.Context()
	ctx = tracing.TracerToContext(ctx, tracer)
	span := trace.SpanFromContext(ctx)

	if languages, err := u.articles.FindAllLanguages(ctx); err != nil {
		tracing.SpanLogErr(span, err)
		lib.ResponseBadRequest(c, err, "Не смогли найти языки")
	} else {
		var respSpan []string
		for _, dbo := range languages {
			respSpan = append(respSpan, dbo.Name)
		}
		span.SetAttributes(attribute.StringSlice("Полученные языки", respSpan))
		c.JSON(http.StatusOK, gin.H{
			"data":    languages,
			"message": "ok",
		})
	}
}

func (u *usecase) SearchCategoriesWithHints(c *gin.Context) {
	tracing.TraceHeader(c)
	tracing.LogRequestTrace(c)
	tracer := tracing.GetTracer(c)
	ctx := c.Request.Context()
	ctx = tracing.TracerToContext(ctx, tracer)
	span := trace.SpanFromContext(ctx)

	// Параметр поиска
	req := c.Query("q")
	span.SetAttributes(attribute.String("1. Поиск", req))

	if categories, err := u.articles.FindCategory(ctx, req); err != nil {
		tracing.SpanLogErr(span, err)
		lib.ResponseBadRequest(c, err, "Не смогли найти категории")
	} else {
		var respSpan []string
		for _, dbo := range categories {
			respSpan = append(respSpan, dbo.Name)
		}
		span.SetAttributes(attribute.StringSlice("Полученные категории", respSpan))
		c.JSON(http.StatusOK, gin.H{
			"data":    categories,
			"message": "ok",
		})
	}
}

func (u *usecase) SearchPeopleWithHints(c *gin.Context) {
	tracing.TraceHeader(c)
	tracing.LogRequestTrace(c)
	tracer := tracing.GetTracer(c)
	ctx := c.Request.Context()
	ctx = tracing.TracerToContext(ctx, tracer)
	span := trace.SpanFromContext(ctx)

	// Параметр поиска
	req := c.Query("q")
	span.SetAttributes(attribute.String("1. Поиск", req))

	if people, err := u.articles.FindPeople(ctx, req); err != nil {
		tracing.SpanLogErr(span, err)
		lib.ResponseBadRequest(c, err, "Не смогли найти людей")
	} else {
		var respSpan []string
		for _, dbo := range people {
			respSpan = append(respSpan, dbo.FullName)
		}
		span.SetAttributes(attribute.StringSlice("Полученные люди", respSpan))
		c.JSON(http.StatusOK, gin.H{
			"data":    people,
			"message": "ok",
		})
	}
}
