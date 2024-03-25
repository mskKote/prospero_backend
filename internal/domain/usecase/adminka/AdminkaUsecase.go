package adminka

import (
	"github.com/gin-gonic/gin"
	"github.com/mskKote/prospero_backend/internal/domain/entity/publisher"
	"github.com/mskKote/prospero_backend/internal/domain/entity/source"
	"github.com/mskKote/prospero_backend/internal/domain/service/articleService"
	"github.com/mskKote/prospero_backend/internal/domain/service/publishersService"
	"github.com/mskKote/prospero_backend/internal/domain/service/sourcesService"
	"github.com/mskKote/prospero_backend/pkg/lib"
	"github.com/mskKote/prospero_backend/pkg/logging"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

var (
	logger = logging.GetLogger()
)

const pageSize int = 6

// usecase - зависимые сервисы
type usecase struct {
	sources    sourcesService.ISourceService
	publishers publishersService.IPublishersService
	articles   articleService.IArticleService
}

func New(
	s *sourcesService.ISourceService,
	p *publishersService.IPublishersService,
	a *articleService.IArticleService) IAdminkaUseCase {
	return &usecase{*s, *p, *a}
}

// AddSourceAndPublisher godoc
//
//	@Summary		Add Source and Publisher
//	@Description	Add a new source along with its corresponding publisher
//	@Tags			sources
//	@Accept			json
//	@Produce		json
//	@Param			dto	body	source.AddSourceAndPublisherDTO	true	"Add Source and Publisher DTO"
//	@Success		200
//	@Router			/addSourceAndPublisher [post]
func (u *usecase) AddSourceAndPublisher(c *gin.Context) {
	sp := &source.AddSourceAndPublisherDTO{}
	if err := c.ShouldBind(&sp); err != nil {
		_ = c.Error(err)
		logger.Error("Ошибка при добавлении", zap.Error(err))
		lib.ResponseBadRequest(c, err, "Неправильное тело запроса")
		return
	}

	// ADD PUBLISHER
	p := &publisher.AddPublisherDTO{
		Name:      sp.Name,
		Country:   sp.Country,
		City:      sp.City,
		Longitude: sp.Longitude,
		Latitude:  sp.Latitude,
	}
	created, err := u.publishers.Create(c, p)
	if err != nil {
		_ = c.Error(err)
		lib.ResponseBadRequest(c, err, "Ошибка при добавлении источника")
		return
	}

	// ADD SOURCE
	s := &source.AddSourceDTO{
		RssURL:      sp.RssUrl,
		PublisherID: created.PublisherID,
	}
	if _, err := u.sources.AddSource(c, *s); err != nil {
		_ = c.Error(err)
		lib.ResponseBadRequest(c, err, "Не получилось добавить источник")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

// ---------------------------------------------------- sources CRUD

// CreateSourceRSS godoc
//
//	@Summary		Create new RSS source
//	@Description	Create a new RSS source
//	@Tags			sources
//	@Accept			json
//	@Produce		json
//	@Param			dto	body	source.AddSourceDTO	true	"Add Source DTO"
//	@Success		200
//	@Router			/RSS/addSource [post]
func (u *usecase) CreateSourceRSS(c *gin.Context) {
	s := source.AddSourceDTO{}
	if err := c.Bind(&s); err != nil {
		_ = c.Error(err)
		lib.ResponseBadRequest(c, err, "Неправильное тело запроса")
		return
	}

	src, err := u.sources.AddSource(c, s)
	if err != nil {
		_ = c.Error(err)
		lib.ResponseBadRequest(c, err, "Не получилось добавить источник")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
		"data":    src,
	})
}

// ReadSourcesRSS godoc
//
//	@Summary		Read RSS sources
//	@Description	Read RSS sources with optional search and pagination
//	@Tags			sources
//	@Produce		json
//	@Param			page	query	int		false	"Page number"
//	@Param			search	query	string	false	"Search query"
//	@Success		200
//	@Router			/RSS/getSources [get]
func (u *usecase) ReadSourcesRSS(c *gin.Context) {
	var src []*source.DTO
	var err error

	// Total
	var total int64
	if count, err := u.sources.Count(c); err != nil {
		_ = c.Error(err)
		lib.ResponseBadRequest(c, err, "Не получилось посчитать количество элементов "+c.Query("search"))
		return
	} else {
		total = count
	}

	pageQuery := c.DefaultQuery("page", "0")
	page, err := strconv.Atoi(pageQuery)
	if err != nil {
		_ = c.Error(err)
		lib.ResponseBadRequest(c, err, "Неправильные параметры запроса")
		return
	}

	if int(total) < pageSize*page {
		_ = c.Error(err)
		lib.ResponseBadRequest(c, err, "Страницы "+c.Query("page")+" нет")
		return
	}

	if search := c.Query("search"); search != "" {
		src, err = u.sources.FindByPublisherName(c, search, page, pageSize)
	} else {
		src, err = u.sources.FindAll(c, page, pageSize)
	}

	if err != nil {
		_ = c.Error(err)
		lib.ResponseBadRequest(c, err, "Не получилось найти RSS источники "+c.Query("search"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
		"data":    src,
	})
}

// ReadSourcesRSSWithPublishers godoc
//
//	@Summary		Read RSS sources with publishers
//	@Description	Read RSS sources with associated publishers
//	@Tags			sources
//	@Produce		json
//	@Param			page	query	int		false	"Page number"
//	@Param			search	query	string	false	"Search query"
//	@Success		200
//	@Router			/RSS/getEnrichedSources [get]
func (u *usecase) ReadSourcesRSSWithPublishers(c *gin.Context) {
	var src []*source.DTO
	var err error

	// Total
	var total int64
	if count, err := u.sources.Count(c); err != nil {
		_ = c.Error(err)
		lib.ResponseBadRequest(c, err, "Не получилось посчитать количество элементов "+c.Query("search"))
		return
	} else {
		total = count
	}

	pageQuery := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageQuery)
	if err != nil {
		_ = c.Error(err)
		lib.ResponseBadRequest(c, err, "Неправильные параметры запроса")
		return
	}

	//if int(total) < pageSize*page {
	//	c.JSON(http.StatusNotFound, gin.H{
	//		"message": "Страницы " + c.Query("page") + " нет"})
	//	return
	//}

	// Поиск
	if search := c.Query("search"); search != "" {
		src, err = u.sources.FindByPublisherName(c, search, page-1, pageSize)
	} else {
		src, err = u.sources.FindAll(c, page-1, pageSize)
	}

	if err != nil {
		_ = c.Error(err)
		lib.ResponseBadRequest(c, err, "Не получилось найти RSS источники "+c.Query("search"))
		return
	}

	// Enrich
	pubMap := map[string]*publisher.DTO{}

	var srcIDs []string
	for _, s := range src {
		srcIDs = append(srcIDs, s.PublisherID)
	}

	if publishers, err := u.publishers.FindPublishersByIDs(c, srcIDs); err != nil {
		_ = c.Error(err)
		lib.ResponseBadRequest(c, err, "Не нашли publishers")
		return
	} else {
		for _, p := range publishers {
			pubMap[p.PublisherID] = p
		}
	}

	// Join 2 сущностей по publisherID
	var srcEnriched []*source.WithPublisher
	for _, s := range src {
		if p, ok := pubMap[s.PublisherID]; ok {
			srcEnriched = append(srcEnriched, &source.WithPublisher{
				Source:    s,
				Publisher: p,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
		"data":    srcEnriched,
		"pagination": gin.H{
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// UpdateSourceRSS godoc
//
//	@Summary		Update RSS source
//	@Description	Update existing RSS source
//	@Tags			sources
//	@Accept			json
//	@Produce		json
//	@Param			dto	body	source.DTO	true	"Update Source DTO"
//	@Success		200
//	@Router			/RSS/updateSource [put]
func (u *usecase) UpdateSourceRSS(c *gin.Context) {
	dto := &source.DTO{}
	if err := c.Bind(&dto); err != nil {
		_ = c.Error(err)
		lib.ResponseBadRequest(c, err, "Неправильное тело запроса")
		return
	}

	data, err := u.sources.Update(c, dto)
	if err != nil {
		_ = c.Error(err)
		lib.ResponseBadRequest(c, err, "Не получилось обновить RSS источник")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
		"data":    data,
	})
}

// DeleteSourceRSS godoc
//
//	@Summary		Delete RSS source
//	@Description	Delete RSS source by ID
//	@Tags			sources
//	@Accept			json
//	@Produce		json
//	@Param			dto	body	source.DeleteSourceDTO	true	"Delete Source DTO"
//	@Success		200
//	@Router			/RSS/removeSource [delete]
func (u *usecase) DeleteSourceRSS(c *gin.Context) {
	dto := source.DeleteSourceDTO{}
	if err := c.Bind(&dto); err != nil {
		_ = c.Error(err)
		lib.ResponseBadRequest(c, err, "Неправильное тело запроса")
		return
	}

	if err := u.sources.Delete(c, dto); err != nil {
		_ = c.Error(err)
		lib.ResponseBadRequest(c, err, "Не получилось удалить RSS источник")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

// ---------------------------------------------------- publishers CRUD

// CreatePublisher godoc
//
//	@Summary		Create new publisher
//	@Description	Create a new publisher
//	@Tags			publishers
//	@Accept			json
//	@Produce		json
//	@Param			dto	body	publisher.AddPublisherDTO	true	"Add Publisher DTO"
//	@Success		200
//	@Router			/addPublisher [post]
func (u *usecase) CreatePublisher(c *gin.Context) {
	s := &publisher.AddPublisherDTO{}
	if err := c.Bind(&s); err != nil {
		_ = c.Error(err)
		logger.Error("Ошибка при добавлении источника", zap.Error(err))
		lib.ResponseBadRequest(c, err, "Неправильное тело запроса")
		return
	}

	data, err := u.publishers.Create(c, s)
	if err != nil {
		_ = c.Error(err)
		lib.ResponseBadRequest(c, err, "Ошибка при добавлении источника")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
		"data":    data,
	})
}

// ReadPublishers godoc
//
//	@Summary		Read publishers
//	@Description	Read publishers with optional search
//	@Tags			publishers
//	@Produce		json
//	@Param			search	query	string	false	"Search query"
//	@Success		200
//	@Router			/getPublishers [get]
func (u *usecase) ReadPublishers(c *gin.Context) {
	var publishers []*publisher.DTO
	var err error

	if search := c.Query("search"); search != "" {
		publishers, err = u.publishers.FindPublishersByName(c, search)
	} else {
		publishers, err = u.publishers.FindAll(c)
	}

	if err != nil {
		_ = c.Error(err)
		lib.ResponseBadRequest(c, err, "Не получилось найти RSS источники "+c.Query("search"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
		"data":    publishers,
	})
}

// UpdatePublisher godoc
//
//	@Summary		Update publisher
//	@Description	Update publisher information
//	@Tags			publishers
//	@Accept			json
//	@Produce		json
//	@Param			dto	body	publisher.DTO	true	"Publisher DTO"
//	@Success		200
//	@Router			/updatePublisher [put]
func (u *usecase) UpdatePublisher(c *gin.Context) {
	dto := &publisher.DTO{}
	if err := c.Bind(&dto); err != nil {
		_ = c.Error(err)
		lib.ResponseBadRequest(c, err, "Неправильное тело запроса")
		return
	}

	if err := u.publishers.Update(c, dto); err != nil {
		_ = c.Error(err)
		lib.ResponseBadRequest(c, err, "Не получилось обновить")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

// DeletePublisher godoc
//
//	@Summary		Delete publisher
//	@Description	Delete publisher by ID
//	@Tags			publishers
//	@Accept			json
//	@Produce		json
//	@Param			dto	body	publisher.DeletePublisherDTO	true	"Delete Publisher DTO"
//	@Success		200
//	@Router			/removePublisher [delete]
func (u *usecase) DeletePublisher(c *gin.Context) {
	dto := &publisher.DeletePublisherDTO{}
	if err := c.Bind(&dto); err != nil {
		_ = c.Error(err)
		lib.ResponseBadRequest(c, err, "Неправильное тело запроса")
		return
	}

	if err := u.publishers.Delete(c, dto); err != nil {
		_ = c.Error(err)
		lib.ResponseBadRequest(c, err, "Не получилось удалить")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

// ---------------------------------------------------- RSS

// Harvest godoc
//
//	@Summary		Harvest RSS
//	@Description	Harvest RSS feeds and parse articles
//	@Tags			sources
//	@Produce		json
//	@Success		200
//	@Router			/RSS/harvest [post]
func (u *usecase) Harvest(c *gin.Context) {
	if err := u.articles.ParseAllOnce(c, true); err != nil {
		_ = c.Error(err)
		lib.ResponseBadRequest(c, err, "Не получилось прочитать источники")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}
