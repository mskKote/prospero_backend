package search

import (
	"github.com/gin-gonic/gin"
	"github.com/mskKote/prospero_backend/internal/controller/http/v1/dto"
	"github.com/mskKote/prospero_backend/internal/controller/http/v1/routes"
	"github.com/mskKote/prospero_backend/internal/domain/service/articleService"
	"github.com/mskKote/prospero_backend/internal/domain/service/publishersService"
	"github.com/mskKote/prospero_backend/pkg/lib"
	"github.com/mskKote/prospero_backend/pkg/logging"
	"github.com/mskKote/prospero_backend/pkg/tracing"
	"go.uber.org/zap"
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
	ctx := c.Request.Context()

	req := dto.GrandFilterRequest{}
	if err := c.ShouldBind(&req); err != nil {
		logger.Error("Ошибка при поиске", zap.Error(err))
		lib.ResponseBadRequest(c, err, "Неправильное тело запроса")
		return
	}

	grandFilter, err := u.articles.FindWithGrandFilter(ctx, req)
	if err != nil {
		lib.ResponseBadRequest(c, err, "Не смогли найти")
		return
	} else {
		c.JSON(http.StatusOK, gin.H{
			"data":    grandFilter,
			"message": "ok",
		})
	}
}
