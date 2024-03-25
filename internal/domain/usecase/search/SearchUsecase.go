package search

import (
	"bytes"
	"encoding/json"
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

// GrandFilter godoc
//
//	@Summary		Perform grand filter search
//	@Description	Perform a grand filter search based on provided parameters
//	@Tags			search
//	@Produce		json
//	@Param			filterStrings		body	[]dto.SearchString		true	"Array of search strings with && as the joining operator"
//	@Param			filterPeople		body	[]dto.SearchPeople		true	"Array of search people"
//	@Param			filterPublishers	body	[]dto.SearchPublishers	true	"Array of search publishers"
//	@Param			filterCountry		body	[]dto.SearchCountry		true	"Array of search countries"
//	@Param			filterCategories	body	[]dto.SearchCategory	true	"Array of search categories"
//	@Param			filterLanguages		body	[]dto.SearchLanguage	true	"Array of search languages"
//	@Param			filterTime			body	dto.SearchTime			true	"Time filter"
//	@Success		200
//	@Router			/grandFilter [post]
func (u *usecase) GrandFilter(c *gin.Context) {
	ctx := c.Request.Context()

	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(c.Request.Body)
	if err != nil {
		logger.Error("Ошибка поиска", zap.Error(err))
		_ = c.Error(err)
		lib.ResponseBadRequest(c, err, "У запроса нет тела")
		return
	}

	req := dto.GrandFilterRequest{}
	if err := json.Unmarshal(buf.Bytes(), &req); err != nil {
		logger.Error("Ошибка поиска", zap.Error(err))
		_ = c.Error(err)
		lib.ResponseBadRequest(c, err, "Неправильное тело запроса")
		return
	}
	//if err := c.ShouldBind(&req); err != nil {
	//	logger.Error("Ошибка поиска", zap.Error(err))
	//	_ = c.Error(err)
	//	lib.ResponseBadRequest(c, err, "Неправильное тело запроса")
	//	return
	//}

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

// SearchDefaultPublisherWithHints  godoc
//
//	@Summary		Search publishers with hints using default search
//	@Description	Search publishers with hints using default search
//	@Tags			search
//	@Produce		json
//	@Success		200
//	@Router			/searchPublisherWithHints/ [post]
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

// SearchPublisherWithHints godoc
//
//	@Summary		Search publishers with hints
//	@Description	Search publishers with hints based on provided search string
//	@Tags			search
//	@Produce		json
//	@Param			search	path	string	true	"Search string"
//	@Success		200
//	@Router			/searchPublisherWithHints/{search} [post]
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

// SearchLanguages godoc
//
//	@Summary		Search languages
//	@Description	Search languages
//	@Tags			search
//	@Produce		json
//	@Success		200	{array}	string
//	@Router			/searchLanguages [post]
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

// SearchCategoriesWithHints godoc
//
//	@Summary		Search categories with hints
//	@Description	Search categories with hints
//	@Tags			search
//	@Produce		json
//	@Success		200
//	@Router			/searchCategoryWithHints [post]
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

// SearchPeopleWithHints godoc
//
//	@Summary		Search people with hints
//	@Description	Search people with hints
//	@Tags			search
//	@Produce		json
//	@Success		200
//	@Router			/searchPeopleWithHints [post]
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
