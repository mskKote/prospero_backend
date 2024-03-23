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

// ---------------------------------------------------- publishers CRUD

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

func (u *usecase) Harvest(c *gin.Context) {
	if err := u.articles.ParseAllOnce(c, true); err != nil {
		_ = c.Error(err)
		lib.ResponseBadRequest(c, err, "Не получилось прочитать источники")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}
