package search

import (
	"github.com/gin-gonic/gin"
	"github.com/mskKote/prospero_backend/internal/controller/http/v1/dto"
	"github.com/mskKote/prospero_backend/internal/domain/service/articleService"
	"github.com/mskKote/prospero_backend/internal/domain/service/publishersService"
	"github.com/mskKote/prospero_backend/pkg/lib"
	"github.com/mskKote/prospero_backend/pkg/logging"
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

func (u *usecase) GrandFilter(c *gin.Context) {
	ctx := c.Request.Context()

	req := dto.GrandFilterRequest{}
	if err := c.ShouldBind(&req); err != nil {
		logger.Error("Ошибка поиска", zap.Error(err))
		_ = c.Error(err)
		lib.ResponseBadRequest(c, err, "Неправильное тело запроса")
		return
	}

	sizeQuery := c.Query("size")
	size, err := strconv.Atoi(sizeQuery)
	if err != nil {
		size = 150
	}

	grandFilter, total, err := u.articles.FindWithGrandFilter(ctx, req, size)
	if err != nil {
		_ = c.Error(err)
		lib.ResponseBadRequest(c, err, "Не смогли найти статьи")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    grandFilter,
		"total":   total,
		"message": "ok",
	})
}

func (u *usecase) SearchDefaultPublisherWithHints(c *gin.Context) {
	// Строка поиска
	publishers, err := u.publishers.FindPublishersByNameViaES(c, "")
	if err != nil {
		_ = c.Error(err)
		lib.ResponseBadRequest(c, err, "Не получилось найти публицистов по умолчанию")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
		"data":    publishers,
	})
}

func (u *usecase) SearchPublisherWithHints(c *gin.Context) {
	// Строка поиска
	search := c.Param("search")

	publishers, err := u.publishers.FindPublishersByNameViaES(c, search)
	if err != nil {
		_ = c.Error(err)
		lib.ResponseBadRequest(c, err, "Не получилось найти публицистов по "+search)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
		"data":    publishers,
	})
}

func (u *usecase) SearchLanguages(c *gin.Context) {
	ctx := c.Request.Context()

	languages, err := u.articles.FindAllLanguages(ctx)
	if err != nil {
		_ = c.Error(err)
		lib.ResponseBadRequest(c, err, "Не смогли найти языки")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    languages,
		"message": "ok",
	})
}

func (u *usecase) SearchCategoriesWithHints(c *gin.Context) {
	ctx := c.Request.Context()

	// Параметр поиска
	req := c.Query("q")

	categories, err := u.articles.FindCategory(ctx, req)
	if err != nil {
		_ = c.Error(err)
		lib.ResponseBadRequest(c, err, "Не смогли найти категории")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    categories,
		"message": "ok",
	})
}

func (u *usecase) SearchPeopleWithHints(c *gin.Context) {
	ctx := c.Request.Context()

	// Параметр поиска
	req := c.Query("q")
	people, err := u.articles.FindPeople(ctx, req)
	if err != nil {
		_ = c.Error(err)
		lib.ResponseBadRequest(c, err, "Не смогли найти людей")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    people,
		"message": "ok",
	})
}
